package services

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"time"

	"exchange/config"
	"exchange/internal/blockchain"
	"exchange/internal/blockchain/crypto"
	"exchange/internal/db/repository"
	"exchange/internal/models"
)

type SweeperService struct {
	cfg        *config.Config
	walletDB   *repository.WalletRepo
	priceSvc   *PriceService
	adapters   map[models.Network]blockchain.BlockchainAdapter
	pollTicker *time.Ticker
	quit       chan struct{}
}

func NewSweeperService(
	cfg *config.Config,
	walletDB *repository.WalletRepo,
	priceSvc *PriceService,
	adapters map[models.Network]blockchain.BlockchainAdapter,
) *SweeperService {
	return &SweeperService{
		cfg:      cfg,
		walletDB: walletDB,
		priceSvc: priceSvc,
		adapters: adapters,
		quit:     make(chan struct{}),
	}
}

func (s *SweeperService) Start(ctx context.Context) {
	log.Println("Starting Sweeper Service...")
	s.pollTicker = time.NewTicker(30 * time.Second)
	
	go func() {
		for {
			select {
			case <-s.pollTicker.C:
				s.processPendingSweeps(ctx)
			case <-s.quit:
				log.Println("Stopping Sweeper Service...")
				s.pollTicker.Stop()
				return
			case <-ctx.Done():
				log.Println("Context done, stopping Sweeper Service...")
				if s.pollTicker != nil {
					s.pollTicker.Stop()
				}
				return
			}
		}
	}()
}

func (s *SweeperService) Stop() {
	close(s.quit)
}

func (s *SweeperService) processPendingSweeps(ctx context.Context) {
	txs, err := s.walletDB.GetPendingSweeps(ctx)
	if err != nil {
		log.Printf("SweeperService ERROR fetching sweeps: %v", err)
		return
	}

	if len(txs) > 0 {
		log.Printf("SweeperService: Found %d pending sweeps to process", len(txs))
	}

	for _, tx := range txs {
		asset, err := s.walletDB.GetAssetByID(ctx, tx.AssetID)
		if err != nil || asset == nil {
			log.Printf("SweeperService ERROR fetching asset for tx %s: %v", tx.ID, err)
			continue
		}

		// 3.2 Minimum Value Check: skip deposits whose USD value is below the
		// sweep threshold to avoid paying more in gas than we recover.
		if s.priceSvc != nil && s.cfg.MinSweepUSDThreshold > 0 {
			prices, priceErr := s.priceSvc.GetPrices(ctx, nil)
			if priceErr == nil {
				if price, ok := prices[asset.Symbol]; ok && price > 0 {
					depositAmt, _, amtErr := new(big.Float).Parse(tx.Amount, 10)
					if amtErr == nil {
						depositUSD, _ := new(big.Float).Mul(depositAmt, big.NewFloat(price)).Float64()
						if depositUSD < s.cfg.MinSweepUSDThreshold {
							log.Printf("SweeperService: tx %s value $%.4f is below threshold $%.2f — skipping sweep",
								tx.ID, depositUSD, s.cfg.MinSweepUSDThreshold)
							continue
						}
					}
				}
			}
		}

		if asset.ContractAddress == nil {
			// Native coin sweep (BNB/TRX)
			err = s.processNativeSweep(ctx, &tx, asset)
		} else {
			// Token sweep (USDT)
			if tx.SweepStatus == models.SweepPendingGas {
				err = s.processTokenSweep(ctx, &tx, asset)
			} else {
				err = s.processTokenGasFunding(ctx, &tx, asset)
			}
		}

		if err != nil {
			log.Printf("SweeperService ERROR processing sweep for tx %s: %v", tx.ID, err)
		}
	}
}


func (s *SweeperService) processNativeSweep(ctx context.Context, tx *models.Transaction, asset *models.Asset) error {
	adapter, ok := s.adapters[asset.Network]
	if !ok {
		return fmt.Errorf("no adapter found for network %s", asset.Network)
	}

	da, err := s.walletDB.GetDepositAddress(ctx, tx.UserID, tx.AssetID)
	if err != nil || da == nil {
		return fmt.Errorf("fetching deposit address: %w", err)
	}

	privKeyHex, err := crypto.DecryptPrivateKey(da.EncryptedPrivateKey, s.cfg.EncryptionKey)
	if err != nil {
		return fmt.Errorf("decrypting key: %w", err)
	}

	bal, err := adapter.GetNativeBalance(ctx, da.Address)
	if err != nil {
		return fmt.Errorf("fetching native balance: %w", err)
	}

	feeEst, err := adapter.EstimateFee(ctx, false)
	if err != nil {
		return fmt.Errorf("estimating fee: %w", err)
	}

	sweepable := new(big.Float).Sub(bal, feeEst.NativeAmount)
	if sweepable.Cmp(big.NewFloat(0)) <= 0 {
		log.Printf("SweeperService: %s balance too low to sweep natively for tx %s (bal=%v, fee=%v)", asset.Symbol, tx.ID, bal, feeEst.NativeAmount)
		return s.walletDB.UpdateSweepStatus(ctx, tx.ID, models.SweepCompleted)
	}

	hotWalletAddr := s.cfg.HotWalletBSCAddress
	if asset.Network == models.NetworkTRON {
		hotWalletAddr = s.cfg.HotWalletTronAddress
	}

	log.Printf("SweeperService: Sweeping %v %s to hot wallet for tx %s", sweepable, asset.Symbol, tx.ID)

	txHash, err := adapter.SendNative(ctx, privKeyHex, hotWalletAddr, sweepable)
	if err != nil {
		return fmt.Errorf("broadcasting sweep: %w", err)
	}

	log.Printf("SweeperService: Sweep broadcasted! txHash: %s", txHash)
	return s.walletDB.UpdateSweepStatus(ctx, tx.ID, models.SweepCompleted)
}

// processTokenGasFunding handles funding a user's address with native gas from the Master Gas Wallet.
func (s *SweeperService) processTokenGasFunding(ctx context.Context, tx *models.Transaction, asset *models.Asset) error {
	adapter, ok := s.adapters[asset.Network]
	if !ok {
		return fmt.Errorf("no adapter found for network %s", asset.Network)
	}

	da, err := s.walletDB.GetDepositAddress(ctx, tx.UserID, tx.AssetID)
	if err != nil || da == nil {
		return fmt.Errorf("fetching deposit address: %w", err)
	}

	feeEst, err := adapter.EstimateFee(ctx, true)
	if err != nil {
		return fmt.Errorf("estimating token fee: %w", err)
	}

	bal, err := adapter.GetNativeBalance(ctx, da.Address)
	if err != nil {
		return fmt.Errorf("fetching native balance: %w", err)
	}

	if bal.Cmp(feeEst.NativeAmount) >= 0 {
		log.Printf("SweeperService: tx %s already has enough gas (bal=%v, need=%v)", tx.ID, bal, feeEst.NativeAmount)
		return s.walletDB.UpdateSweepStatus(ctx, tx.ID, models.SweepPendingGas)
	}

	gasNeeded := new(big.Float).Sub(feeEst.NativeAmount, bal)

	masterGasKey := s.cfg.MasterGasWalletBSCKey
	if asset.Network == models.NetworkTRON {
		masterGasKey = s.cfg.MasterGasWalletTronKey
	}

	log.Printf("SweeperService: Funding %v gas to %s for tx %s", gasNeeded, da.Address, tx.ID)

	txHash, err := adapter.SendNative(ctx, masterGasKey, da.Address, gasNeeded)
	if err != nil {
		return fmt.Errorf("broadcasting gas funding: %w", err)
	}

	log.Printf("SweeperService: Gas funding broadcasted! txHash: %s", txHash)
	return s.walletDB.UpdateSweepStatus(ctx, tx.ID, models.SweepPendingGas)
}

// processTokenSweep handles sweeping the actual tokens once gas is confirmed.
func (s *SweeperService) processTokenSweep(ctx context.Context, tx *models.Transaction, asset *models.Asset) error {
	adapter, ok := s.adapters[asset.Network]
	if !ok {
		return fmt.Errorf("no adapter found for network %s", asset.Network)
	}

	da, err := s.walletDB.GetDepositAddress(ctx, tx.UserID, tx.AssetID)
	if err != nil || da == nil {
		return fmt.Errorf("fetching deposit address: %w", err)
	}

	privKeyHex, err := crypto.DecryptPrivateKey(da.EncryptedPrivateKey, s.cfg.EncryptionKey)
	if err != nil {
		return fmt.Errorf("decrypting key: %w", err)
	}

	bal, err := adapter.GetTokenBalance(ctx, da.Address, *asset.ContractAddress, asset.Decimals)
	if err != nil {
		return fmt.Errorf("fetching token balance: %w", err)
	}

	if bal.Cmp(big.NewFloat(0)) <= 0 {
		return s.walletDB.UpdateSweepStatus(ctx, tx.ID, models.SweepCompleted)
	}

	hotWalletAddr := s.cfg.HotWalletBSCAddress
	if asset.Network == models.NetworkTRON {
		hotWalletAddr = s.cfg.HotWalletTronAddress
	}

	log.Printf("SweeperService: Sweeping %v %s tokens to hot wallet for tx %s", bal, asset.Symbol, tx.ID)

	txHash, err := adapter.SendToken(ctx, privKeyHex, hotWalletAddr, *asset.ContractAddress, bal, asset.Decimals)
	if err != nil {
		return fmt.Errorf("broadcasting token sweep: %w", err)
	}

	log.Printf("SweeperService: Token Sweep broadcasted! txHash: %s", txHash)
	return s.walletDB.UpdateSweepStatus(ctx, tx.ID, models.SweepCompleted)
}
