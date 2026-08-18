package tron

import (
	"context"
	"log"
	"math/big"
	"time"

	"exchange/internal/blockchain"
	bscCrypto "exchange/internal/blockchain/crypto"
)

// trc20TxListResp is a partial shape of the TronGrid TRC-20 tx list response.
type trc20TxListResp struct {
	Data []struct {
		TransactionID string `json:"transaction_id"`
		From          string `json:"from"`
		To            string `json:"to"`
		Value         string `json:"value"`          // token amount in base units
		TokenInfo     struct {
			Decimals int    `json:"decimals"`
			Symbol   string `json:"symbol"`
		} `json:"token_info"`
		BlockTimestamp int64 `json:"block_timestamp"` // ms since epoch
	} `json:"data"`
	Meta struct {
		At          int64  `json:"at"`
		Fingerprint string `json:"fingerprint"` // pagination cursor
	} `json:"meta"`
}

// TronMonitor polls TronGrid for inbound TRC-20 transfers to watched addresses.
type TronMonitor struct {
	client       *Client
	addresses    map[string]bool // set of deposit addresses to watch
	contracts    []string        // TRC-20 contract addresses to watch
	onDeposit    OnTronDepositFunc
	pollInterval time.Duration
}

// OnTronDepositFunc is called when an inbound transfer is detected.
type OnTronDepositFunc func(tx blockchain.IncomingTx)

// NewTronMonitor creates a TRON deposit monitor.
func NewTronMonitor(
	client *Client,
	contracts []string,
	onDeposit OnTronDepositFunc,
	pollInterval time.Duration,
) *TronMonitor {
	return &TronMonitor{
		client:       client,
		addresses:    make(map[string]bool),
		contracts:    contracts,
		onDeposit:    onDeposit,
		pollInterval: pollInterval,
	}
}

// AddAddress registers a TRON address for deposit monitoring.
func (m *TronMonitor) AddAddress(addr string) {
	m.addresses[addr] = true
}

// RemoveAddress deregisters a TRON address.
func (m *TronMonitor) RemoveAddress(addr string) {
	delete(m.addresses, addr)
}

// Start begins the polling loop. It blocks until ctx is cancelled.
func (m *TronMonitor) Start(ctx context.Context) {
	log.Printf("tron monitor: starting (poll every %s)", m.pollInterval)
	ticker := time.NewTicker(m.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("tron monitor: stopping")
			return
		case <-ticker.C:
			m.poll(ctx)
		}
	}
}

// poll fetches the latest TRC-20 transactions for each watched address.
func (m *TronMonitor) poll(ctx context.Context) {
	for addr := range m.addresses {
		for _, contract := range m.contracts {
			path := "/v1/accounts/" + addr + "/transactions/trc20" +
				"?contract_address=" + contract +
				"&only_to=true&limit=20"

			var resp trc20TxListResp
			if err := m.client.get(ctx, path, &resp); err != nil {
				log.Printf("tron monitor: polling %s: %v", addr, err)
				continue
			}

			for _, tx := range resp.Data {
				if tx.To != addr {
					continue
				}

				amount := new(big.Int)
				amount.SetString(tx.Value, 10)

				m.onDeposit(blockchain.IncomingTx{
					TxHash:      tx.TransactionID,
					FromAddress: tx.From,
					ToAddress:   tx.To,
					Amount:      amount,
					Decimals:    tx.TokenInfo.Decimals,
					Asset:       tx.TokenInfo.Symbol,
					Network:     "TRON",
					Timestamp:   time.UnixMilli(tx.BlockTimestamp),
				})

				_ = bscCrypto.FromBaseUnits // keep import
			}
		}
	}
}

// GetIncomingTransactions implements blockchain.BlockchainAdapter for one-off queries.
func (c *Client) GetIncomingTransactions(ctx context.Context, address string, fromBlock int64) ([]blockchain.IncomingTx, error) {
	// TRON does not use block numbers for pagination in the TronGrid API —
	// it uses a fingerprint cursor. For the monitor use TronMonitor.Start().
	return nil, nil
}
