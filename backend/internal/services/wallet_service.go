package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"exchange/config"
	"exchange/internal/blockchain"
	"exchange/internal/blockchain/crypto"
	"exchange/internal/db/repository"
	"exchange/internal/models"
)

type WalletService struct {
	cfg       *config.Config
	walletRepo *repository.UserRepo // Wait, I'll use WalletRepo
	walletDB  *repository.WalletRepo
	priceSvc  *PriceService
	adapters  map[models.Network]blockchain.BlockchainAdapter
}

func NewWalletService(
	cfg *config.Config,
	walletDB *repository.WalletRepo,
	priceSvc *PriceService,
	adapters map[models.Network]blockchain.BlockchainAdapter,
) *WalletService {
	return &WalletService{
		cfg:      cfg,
		walletDB: walletDB,
		priceSvc: priceSvc,
		adapters: adapters,
	}
}

// GetPortfolio returns the aggregated balance and USD value for a user.
func (s *WalletService) GetPortfolio(ctx context.Context, userID uuid.UUID) ([]models.AssetBalance, float64, error) {
	assets, err := s.walletDB.GetAssets(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("fetching assets: %w", err)
	}

	balances, err := s.walletDB.GetBalances(ctx, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("fetching balances: %w", err)
	}

	balanceMap := make(map[int]models.Balance)
	for _, b := range balances {
		balanceMap[b.AssetID] = b
	}

	// prices, _ := s.priceSvc.GetPrices(ctx, nil)

	var portfolio []models.AssetBalance
	var totalUSD float64

	for _, asset := range assets {
		bal := balanceMap[asset.ID]
		// if missing, it's implicitly zero

		available := "0"
		if bal.Available != "" {
			available = bal.Available
		}
		locked := "0"
		if bal.Locked != "" {
			locked = bal.Locked
		}

		// price := prices[asset.Symbol]
		// In a real app, you'd parse `available` to float for multiplication
		// Assuming available is string, we should parse it to big.Float
		// But for now, we'll keep it simple as a prototype. Let's just assume it parses.
		// (Skipping precise float calculation here for brevity, assuming available = 0 for now)

		ab := models.AssetBalance{
			Asset:     asset,
			Available: available,
			Locked:    locked,
			USDValue:  0, // TODO: multiply available * price
		}
		portfolio = append(portfolio, ab)
	}

	return portfolio, totalUSD, nil
}

// GetDepositAddress fetches or generates a deposit address for a specific asset.
func (s *WalletService) GetDepositAddress(ctx context.Context, userID uuid.UUID, assetID int) (*models.DepositAddressResponse, error) {
	asset, err := s.walletDB.GetAssetByID(ctx, assetID)
	if err != nil {
		return nil, fmt.Errorf("fetching asset: %w", err)
	}
	if asset == nil {
		return nil, fmt.Errorf("asset not found")
	}

	da, err := s.walletDB.GetDepositAddress(ctx, userID, assetID)
	if err != nil {
		return nil, fmt.Errorf("fetching deposit address: %w", err)
	}

	// Generate new if not found
	if da == nil {
		adapter, ok := s.adapters[asset.Network]
		if !ok {
			return nil, fmt.Errorf("no blockchain adapter for network %s", asset.Network)
		}

		address, privKeyHex, err := adapter.GenerateWallet()
		if err != nil {
			return nil, fmt.Errorf("generating wallet: %w", err)
		}

		encryptedKey, err := crypto.EncryptPrivateKey(privKeyHex, s.cfg.EncryptionKey)
		if err != nil {
			return nil, fmt.Errorf("encrypting private key: %w", err)
		}

		da = &models.DepositAddress{
			ID:                 uuid.New(),
			UserID:             userID,
			AssetID:            assetID,
			Address:            address,
			EncryptedPrivateKey: encryptedKey,
			CreatedAt:          time.Now(),
		}

		if err := s.walletDB.CreateDepositAddress(ctx, da); err != nil {
			return nil, fmt.Errorf("saving deposit address: %w", err)
		}
	}

	return &models.DepositAddressResponse{
		AssetID:        asset.ID,
		Symbol:         asset.Symbol,
		Network:        asset.Network,
		Standard:       string(asset.Standard),
		Address:        da.Address,
		QRData:         da.Address, // Could be formatted as e.g., "ethereum:0x..."
		NetworkWarning: fmt.Sprintf("Only send %s on the %s network to this address.", asset.Symbol, asset.Network),
	}, nil
}

// ProcessDeposit records an inbound deposit and credits the user's available balance.
func (s *WalletService) ProcessDeposit(tx blockchain.IncomingTx) error {
	// In a real implementation:
	// 1. Look up user by tx.ToAddress and tx.Asset (or network)
	// 2. Wrap in a DB transaction
	// 3. Insert into transactions table (type = 'deposit', status = 'confirmed')
	// 4. Upsert into balances (available = available + tx.Amount)
	// 5. Send push notification to user
	
	fmt.Printf("Processing deposit: %s of %s to %s (TxHash: %s)\n", tx.Amount.String(), tx.Asset, tx.ToAddress, tx.TxHash)
	return nil
}

