package tron

import (
	"context"
	"log"
	"math/big"
	"sync"
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

// nativeTxListResp is a partial shape of the TronGrid native TRX tx list response.
type nativeTxListResp struct {
	Data []struct {
		TxID       string `json:"txID"`
		BlockTimestamp int64 `json:"block_timestamp"`
		RawData    struct {
			Contract []struct {
				Type      string `json:"type"` // "TransferContract"
				Parameter struct {
					Value struct {
						Amount   int64  `json:"amount"`     // in SUN
						ToAddress string `json:"to_address"` // hex format
						OwnerAddress string `json:"owner_address"`
					} `json:"value"`
				} `json:"parameter"`
			} `json:"contract"`
		} `json:"raw_data"`
	} `json:"data"`
	Meta struct {
		Fingerprint string `json:"fingerprint"`
	} `json:"meta"`
}

// TronMonitor polls TronGrid for inbound TRC-20 and native TRX transfers to watched addresses.
type TronMonitor struct {
	client       *Client
	mu           sync.RWMutex
	addresses    map[string]bool // set of deposit addresses to watch
	contracts    []string        // TRC-20 contract addresses to watch
	onDeposit    OnTronDepositFunc
	pollInterval time.Duration

	// seenTxIDs prevents processing the same transaction more than once.
	// TronGrid returns all recent transactions on every poll (no block cursor).
	seenMu  sync.Mutex
	seenIDs map[string]bool
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
		seenIDs:      make(map[string]bool),
	}
}

// AddAddress registers a TRON address for deposit monitoring.
func (m *TronMonitor) AddAddress(addr string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.addresses[addr] = true
}

// RemoveAddress deregisters a TRON address.
func (m *TronMonitor) RemoveAddress(addr string) {
	m.mu.Lock()
	defer m.mu.Unlock()
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

// markSeen returns true if this txID has NOT been seen before, and records it.
// Returns false if already seen (skip).
func (m *TronMonitor) markSeen(txID string) bool {
	m.seenMu.Lock()
	defer m.seenMu.Unlock()
	if m.seenIDs[txID] {
		return false
	}
	m.seenIDs[txID] = true
	return true
}

// poll fetches the latest transactions for each watched address.
func (m *TronMonitor) poll(ctx context.Context) {
	m.mu.RLock()
	addresses := make([]string, 0, len(m.addresses))
	for addr := range m.addresses {
		addresses = append(addresses, addr)
	}
	m.mu.RUnlock()

	if len(addresses) == 0 {
		return
	}

	for _, addr := range addresses {
		// ── 1. TRC-20 token deposits ─────────────────────────────────────────
		for _, contract := range m.contracts {
			path := "/v1/accounts/" + addr + "/transactions/trc20" +
				"?contract_address=" + contract +
				"&only_to=true&limit=20"

			var resp trc20TxListResp
			if err := m.client.get(ctx, path, &resp); err != nil {
				log.Printf("tron monitor: polling TRC-20 for %s: %v", addr, err)
				continue
			}

			for _, tx := range resp.Data {
				if tx.To != addr {
					continue
				}
				if !m.markSeen(tx.TransactionID) {
					continue // already processed
				}

				amount := new(big.Int)
				amount.SetString(tx.Value, 10)

				log.Printf("tron monitor: detected TRC-20 deposit of %s %s to %s (tx: %s)",
					tx.Value, tx.TokenInfo.Symbol, addr, tx.TransactionID)

				m.onDeposit(blockchain.IncomingTx{
					TxHash:      tx.TransactionID,
					FromAddress: tx.From,
					ToAddress:   tx.To,
					Amount:      amount,
					Decimals:    tx.TokenInfo.Decimals,
					Asset:       contract, // pass contract address for asset resolution
					Network:     "TRON",
					Timestamp:   time.UnixMilli(tx.BlockTimestamp),
				})
			}
		}

		// ── 2. Native TRX deposits ────────────────────────────────────────────
		// Uses the general transactions endpoint filtering for TransferContract type.
		nativePath := "/v1/accounts/" + addr + "/transactions" +
			"?only_to=true&limit=20"

		var nativeResp nativeTxListResp
		if err := m.client.get(ctx, nativePath, &nativeResp); err != nil {
			log.Printf("tron monitor: polling native TRX for %s: %v", addr, err)
			continue
		}

		for _, tx := range nativeResp.Data {
			if len(tx.RawData.Contract) == 0 {
				continue
			}
			contract := tx.RawData.Contract[0]
			if contract.Type != "TransferContract" {
				continue // not a plain TRX transfer
			}
			if !m.markSeen(tx.TxID) {
				continue // already processed
			}

			toHex := contract.Parameter.Value.ToAddress
			// TronGrid returns to_address in hex — convert to Base58 for comparison.
			toBase58, err := hexToTronAddress(toHex)
			if err != nil || toBase58 != addr {
				continue
			}

			sunAmt := contract.Parameter.Value.Amount
			amount := big.NewInt(sunAmt)

			log.Printf("tron monitor: detected native TRX deposit of %d SUN to %s (tx: %s)",
				sunAmt, addr, tx.TxID)

			m.onDeposit(blockchain.IncomingTx{
				TxHash:      tx.TxID,
				FromAddress: contract.Parameter.Value.OwnerAddress,
				ToAddress:   addr,
				Amount:      amount,
				Decimals:    6, // TRX has 6 decimals (SUN)
				Asset:       "NATIVE",
				Network:     "TRON",
				Timestamp:   time.UnixMilli(tx.BlockTimestamp),
			})
		}
	}
	_ = bscCrypto.FromBaseUnits // keep import
}

// GetIncomingTransactions implements blockchain.BlockchainAdapter for one-off queries.
func (c *Client) GetIncomingTransactions(ctx context.Context, address string, fromBlock int64) ([]blockchain.IncomingTx, error) {
	// TRON does not use block numbers for pagination in the TronGrid API —
	// it uses a fingerprint cursor. For monitoring use TronMonitor.Start().
	return nil, nil
}
