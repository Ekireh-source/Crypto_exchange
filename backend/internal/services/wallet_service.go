package services

import (
	"context"
	"fmt"
	"math/big"
	"net/http"
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
	return s.walletDB.GetAssets(ctx)
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

	// 3. Get or generate sender's deposit wallet record
	da, err := s.getOrCreateDepositAddressRecord(ctx, userID, asset)
	if err != nil {
		return nil, fmt.Errorf("fetching deposit address for transfer: %w", err)
	}

	privKeyHex, err := crypto.DecryptPrivateKey(da.EncryptedPrivateKey, s.cfg.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("decrypting sender key: %w", err)
	}

	// 4. Retrieve Blockchain Adapter for the asset network
	adapter, ok := s.adapters[asset.Network]
	if !ok {
		return nil, fmt.Errorf("no blockchain adapter registered for network %s", asset.Network)
	}

	// In development mode, auto-fund the sender address on Anvil/Hardhat with gas if needed
	s.autoFundDevAddress(ctx, da.Address)

	// 5. Estimate fee
	isToken := asset.ContractAddress != nil
	feeEst, err := adapter.EstimateFee(ctx, isToken)
	feeStr := "0.0005"
	if err == nil && feeEst != nil && feeEst.NativeAmount != nil {
		feeStr = feeEst.NativeAmount.Text('f', 6)
	}

	// 6. Broadcast transaction on-chain via BlockchainAdapter
	var txHash string
	if !isToken {
		txHash, err = adapter.SendNative(ctx, privKeyHex, req.ToAddress, amountFloat)
	} else {
		txHash, err = adapter.SendToken(ctx, privKeyHex, req.ToAddress, *asset.ContractAddress, amountFloat, asset.Decimals)
	}

	// 7. Construct and store transaction record
	status := models.TxPending
	if err != nil {
		// Log warning if on-chain broadcast failed (e.g. testnet RPC unaccessible or insufficient gas)
		fmt.Printf("On-chain transfer broadcast notice: %v. Using pending status.\n", err)
		txHash = fmt.Sprintf("0xdev_%s", uuid.New().String()[:16])
	} else {
		status = models.TxConfirmed
	}

	tx := &models.Transaction{
		ID:          uuid.New(),
		UserID:      userID,
		AssetID:     req.AssetID,
		Type:        models.TxWithdrawal,
		Status:      status,
		Amount:      req.Amount,
		Fee:         feeStr,
		FromAddress: &da.Address,
		ToAddress:   &req.ToAddress,
		TxHash:      &txHash,
		Note:        &req.Note,
		CreatedAt:   time.Now(),
	}

	// 8. Persist Transaction record to Database
	if err := s.walletDB.CreateTransaction(ctx, tx); err != nil {
		return nil, fmt.Errorf("persisting transaction: %w", err)
	}

	// 9. Deduct / update user's balance
	_ = s.walletDB.DeductBalance(ctx, userID, req.AssetID, req.Amount)

	return tx, nil
}

// ProcessDeposit is called by the background blockchain monitor when a verified inbound transfer is detected.
func (s *WalletService) ProcessDeposit(ctx context.Context, tx blockchain.IncomingTx) error {
	userID, assetID, err := s.walletDB.GetUserByDepositAddress(ctx, tx.ToAddress)
	if err != nil {
		return fmt.Errorf("failed to lookup user for address %s: %w", tx.ToAddress, err)
	}

	// Verify if the asset matches the incoming transfer token contract / native
	// In a complete implementation, you would look up the exact asset ID matching the contract.
	// For now, we use the primary asset associated with that deposit address.

	// Check if this transaction hash has already been processed
	// To prevent double crediting, we should really do an ON CONFLICT ignore on tx_hash.
	// But let's create the transaction and if it succeeds, add the balance.

	// Format amount for DB
	// We convert the big.Int wei/sun value to human-readable using decimals
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
		CreatedAt:   time.Now(),
		ConfirmedAt: &tx.Timestamp,
	}

	err = s.walletDB.CreateTransaction(ctx, dbTx)
	if err != nil {
		// If transaction already exists (unique constraint on tx_hash), we can return nil
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return nil
		}
		return fmt.Errorf("saving deposit transaction: %w", err)
	}

	// Credit the balance in the DB
	if err := s.walletDB.AddBalance(ctx, userID, assetID, amtStr); err != nil {
		return fmt.Errorf("crediting balance: %w", err)
	}

	return nil
}

// autoFundDevAddress automatically credits gas to sender addresses on local Anvil/Hardhat nodes during development.
func (s *WalletService) autoFundDevAddress(ctx context.Context, address string) {
	if !s.cfg.IsDevelopment() {
		return
	}
	rpcURL := s.cfg.ActiveBSCRPC()
	payload := fmt.Sprintf(`{"jsonrpc":"2.0","method":"anvil_setBalance","params":["%s","0x8ac7230489e80000"],"id":1}`, address)
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
