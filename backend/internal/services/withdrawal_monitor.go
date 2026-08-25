package services

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"time"

	"exchange/config"
	"exchange/internal/blockchain"
	"exchange/internal/blockchain/bsc"
	"exchange/internal/db/repository"
	"exchange/internal/models"
)

const (
	// withdrawalStalenessThreshold is how long a withdrawal must be pending
	// before we consider it "stuck" and check/retry it on-chain.
	withdrawalStalenessThreshold = 10 * time.Minute

	// withdrawalMaxAge is the maximum time we will attempt to retry a stuck tx.
	// After this, it will just sit in 'pending' so we don't retry infinitely on restarts.
	withdrawalMaxAge = 6 * time.Hour

	// withdrawalPollInterval is how often the monitor runs.
	withdrawalPollInterval = 5 * time.Minute

	// gasBumpPercent is the percentage increase applied to gas price on retry.
	// A 10% bump is sufficient on most EVM chains and TRON to unstick a tx.
	gasBumpPercent = 10
)

// WithdrawalMonitor is a background service that checks for pending withdrawals
// that may have been dropped by the mempool due to low gas, and re-broadcasts
// them with a higher gas price (Phase 5.3).
type WithdrawalMonitor struct {
	cfg       *config.Config
	walletDB  *repository.WalletRepo
	walletSvc *WalletService
	adapters  map[models.Network]blockchain.BlockchainAdapter
	pollTicker *time.Ticker
	quit       chan struct{}
}

// NewWithdrawalMonitor creates a WithdrawalMonitor.
func NewWithdrawalMonitor(
	cfg *config.Config,
	walletDB *repository.WalletRepo,
	walletSvc *WalletService,
	adapters map[models.Network]blockchain.BlockchainAdapter,
) *WithdrawalMonitor {
	return &WithdrawalMonitor{
		cfg:       cfg,
		walletDB:  walletDB,
		walletSvc: walletSvc,
		adapters:  adapters,
		quit:      make(chan struct{}),
	}
}

// Start begins the polling loop in a background goroutine.
func (m *WithdrawalMonitor) Start(ctx context.Context) {
	log.Println("WithdrawalMonitor: starting (poll every 5m)")
	m.pollTicker = time.NewTicker(withdrawalPollInterval)

	go func() {
		// Run once immediately on startup to handle any pre-existing stuck txs.
		m.processStuckWithdrawals(ctx)

		for {
			select {
			case <-m.pollTicker.C:
				m.processStuckWithdrawals(ctx)
			case <-m.quit:
				log.Println("WithdrawalMonitor: stopping")
				m.pollTicker.Stop()
				return
			case <-ctx.Done():
				log.Println("WithdrawalMonitor: context done, stopping")
				if m.pollTicker != nil {
					m.pollTicker.Stop()
				}
				return
			}
		}
	}()
}

// Stop signals the monitor to shut down.
func (m *WithdrawalMonitor) Stop() {
	close(m.quit)
}

// processStuckWithdrawals fetches pending withdrawals that haven't confirmed
// within withdrawalStalenessThreshold and acts on each one.
func (m *WithdrawalMonitor) processStuckWithdrawals(ctx context.Context) {
	txs, err := m.walletDB.GetStuckWithdrawals(ctx, withdrawalStalenessThreshold, withdrawalMaxAge)
	if err != nil {
		log.Printf("WithdrawalMonitor ERROR fetching stuck withdrawals: %v", err)
		return
	}

	if len(txs) > 0 {
		log.Printf("WithdrawalMonitor: found %d stuck withdrawal(s) to check", len(txs))
	}

	for _, tx := range txs {
		if err := m.handleStuckWithdrawal(ctx, &tx); err != nil {
			log.Printf("WithdrawalMonitor ERROR handling tx %s: %v", tx.ID, err)
		}
	}
}

// handleStuckWithdrawal processes a single potentially-stuck withdrawal:
//  1. Checks on-chain status via GetTxStatus.
//  2. If confirmed -> update DB to 'confirmed'.
//  3. If still pending -> re-broadcast with a bumped gas price and update tx_hash.
func (m *WithdrawalMonitor) handleStuckWithdrawal(ctx context.Context, tx *models.Transaction) error {
	if tx.TxHash == nil || *tx.TxHash == "" {
		log.Printf("WithdrawalMonitor: tx %s has no tx_hash, skipping", tx.ID)
		return nil
	}

	asset, err := m.walletDB.GetAssetByID(ctx, tx.AssetID)
	if err != nil || asset == nil {
		return fmt.Errorf("fetching asset for tx %s: %w", tx.ID, err)
	}

	adapter, ok := m.adapters[asset.Network]
	if !ok {
		return fmt.Errorf("no adapter for network %s", asset.Network)
	}

	// 1. Check on-chain status.
	confirmed, err := adapter.GetTxStatus(ctx, *tx.TxHash)
	if err != nil {
		return fmt.Errorf("checking on-chain status for tx %s: %w", tx.ID, err)
	}

	if confirmed {
		// The tx landed on-chain — mark it confirmed in the DB.
		log.Printf("WithdrawalMonitor: tx %s (on-chain: %s) is now confirmed — updating DB", tx.ID, *tx.TxHash)
		return m.walletDB.UpdateTransactionStatus(ctx, tx.ID, models.TxConfirmed, *tx.TxHash)
	}

	// Disable automatic retries for TRON to prevent double spends due to lack of RBF (replace-by-fee).
	if asset.Network == models.NetworkTRON {
		log.Printf("WithdrawalMonitor: tx %s (on-chain: %s) is still pending on TRON. Automatic retry disabled to prevent double-spends.", tx.ID, *tx.TxHash)
		return nil
	}

	// 2. Still pending — re-broadcast with bumped gas (EVM chains only).
	log.Printf("WithdrawalMonitor: tx %s (on-chain: %s) is still pending — re-broadcasting with %d%% gas bump",
		tx.ID, *tx.TxHash, gasBumpPercent)

	newHash, err := m.rebroadcast(ctx, tx, asset, adapter)
	if err != nil {
		return fmt.Errorf("re-broadcasting tx %s: %w", tx.ID, err)
	}

	log.Printf("WithdrawalMonitor: tx %s re-broadcast successful, new hash: %s", tx.ID, newHash)
	return m.walletDB.UpdateWithdrawalTxHash(ctx, tx.ID, newHash)
}

// rebroadcast re-submits a withdrawal from the Hot Wallet with a bumped gas price.
// For BSC (EVM), the same nonce is reused so the new tx replaces the stuck one.
// For TRON, a new transaction is created (TRON does not support nonce-replacement).
func (m *WithdrawalMonitor) rebroadcast(
	ctx context.Context,
	tx *models.Transaction,
	asset *models.Asset,
	adapter blockchain.BlockchainAdapter,
) (string, error) {
	amountFloat, ok := new(big.Float).SetString(tx.Amount)
	if !ok {
		return "", fmt.Errorf("parsing amount %q for tx %s", tx.Amount, tx.ID)
	}

	if tx.ToAddress == nil || *tx.ToAddress == "" {
		return "", fmt.Errorf("tx %s missing to_address", tx.ID)
	}

	// For BSC, use the bump-aware sender that injects an elevated gas price.
	if asset.Network == models.NetworkBSC {
		bscAdapter, ok := adapter.(*bsc.Client)
		if !ok {
			return "", fmt.Errorf("expected *bsc.Client for BSC network, got %T", adapter)
		}

		hotWalletKey := m.cfg.HotWalletBSCKey
		if hotWalletKey == "" {
			return "", fmt.Errorf("HOT_WALLET_BSC_KEY not configured")
		}

		isToken := asset.ContractAddress != nil
		return bscAdapter.SendWithBumpedGas(ctx, hotWalletKey, *tx.ToAddress, amountFloat, asset.ContractAddress, asset.Decimals, gasBumpPercent, isToken)
	}

	// TRON fallback: create a fresh transaction (retry).
	hotWalletKey := m.cfg.HotWalletTronKey
	if hotWalletKey == "" {
		return "", fmt.Errorf("HOT_WALLET_TRON_KEY not configured")
	}

	if asset.ContractAddress != nil {
		return adapter.SendToken(ctx, hotWalletKey, *tx.ToAddress, *asset.ContractAddress, amountFloat, asset.Decimals)
	}
	return adapter.SendNative(ctx, hotWalletKey, *tx.ToAddress, amountFloat)
}
