package bsc

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	"exchange/internal/blockchain"
)

// OnDepositFunc is called by the monitor when a confirmed inbound transfer
// is detected.  The service layer wires this to ProcessDeposit.
type OnDepositFunc func(tx blockchain.IncomingTx)

// Monitor watches a set of deposit addresses for inbound BNB and BEP-20
// transfers. It polls via eth_getLogs every pollInterval.
type Monitor struct {
	client          *Client
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
	m.addresses[common.HexToAddress(addr)] = true
}

// RemoveAddress stops watching the given address.
func (m *Monitor) RemoveAddress(addr string) {
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
	if len(m.addresses) == 0 {
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

	// ── BEP-20 token Transfer events ──────────────────────────────────────────
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
			if !m.addresses[to] {
				continue // not one of our deposit addresses
			}

			from := common.HexToAddress(l.Topics[1].Hex())
			amount := new(big.Int).SetBytes(l.Data)

			m.onDeposit(blockchain.IncomingTx{
				TxHash:      l.TxHash.Hex(),
				FromAddress: from.Hex(),
				ToAddress:   to.Hex(),
				Amount:      amount,
				Decimals:    18,
				Network:     m.client.Network(),
				BlockNumber: int64(l.BlockNumber),
				Timestamp:   time.Now(), // block timestamp would require extra RPC call
			})
		}
	}

	m.lastBlock = int64(currentBlock)
	return nil
}
