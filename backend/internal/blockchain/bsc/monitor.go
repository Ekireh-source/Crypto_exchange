package bsc

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"exchange/internal/blockchain"
)

// OnDepositFunc is called by the monitor when a confirmed inbound transfer
// is detected.  The service layer wires this to ProcessDeposit.
type OnDepositFunc func(tx blockchain.IncomingTx)

// Monitor watches a set of deposit addresses for inbound BNB and BEP-20
// transfers. It polls via eth_getLogs every pollInterval.
type Monitor struct {
	client          *Client
	mu              sync.RWMutex
	addresses       map[common.Address]bool // set of deposit addresses to watch
	tokenContracts  []common.Address        // BEP-20 contracts to watch
	onDeposit       OnDepositFunc
	pollInterval    time.Duration
	lastBlock       int64
}

// NewMonitor creates a deposit monitor.
// tokenContracts is the list of BEP-20 contract addresses to listen on.
func NewMonitor(
	client *Client,
	tokenContracts []string,
	onDeposit OnDepositFunc,
	pollInterval time.Duration,
) *Monitor {
	contracts := make([]common.Address, len(tokenContracts))
	for i, c := range tokenContracts {
		contracts[i] = common.HexToAddress(c)
	}
	return &Monitor{
		client:         client,
		addresses:      make(map[common.Address]bool),
		tokenContracts: contracts,
		onDeposit:      onDeposit,
		pollInterval:   pollInterval,
	}
}

// AddAddress registers an address to watch for incoming deposits.
func (m *Monitor) AddAddress(addr string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.addresses[common.HexToAddress(addr)] = true
}

// RemoveAddress stops watching the given address.
func (m *Monitor) RemoveAddress(addr string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.addresses, common.HexToAddress(addr))
}

// Start begins the polling loop.  It blocks until ctx is cancelled.
func (m *Monitor) Start(ctx context.Context) {
	log.Printf("bsc monitor: starting (poll every %s)", m.pollInterval)

	ticker := time.NewTicker(m.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("bsc monitor: stopping")
			return
		case <-ticker.C:
			if err := m.poll(ctx); err != nil {
				log.Printf("bsc monitor: poll error: %v", err)
			}
		}
	}
}

// poll fetches Transfer events since lastBlock and calls onDeposit for matches.
func (m *Monitor) poll(ctx context.Context) error {
	m.mu.RLock()
	count := len(m.addresses)
	m.mu.RUnlock()
	if count == 0 {
		return nil
	}

	currentBlock, err := m.client.rpc.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("fetching block number: %w", err)
	}

	if m.lastBlock == 0 {
		// First poll — start from 50 blocks back to catch recent deposits.
		m.lastBlock = int64(currentBlock) - 50
	}

	// ── 1. Native Transfers (ETH/BNB) ─────────────────────────────────────────
	for i := m.lastBlock + 1; i <= int64(currentBlock); i++ {
		block, err := m.client.rpc.BlockByNumber(ctx, big.NewInt(i))
		if err != nil {
			log.Printf("bsc monitor: error fetching block %d: %v", i, err)
			continue
		}

		for _, tx := range block.Transactions() {
			if tx.To() == nil {
				continue // contract creation
			}
			m.mu.RLock()
			isWatched := m.addresses[*tx.To()]
			m.mu.RUnlock()

			if isWatched && tx.Value().Cmp(big.NewInt(0)) > 0 {
				sender, err := types.Sender(types.LatestSignerForChainID(m.client.chainID), tx)
				fromAddr := ""
				if err == nil {
					fromAddr = sender.Hex()
				}
				log.Printf("bsc monitor: detected native deposit of %s wei to %s", tx.Value().String(), tx.To().Hex())
				m.onDeposit(blockchain.IncomingTx{
					TxHash:      tx.Hash().Hex(),
					FromAddress: fromAddr,
					ToAddress:   tx.To().Hex(),
					Amount:      tx.Value(),
					Decimals:    18,
					Asset:       "NATIVE", // Special indicator for native coin
					Network:     m.client.Network(),
					BlockNumber: i,
					Timestamp:   time.Now(),
				})
			}
		}
	}

	// ── 2. BEP-20 token Transfer events ───────────────────────────────────────
	transferABI, _ := abi.JSON(strings.NewReader(`[{
		"anonymous":false,
		"name":"Transfer",
		"type":"event",
		"inputs":[
			{"indexed":true,"name":"from","type":"address"},
			{"indexed":true,"name":"to","type":"address"},
			{"name":"value","type":"uint256"}
		]
	}]`))
	transferSig := transferABI.Events["Transfer"].ID

	for _, contract := range m.tokenContracts {
		query := ethereum.FilterQuery{
			FromBlock: big.NewInt(m.lastBlock + 1),
			ToBlock:   big.NewInt(int64(currentBlock)),
			Addresses: []common.Address{contract},
			Topics:    [][]common.Hash{{transferSig}},
		}

		logs, err := m.client.rpc.FilterLogs(ctx, query)
		if err != nil {
			log.Printf("bsc monitor: FilterLogs error: %v", err)
			continue
		}

		for _, l := range logs {
			if len(l.Topics) < 3 {
				continue
			}
			to := common.HexToAddress(l.Topics[2].Hex())
			
			m.mu.RLock()
			isWatched := m.addresses[to]
			m.mu.RUnlock()

			if !isWatched {
				continue // not one of our deposit addresses
			}

			from := common.HexToAddress(l.Topics[1].Hex())
			amount := new(big.Int).SetBytes(l.Data)

			log.Printf("bsc monitor: detected token deposit of %s to %s", amount.String(), to.Hex())
			m.onDeposit(blockchain.IncomingTx{
				TxHash:      l.TxHash.Hex(),
				FromAddress: from.Hex(),
				ToAddress:   to.Hex(),
				Amount:      amount,
				Decimals:    18, // Should be dynamic based on asset, simplifying for MVP
				Asset:       contract.Hex(), // Pass contract address so ProcessDeposit can map to assetID
				Network:     m.client.Network(),
				BlockNumber: int64(l.BlockNumber),
				Timestamp:   time.Now(),
			})
		}
	}

	m.lastBlock = int64(currentBlock)
	return nil
}
