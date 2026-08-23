package services

import (
	"context"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"exchange/config"
	"exchange/internal/blockchain"
	"exchange/internal/blockchain/crypto"
	"exchange/internal/db/repository"
	"exchange/internal/models"
)

type WalletService struct {
	cfg        *config.Config
	walletRepo *repository.UserRepo // Wait, I'll use WalletRepo
	walletDB   *repository.WalletRepo
	priceSvc   *PriceService
	adapters   map[models.Network]blockchain.BlockchainAdapter
	scanners   map[models.Network]blockchain.DepositScanner
}

func NewWalletService(
	cfg *config.Config,
	walletDB *repository.WalletRepo,
	priceSvc *PriceService,
	adapters map[models.Network]blockchain.BlockchainAdapter,
	scanners map[models.Network]blockchain.DepositScanner,
) *WalletService {
	return &WalletService{
		cfg:      cfg,
		walletDB: walletDB,
		priceSvc: priceSvc,
		adapters: adapters,
		scanners: scanners,
	}
}

// GetAssets fetches all active supported assets.
func (s *WalletService) GetAssets(ctx context.Context) ([]models.Asset, error) {
	assets, err := s.walletDB.GetAssets(ctx)
	if err != nil {
		return nil, err
	}
	
	if s.priceSvc != nil {
		prices, _ := s.priceSvc.GetPrices(ctx, nil)
		for i := range assets {
			assets[i].CurrentPrice = prices[assets[i].Symbol]
		}
	}
	
	return assets, nil
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

	var prices map[string]float64
	if s.priceSvc != nil {
		prices, _ = s.priceSvc.GetPrices(ctx, nil)
	} else {
		prices = make(map[string]float64)
	}

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

		price := prices[asset.Symbol]
		
		// Parse available as float64 to calculate USD value
		availableFloat, _ := strconv.ParseFloat(available, 64)
		usdValue := availableFloat * price
		totalUSD += usdValue

		// Also populate CurrentPrice so the portfolio includes the live price
		asset.CurrentPrice = price

		ab := models.AssetBalance{
			Asset:     asset,
			Available: available,
			Locked:    locked,
			USDValue:  usdValue,
		}
		portfolio = append(portfolio, ab)
	}

	return portfolio, totalUSD, nil
}

// getOrCreateDepositAddressRecord handles DB creation of the deposit address.
func (s *WalletService) getOrCreateDepositAddressRecord(ctx context.Context, userID uuid.UUID, asset *models.Asset) (*models.DepositAddress, error) {
	da, err := s.walletDB.GetDepositAddress(ctx, userID, asset.ID)
	if err != nil {
		return nil, fmt.Errorf("checking existing address: %w", err)
	}
	if da != nil {
		return da, nil // Already exists
	}

	adapter, ok := s.adapters[asset.Network]
	if !ok {
		return nil, fmt.Errorf("no blockchain adapter for network %s", asset.Network)
	}

	address, privKeyHex, err := adapter.GenerateWallet()
	if err != nil {
		return nil, fmt.Errorf("generating wallet: %w", err)
	}

	encKey, err := crypto.EncryptPrivateKey(privKeyHex, s.cfg.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("encrypting private key: %w", err)
	}

	newDA := &models.DepositAddress{
		ID:                  uuid.New(),
		UserID:              userID,
		AssetID:             asset.ID,
		Address:             address,
		EncryptedPrivateKey: encKey,
		CreatedAt:           time.Now(),
	}

	if err := s.walletDB.CreateDepositAddress(ctx, newDA); err != nil {
		return nil, fmt.Errorf("saving deposit address: %w", err)
	}

	// In development mode, auto-fund the Hot Wallet on Anvil/Hardhat so it can send out withdrawals.
	// Only relevant for BSC — TRON uses a different testnet (Shasta).
	if asset.Network == models.NetworkBSC {
		s.autoFundDevAddress(ctx, s.cfg.HotWalletBSCAddress)
	}

	// Tell the blockchain scanner to start watching this new address immediately
	if s.scanners != nil {
		if scanner, ok := s.scanners[asset.Network]; ok {
			scanner.AddAddress(address)
		}
	}

	return newDA, nil
}

// GetDepositAddress fetches or generates a deposit address response for a specific asset.
func (s *WalletService) GetDepositAddress(ctx context.Context, userID uuid.UUID, assetID int) (*models.DepositAddressResponse, error) {
	asset, err := s.walletDB.GetAssetByID(ctx, assetID)
	if err != nil {
		return nil, fmt.Errorf("fetching asset: %w", err)
	}
	if asset == nil {
		return nil, fmt.Errorf("asset not found")
	}

	da, err := s.getOrCreateDepositAddressRecord(ctx, userID, asset)
	if err != nil {
		return nil, err
	}

	return &models.DepositAddressResponse{
		AssetID:        asset.ID,
		Symbol:         asset.Symbol,
		Network:        asset.Network,
		Standard:       string(asset.Standard),
		Address:        da.Address,
		QRData:         da.Address,
		NetworkWarning: fmt.Sprintf("Only send %s on the %s network to this address.", asset.Symbol, asset.Network),
	}, nil
}

// SendRequest represents a request to transfer crypto.
type SendRequest struct {
	AssetID   int    `json:"asset_id"`
	ToAddress string `json:"to_address"`
	Amount    string `json:"amount"`
	Note      string `json:"note"`
}

// SendCrypto processes an outbound crypto transfer by interacting with the blockchain adapter.
func (s *WalletService) SendCrypto(ctx context.Context, userID uuid.UUID, req SendRequest) (*models.Transaction, error) {
	// 1. Validate amount
	amountFloat, ok := new(big.Float).SetString(req.Amount)
	if !ok || amountFloat.Cmp(big.NewFloat(0)) <= 0 {
		return nil, fmt.Errorf("invalid amount: %s", req.Amount)
	}

	// 2. Fetch Asset details
	asset, err := s.walletDB.GetAssetByID(ctx, req.AssetID)
	if err != nil {
		return nil, fmt.Errorf("fetching asset: %w", err)
	}
	if asset == nil {
		return nil, fmt.Errorf("asset not found")
	}

	// 3. Check user's balance
	balances, err := s.walletDB.GetBalances(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("fetching balances: %w", err)
	}
	
	var userBal *big.Float
	for _, b := range balances {
		if b.AssetID == req.AssetID {
			userBal, _ = new(big.Float).SetString(b.Available)
			break
		}
	}
	if userBal == nil {
		userBal = big.NewFloat(0)
	}

	// 4. Calculate Fee in Token (using Price Service)
	var feeAmt float64 = 0
	if s.priceSvc != nil {
		prices, err := s.priceSvc.GetPrices(ctx, nil)
		if err == nil {
			if price, exists := prices[asset.Symbol]; exists && price > 0 {
				feeAmt = s.cfg.WithdrawalFeeUSD / price
			}
		}
	}
	feeFloat := big.NewFloat(feeAmt)

	// 5. Total required = amount + fee
	totalRequired := new(big.Float).Add(amountFloat, feeFloat)
	if userBal.Cmp(totalRequired) < 0 {
		return nil, fmt.Errorf("insufficient balance. You need %v %s (including %v fee)", totalRequired, asset.Symbol, feeFloat)
	}

	// 6. Retrieve Blockchain Adapter and Hot Wallet
	adapter, ok := s.adapters[asset.Network]
	if !ok {
		return nil, fmt.Errorf("no blockchain adapter registered for network %s", asset.Network)
	}

	hotWalletPrivKey := s.cfg.HotWalletBSCKey
	hotWalletAddr := s.cfg.HotWalletBSCAddress
	if asset.Network == models.NetworkTRON {
		hotWalletPrivKey = s.cfg.HotWalletTronKey
		hotWalletAddr = s.cfg.HotWalletTronAddress
	}

	if hotWalletPrivKey == "" {
		return nil, fmt.Errorf("hot wallet not configured for network %s", asset.Network)
	}

	// 7. Broadcast transaction on-chain via BlockchainAdapter from the Hot Wallet
	var txHash string
	isToken := asset.ContractAddress != nil
	if !isToken {
		txHash, err = adapter.SendNative(ctx, hotWalletPrivKey, req.ToAddress, amountFloat)
	} else {
		txHash, err = adapter.SendToken(ctx, hotWalletPrivKey, req.ToAddress, *asset.ContractAddress, amountFloat, asset.Decimals)
	}

	if err != nil {
		return nil, fmt.Errorf("on-chain broadcast failed: %w", err)
	}

	// 8. Construct and store transaction record
	status := models.TxConfirmed
	totalDeduction := totalRequired.Text('f', 8)

	tx := &models.Transaction{
		ID:          uuid.New(),
		UserID:      userID,
		AssetID:     req.AssetID,
		Type:        models.TxWithdrawal,
		Status:      status,
		Amount:      req.Amount,
		Fee:         feeFloat.Text('f', 8),
		FromAddress: &hotWalletAddr,
		ToAddress:   &req.ToAddress,
		TxHash:      &txHash,
		Note:        &req.Note,
		SweepStatus: models.SweepNotNeeded,
		CreatedAt:   time.Now(),
	}

	if err := s.walletDB.CreateTransaction(ctx, tx); err != nil {
		return nil, fmt.Errorf("persisting transaction: %w", err)
	}

	// 9. Deduct the full amount + fee from the user's DB balance
	_ = s.walletDB.DeductBalance(ctx, userID, req.AssetID, totalDeduction)

	return tx, nil
}

// GetTransactionByID fetches a single transaction by its ID for the authenticated user.
func (s *WalletService) GetTransactionByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.Transaction, error) {
	return s.walletDB.GetTransactionByID(ctx, id, userID)
}

// GetSwapByID fetches a single swap by its ID for the authenticated user.
func (s *WalletService) GetSwapByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.SwapRecord, error) {
	return s.walletDB.GetSwapByID(ctx, id, userID)
}

// GetWatchlist returns a list of watched asset IDs.
func (s *WalletService) GetWatchlist(ctx context.Context, userID uuid.UUID) ([]int, error) {
	return s.walletDB.GetWatchlist(ctx, userID)
}

// ToggleWatchlist toggles the watched status of an asset.
func (s *WalletService) ToggleWatchlist(ctx context.Context, userID uuid.UUID, assetID int) error {
	return s.walletDB.ToggleWatchlist(ctx, userID, assetID)
}

// ProcessDeposit is called by the background blockchain monitor when a verified inbound transfer is detected.
func (s *WalletService) ProcessDeposit(ctx context.Context, tx blockchain.IncomingTx) error {
	userID, addrAssetID, err := s.walletDB.GetUserByDepositAddress(ctx, tx.ToAddress)
	if err != nil {
		return fmt.Errorf("failed to lookup user for address %s: %w", tx.ToAddress, err)
	}

	// Resolve the correct assetID for this deposit.
	//
	// A deposit address is stored per-asset, so addrAssetID is the native asset
	// (e.g. BNB or TRX).  When a token arrives (tx.Asset != "NATIVE") we must
	// look up the asset by its contract address so the right balance is credited.
	assetID := addrAssetID
	if tx.Asset != "NATIVE" && tx.Asset != "" {
		tokenAsset, err := s.walletDB.GetAssetByContractAddress(ctx, tx.Asset, models.Network(tx.Network))
		if err != nil {
			return fmt.Errorf("resolving token asset for contract %s: %w", tx.Asset, err)
		}
		if tokenAsset == nil {
			// We received a token we don't support — log and skip rather than error.
			fmt.Printf("ProcessDeposit: unsupported token contract %s on %s — ignoring deposit\n", tx.Asset, tx.Network)
			return nil
		}
		assetID = tokenAsset.ID
	}

	// Convert the raw on-chain integer amount to a human-readable decimal string.
	denom := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(tx.Decimals)), nil))
	floatAmt := new(big.Float).SetInt(tx.Amount)
	humanAmt := new(big.Float).Quo(floatAmt, denom)
	amtStr := humanAmt.Text('f', -1)

	txHash := tx.TxHash
	fromAddr := tx.FromAddress
	toAddr := tx.ToAddress

	dbTx := &models.Transaction{
		ID:          uuid.New(),
		UserID:      userID,
		AssetID:     assetID,
		Type:        models.TxDeposit,
		Status:      models.TxConfirmed,
		Amount:      amtStr,
		Fee:         "0",
		TxHash:      &txHash,
		FromAddress: &fromAddr,
		ToAddress:   &toAddr,
		SweepStatus: models.SweepPendingSweep,
		CreatedAt:   time.Now(),
		ConfirmedAt: &tx.Timestamp,
	}

	fmt.Printf("ProcessDeposit: saving %s %s deposit (tx: %s, user: %s, assetID: %d)\n",
		amtStr, tx.Asset, txHash, userID, assetID)

	err = s.walletDB.CreateTransaction(ctx, dbTx)
	if err != nil {
		// Deduplicate: if the tx_hash was already processed, skip silently.
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			fmt.Printf("ProcessDeposit: tx %s already processed — skipping\n", txHash)
			return nil
		}
		return fmt.Errorf("saving deposit transaction: %w", err)
	}

	// Credit the balance in the DB.
	if err := s.walletDB.AddBalance(ctx, userID, assetID, amtStr); err != nil {
		return fmt.Errorf("crediting balance for tx %s: %w", txHash, err)
	}

	fmt.Printf("ProcessDeposit: credited %s to user %s (assetID: %d)\n", amtStr, userID, assetID)
	return nil
}

// autoFundDevAddress automatically credits gas to addresses on local Anvil nodes during development.
func (s *WalletService) autoFundDevAddress(ctx context.Context, address string) {
	if !s.cfg.IsDevelopment() {
		return
	}
	rpcURL := s.cfg.ActiveBSCRPC()
	// Fund with 100 BNB for testing
	payload := fmt.Sprintf(`{"jsonrpc":"2.0","method":"anvil_setBalance","params":["%s","0x56bc75e2d63100000"],"id":1}`, address)
	req, err := http.NewRequestWithContext(ctx, "POST", rpcURL, strings.NewReader(payload))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Do(req)
	if err == nil && resp != nil {
		resp.Body.Close()
	}
}

// GetTransactions fetches paginated transactions for a user.
func (s *WalletService) GetTransactions(ctx context.Context, userID uuid.UUID, limit, page int, txType string) ([]models.Transaction, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	return s.walletDB.GetTransactions(ctx, userID, limit, offset, txType)
}

// SwapCrypto executes an internal market swap between two assets.
func (s *WalletService) SwapCrypto(ctx context.Context, userID uuid.UUID, req models.SwapRequest) (*models.Swap, error) {
	if req.FromAssetID == req.ToAssetID {
		return nil, fmt.Errorf("cannot swap the same asset")
	}

	fromAmount, ok := new(big.Float).SetString(req.Amount)
	if !ok || fromAmount.Cmp(big.NewFloat(0)) <= 0 {
		return nil, fmt.Errorf("invalid swap amount")
	}

	fromAsset, err := s.walletDB.GetAssetByID(ctx, req.FromAssetID)
	if err != nil || fromAsset == nil {
		return nil, fmt.Errorf("invalid from_asset")
	}

	toAsset, err := s.walletDB.GetAssetByID(ctx, req.ToAssetID)
	if err != nil || toAsset == nil {
		return nil, fmt.Errorf("invalid to_asset")
	}

	// Verify balance
	balances, err := s.walletDB.GetBalances(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("fetching balances: %w", err)
	}
	var userBal *big.Float
	for _, b := range balances {
		if b.AssetID == req.FromAssetID {
			userBal, _ = new(big.Float).SetString(b.Available)
			break
		}
	}
	if userBal == nil || userBal.Cmp(fromAmount) < 0 {
		return nil, fmt.Errorf("insufficient balance for swap")
	}

	// Fetch prices
	if s.priceSvc == nil {
		return nil, fmt.Errorf("pricing service unavailable")
	}
	prices, err := s.priceSvc.GetPrices(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("fetching prices: %w", err)
	}
	fromPrice := prices[fromAsset.Symbol]
	toPrice := prices[toAsset.Symbol]

	if fromPrice <= 0 || toPrice <= 0 {
		return nil, fmt.Errorf("market prices unavailable for swap pair")
	}

	// Calculate swap values
	exchangeRate := fromPrice / toPrice
	grossToAmount := new(big.Float).Mul(fromAmount, big.NewFloat(exchangeRate))
	
	feeRate := s.cfg.ExchangeFeeRate // e.g., 0.001 for 0.1%
	feeAmount := new(big.Float).Mul(grossToAmount, big.NewFloat(feeRate))
	netToAmount := new(big.Float).Sub(grossToAmount, feeAmount)

	swap := &models.Swap{
		UserID:       userID,
		FromAssetID:  req.FromAssetID,
		ToAssetID:    req.ToAssetID,
		FromAmount:   fromAmount.Text('f', 8),
		ToAmount:     netToAmount.Text('f', 8),
		ExchangeRate: fmt.Sprintf("%f", exchangeRate),
		FeeAmount:    feeAmount.Text('f', 8),
	}

	if err := s.walletDB.ExecuteSwap(ctx, swap); err != nil {
		return nil, fmt.Errorf("executing swap: %w", err)
	}

	return swap, nil
}
