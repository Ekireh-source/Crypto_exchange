package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"exchange/internal/db"
	"exchange/internal/models"
)

type WalletRepo struct {
	pool *db.Pool
}

func NewWalletRepo(pool *db.Pool) *WalletRepo {
	return &WalletRepo{pool: pool}
}

// GetAssets fetches all active assets.
func (r *WalletRepo) GetAssets(ctx context.Context) ([]models.Asset, error) {
	query := `
		SELECT id, symbol, name, network, standard, contract_address, decimals, is_active, logo_url
		FROM assets
		WHERE is_active = true
		ORDER BY id ASC
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("querying assets: %w", err)
	}
	defer rows.Close()

	var assets []models.Asset
	for rows.Next() {
		var a models.Asset
		if err := rows.Scan(
			&a.ID, &a.Symbol, &a.Name, &a.Network, &a.Standard,
			&a.ContractAddress, &a.Decimals, &a.IsActive, &a.LogoURL,
		); err != nil {
			return nil, fmt.Errorf("scanning asset: %w", err)
		}
		assets = append(assets, a)
	}
	return assets, nil
}

// GetAssetByID fetches a single asset by its ID.
func (r *WalletRepo) GetAssetByID(ctx context.Context, id int) (*models.Asset, error) {
	query := `
		SELECT id, symbol, name, network, standard, contract_address, decimals, is_active, logo_url
		FROM assets
		WHERE id = $1
	`
	var a models.Asset
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&a.ID, &a.Symbol, &a.Name, &a.Network, &a.Standard,
		&a.ContractAddress, &a.Decimals, &a.IsActive, &a.LogoURL,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // not found
		}
		return nil, fmt.Errorf("querying asset by id: %w", err)
	}
	return &a, nil
}

// GetAssetByContractAddress looks up an asset by its on-chain contract address and network.
// Used during deposit processing to resolve the correct asset for token transfers.
func (r *WalletRepo) GetAssetByContractAddress(ctx context.Context, contractAddress string, network models.Network) (*models.Asset, error) {
	query := `
		SELECT id, symbol, name, network, standard, contract_address, decimals, is_active, logo_url
		FROM assets
		WHERE LOWER(contract_address) = LOWER($1)
		  AND network = $2
		  AND is_active = true
		LIMIT 1
	`
	var a models.Asset
	err := r.pool.QueryRow(ctx, query, contractAddress, network).Scan(
		&a.ID, &a.Symbol, &a.Name, &a.Network, &a.Standard,
		&a.ContractAddress, &a.Decimals, &a.IsActive, &a.LogoURL,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // not found
		}
		return nil, fmt.Errorf("querying asset by contract address: %w", err)
	}
	return &a, nil
}

// GetBalances fetches all balances for a user.
func (r *WalletRepo) GetBalances(ctx context.Context, userID uuid.UUID) ([]models.Balance, error) {
	query := `
		SELECT user_id, asset_id, available, locked
		FROM balances
		WHERE user_id = $1
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("querying balances: %w", err)
	}
	defer rows.Close()

	var balances []models.Balance
	for rows.Next() {
		var b models.Balance
		if err := rows.Scan(&b.UserID, &b.AssetID, &b.Available, &b.Locked); err != nil {
			return nil, fmt.Errorf("scanning balance: %w", err)
		}
		balances = append(balances, b)
	}
	return balances, nil
}

// GetDepositAddress fetches a generated deposit address for a specific user and asset.
func (r *WalletRepo) GetDepositAddress(ctx context.Context, userID uuid.UUID, assetID int) (*models.DepositAddress, error) {
	query := `
		SELECT id, user_id, asset_id, address, encrypted_private_key, created_at
		FROM deposit_addresses
		WHERE user_id = $1 AND asset_id = $2
	`
	var da models.DepositAddress
	err := r.pool.QueryRow(ctx, query, userID, assetID).Scan(
		&da.ID, &da.UserID, &da.AssetID, &da.Address, &da.EncryptedPrivateKey, &da.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // not found
		}
		return nil, fmt.Errorf("querying deposit address: %w", err)
	}
	return &da, nil
}

// GetAllDepositAddresses fetches all deposit addresses in the system.
func (r *WalletRepo) GetAllDepositAddresses(ctx context.Context) ([]models.DepositAddress, error) {
	query := `
		SELECT id, user_id, asset_id, address, encrypted_private_key, created_at
		FROM deposit_addresses
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("querying all deposit addresses: %w", err)
	}
	defer rows.Close()

	var das []models.DepositAddress
	for rows.Next() {
		var da models.DepositAddress
		if err := rows.Scan(
			&da.ID, &da.UserID, &da.AssetID, &da.Address, &da.EncryptedPrivateKey, &da.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning deposit address: %w", err)
		}
		das = append(das, da)
	}
	return das, nil
}

// GetUserByDepositAddress returns the user ID and asset ID associated with a given deposit address.
func (r *WalletRepo) GetUserByDepositAddress(ctx context.Context, address string) (uuid.UUID, int, error) {
	query := `
		SELECT user_id, asset_id
		FROM deposit_addresses
		WHERE LOWER(address) = LOWER($1)
	`
	var userID uuid.UUID
	var assetID int
	err := r.pool.QueryRow(ctx, query, address).Scan(&userID, &assetID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return uuid.Nil, 0, fmt.Errorf("address not found: %s", address)
		}
		return uuid.Nil, 0, fmt.Errorf("querying user by address: %w", err)
	}
	return userID, assetID, nil
}

// CreateDepositAddress stores a newly generated deposit address.
func (r *WalletRepo) CreateDepositAddress(ctx context.Context, da *models.DepositAddress) error {
	query := `
		INSERT INTO deposit_addresses (id, user_id, asset_id, address, encrypted_private_key, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.pool.Exec(ctx, query,
		da.ID, da.UserID, da.AssetID, da.Address, da.EncryptedPrivateKey, da.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting deposit address: %w", err)
	}
	return nil
}

// GetTransactions fetches paginated transactions for a user with optional type filtering.
func (r *WalletRepo) GetTransactions(ctx context.Context, userID uuid.UUID, limit, offset int, txType string) ([]models.Transaction, int, error) {
	countQuery := `SELECT COUNT(*) FROM transactions WHERE user_id = $1`
	args := []interface{}{userID}
	if txType != "" {
		countQuery += ` AND type = $2`
		args = append(args, txType)
	}

	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting transactions: %w", err)
	}

	query := `
		SELECT id, user_id, asset_id, type, status, amount, fee, tx_hash, from_address, to_address, note, sweep_status, created_at, confirmed_at
		FROM transactions
		WHERE user_id = $1
	`
	queryArgs := []interface{}{userID}
	argIdx := 2

	if txType != "" {
		query += fmt.Sprintf(` AND type = $%d`, argIdx)
		queryArgs = append(queryArgs, txType)
		argIdx++
	}

	query += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
	queryArgs = append(queryArgs, limit, offset)

	rows, err := r.pool.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("querying transactions: %w", err)
	}
	defer rows.Close()

	var txs []models.Transaction
	for rows.Next() {
		var tx models.Transaction
		if err := rows.Scan(
			&tx.ID, &tx.UserID, &tx.AssetID, &tx.Type, &tx.Status,
			&tx.Amount, &tx.Fee, &tx.TxHash, &tx.FromAddress, &tx.ToAddress,
			&tx.Note, &tx.SweepStatus, &tx.CreatedAt, &tx.ConfirmedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scanning transaction: %w", err)
		}
		txs = append(txs, tx)
	}

	return txs, total, nil
}

// GetTransactionByID fetches a single transaction by ID for a specific user.
func (r *WalletRepo) GetTransactionByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.Transaction, error) {
	query := `
		SELECT id, user_id, asset_id, type, status, amount, fee, tx_hash, from_address, to_address, note, sweep_status, created_at, confirmed_at
		FROM transactions
		WHERE id = $1 AND user_id = $2
	`
	var tx models.Transaction
	err := r.pool.QueryRow(ctx, query, id, userID).Scan(
		&tx.ID, &tx.UserID, &tx.AssetID, &tx.Type, &tx.Status,
		&tx.Amount, &tx.Fee, &tx.TxHash, &tx.FromAddress, &tx.ToAddress,
		&tx.Note, &tx.SweepStatus, &tx.CreatedAt, &tx.ConfirmedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // not found
		}
		return nil, fmt.Errorf("querying transaction by id: %w", err)
	}
	return &tx, nil
}

// GetSwapByID fetches a single swap by ID for a specific user.
func (r *WalletRepo) GetSwapByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.SwapRecord, error) {
	query := `
		SELECT id, user_id, from_asset_id, to_asset_id, from_amount, to_amount, exchange_rate, fee_amount, created_at
		FROM swaps
		WHERE id = $1 AND user_id = $2
	`
	var swap models.SwapRecord
	err := r.pool.QueryRow(ctx, query, id, userID).Scan(
		&swap.ID, &swap.UserID, &swap.FromAssetID, &swap.ToAssetID,
		&swap.FromAmount, &swap.ToAmount, &swap.Rate, &swap.Fee, &swap.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("querying swap by id: %w", err)
	}
	return &swap, nil
}

// CreateTransaction inserts a new transaction record into the database.
func (r *WalletRepo) CreateTransaction(ctx context.Context, tx *models.Transaction) error {
	if tx.SweepStatus == "" {
		tx.SweepStatus = models.SweepNotNeeded
	}

	query := `
		INSERT INTO transactions (id, user_id, asset_id, type, status, amount, fee, tx_hash, from_address, to_address, note, sweep_status, created_at, confirmed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`
	_, err := r.pool.Exec(ctx, query,
		tx.ID, tx.UserID, tx.AssetID, tx.Type, tx.Status,
		tx.Amount, tx.Fee, tx.TxHash, tx.FromAddress, tx.ToAddress,
		tx.Note, tx.SweepStatus, tx.CreatedAt, tx.ConfirmedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting transaction: %w", err)
	}
	return nil
}

// UpdateTransactionStatus updates the status and tx_hash of a transaction.
func (r *WalletRepo) UpdateTransactionStatus(ctx context.Context, txID uuid.UUID, status models.TxStatus, txHash string) error {
	query := `
		UPDATE transactions
		SET status = $1, tx_hash = $2, confirmed_at = NOW()
		WHERE id = $3
	`
	_, err := r.pool.Exec(ctx, query, status, txHash, txID)
	if err != nil {
		return fmt.Errorf("updating transaction status: %w", err)
	}
	return nil
}

// GetPendingSweeps fetches transactions that need sweeping.
func (r *WalletRepo) GetPendingSweeps(ctx context.Context) ([]models.Transaction, error) {
	query := `
		SELECT id, user_id, asset_id, type, status, amount, fee, tx_hash, from_address, to_address, note, sweep_status, created_at, confirmed_at
		FROM transactions
		WHERE sweep_status IN ($1, $2)
		  AND status = $3
	`
	rows, err := r.pool.Query(ctx, query, models.SweepPendingSweep, models.SweepPendingGas, models.TxConfirmed)
	if err != nil {
		return nil, fmt.Errorf("querying pending sweeps: %w", err)
	}
	defer rows.Close()

	var txs []models.Transaction
	for rows.Next() {
		var tx models.Transaction
		if err := rows.Scan(
			&tx.ID, &tx.UserID, &tx.AssetID, &tx.Type, &tx.Status,
			&tx.Amount, &tx.Fee, &tx.TxHash, &tx.FromAddress, &tx.ToAddress,
			&tx.Note, &tx.SweepStatus, &tx.CreatedAt, &tx.ConfirmedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning pending sweep: %w", err)
		}
		txs = append(txs, tx)
	}
	return txs, nil
}

// UpdateSweepStatus updates the sweep_status of a transaction.
func (r *WalletRepo) UpdateSweepStatus(ctx context.Context, txID uuid.UUID, sweepStatus models.SweepStatus) error {
	query := `
		UPDATE transactions
		SET sweep_status = $1
		WHERE id = $2
	`
	_, err := r.pool.Exec(ctx, query, sweepStatus, txID)
	if err != nil {
		return fmt.Errorf("updating sweep status: %w", err)
	}
	return nil
}

// DeductBalance decreases a user's available balance for a specific asset.
func (r *WalletRepo) DeductBalance(ctx context.Context, userID uuid.UUID, assetID int, amount string) error {
	query := `
		INSERT INTO balances (user_id, asset_id, available, locked)
		VALUES ($1, $2, 0, 0)
		ON CONFLICT (user_id, asset_id) DO UPDATE
		SET available = GREATEST(0, balances.available - $3::numeric)
	`
	_, err := r.pool.Exec(ctx, query, userID, assetID, amount)
	if err != nil {
		return fmt.Errorf("deducting balance: %w", err)
	}
	return nil
}

// AddBalance increases a user's available balance for a specific asset.
func (r *WalletRepo) AddBalance(ctx context.Context, userID uuid.UUID, assetID int, amount string) error {
	query := `
		INSERT INTO balances (user_id, asset_id, available, locked)
		VALUES ($1, $2, $3, 0)
		ON CONFLICT (user_id, asset_id) DO UPDATE
		SET available = balances.available + $3::numeric
	`
	_, err := r.pool.Exec(ctx, query, userID, assetID, amount)
	if err != nil {
		return fmt.Errorf("adding balance: %w", err)
	}
	return nil
}

// GetStuckWithdrawals returns pending withdrawal transactions that were
// created more than olderThan ago and are still in 'pending' status.
// These are candidates for on-chain status checks and potential re-broadcast.
func (r *WalletRepo) GetStuckWithdrawals(ctx context.Context, olderThan time.Duration, youngerThan time.Duration) ([]models.Transaction, error) {
	cutoffOlder := time.Now().Add(-olderThan)
	cutoffYounger := time.Now().Add(-youngerThan)
	query := `
		SELECT id, user_id, asset_id, type, status, amount, fee, tx_hash, from_address, to_address, note, sweep_status, created_at, confirmed_at
		FROM transactions
		WHERE type = 'withdrawal' AND status = 'pending' AND created_at < $1 AND created_at > $2
	`
	rows, err := r.pool.Query(ctx, query, cutoffOlder, cutoffYounger)
	if err != nil {
		return nil, fmt.Errorf("querying stuck withdrawals: %w", err)
	}
	defer rows.Close()

	var txs []models.Transaction
	for rows.Next() {
		var tx models.Transaction
		if err := rows.Scan(
			&tx.ID, &tx.UserID, &tx.AssetID, &tx.Type, &tx.Status,
			&tx.Amount, &tx.Fee, &tx.TxHash, &tx.FromAddress, &tx.ToAddress,
			&tx.Note, &tx.SweepStatus, &tx.CreatedAt, &tx.ConfirmedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning stuck withdrawal: %w", err)
		}
		txs = append(txs, tx)
	}
	return txs, nil
}

// UpdateWithdrawalTxHash updates the tx_hash of a withdrawal after it has been
// re-broadcast with a higher gas price. The status remains 'pending' until confirmed.
func (r *WalletRepo) UpdateWithdrawalTxHash(ctx context.Context, txID uuid.UUID, newHash string) error {
	query := `
		UPDATE transactions
		SET tx_hash = $1
		WHERE id = $2
	`
	_, err := r.pool.Exec(ctx, query, newHash, txID)
	if err != nil {
		return fmt.Errorf("updating withdrawal tx hash: %w", err)
	}
	return nil
}

// ExecuteSwap atomically performs an internal swap between two assets for a user.
// It uses SELECT ... FOR UPDATE to prevent race conditions on balances.
func (r *WalletRepo) ExecuteSwap(ctx context.Context, swap *models.Swap) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning swap tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Lock and fetch FROM balance
	var fromBalStr string
	err = tx.QueryRow(ctx, `SELECT available FROM balances WHERE user_id = $1 AND asset_id = $2 FOR UPDATE`, swap.UserID, swap.FromAssetID).Scan(&fromBalStr)
	if err != nil {
		return fmt.Errorf("locking from_asset balance: %w", err)
	}

	// 2. Lock and fetch TO balance (or create if doesn't exist)
	var toBalStr string
	err = tx.QueryRow(ctx, `SELECT available FROM balances WHERE user_id = $1 AND asset_id = $2 FOR UPDATE`, swap.UserID, swap.ToAssetID).Scan(&toBalStr)
	if err != nil {
		if err == pgx.ErrNoRows {
			// Initialize if doesn't exist
			_, err = tx.Exec(ctx, `INSERT INTO balances (user_id, asset_id, available) VALUES ($1, $2, '0')`, swap.UserID, swap.ToAssetID)
			if err != nil {
				return fmt.Errorf("initializing to_asset balance: %w", err)
			}
			toBalStr = "0"
		} else {
			return fmt.Errorf("locking to_asset balance: %w", err)
		}
	}

	// 3. Deduct from FROM balance
	_, err = tx.Exec(ctx, `UPDATE balances SET available = available - $1::numeric WHERE user_id = $2 AND asset_id = $3`, swap.FromAmount, swap.UserID, swap.FromAssetID)
	if err != nil {
		return fmt.Errorf("deducting from_asset: %w", err)
	}

	// 4. Add to TO balance
	_, err = tx.Exec(ctx, `UPDATE balances SET available = available + $1::numeric WHERE user_id = $2 AND asset_id = $3`, swap.ToAmount, swap.UserID, swap.ToAssetID)
	if err != nil {
		return fmt.Errorf("crediting to_asset: %w", err)
	}

	// 5. Insert swap record
	swap.ID = uuid.New()
	_, err = tx.Exec(ctx, `
		INSERT INTO swaps (id, user_id, from_asset_id, to_asset_id, from_amount, to_amount, exchange_rate, fee_amount)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, swap.ID, swap.UserID, swap.FromAssetID, swap.ToAssetID, swap.FromAmount, swap.ToAmount, swap.ExchangeRate, swap.FeeAmount)
	if err != nil {
		return fmt.Errorf("inserting swap record: %w", err)
	}

	// 5a. Insert debit transaction
	now := time.Now()
	_, err = tx.Exec(ctx, `
		INSERT INTO transactions (id, user_id, asset_id, type, status, amount, fee, tx_hash, from_address, to_address, note, sweep_status, created_at, confirmed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`, uuid.New(), swap.UserID, swap.FromAssetID, models.TxSwap, models.TxConfirmed, swap.FromAmount, swap.FeeAmount, nil, nil, nil, nil, models.SweepNotNeeded, now, now)
	if err != nil {
		return fmt.Errorf("inserting swap debit transaction: %w", err)
	}

	// 5b. Insert credit transaction
	_, err = tx.Exec(ctx, `
		INSERT INTO transactions (id, user_id, asset_id, type, status, amount, fee, tx_hash, from_address, to_address, note, sweep_status, created_at, confirmed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`, uuid.New(), swap.UserID, swap.ToAssetID, models.TxSwap, models.TxConfirmed, swap.ToAmount, "0", nil, nil, nil, nil, models.SweepNotNeeded, now, now)
	if err != nil {
		return fmt.Errorf("inserting swap credit transaction: %w", err)
	}

	// 6. Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing swap tx: %w", err)
	}

	return nil
}

// GetWatchlist returns the list of asset IDs the user is watching.
func (r *WalletRepo) GetWatchlist(ctx context.Context, userID uuid.UUID) ([]int, error) {
	query := `SELECT asset_id FROM watchlists WHERE user_id = $1`
	
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assetIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		assetIDs = append(assetIDs, id)
	}

	return assetIDs, nil
}

// ToggleWatchlist adds an asset to the watchlist if missing, or removes it if present.
func (r *WalletRepo) ToggleWatchlist(ctx context.Context, userID uuid.UUID, assetID int) error {
	// We'll use a transaction just to be clean, though it's not strictly necessary.
	// Actually, an INSERT ... ON CONFLICT ... could work, but we want to toggle (remove if present).
	
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var exists bool
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM watchlists WHERE user_id = $1 AND asset_id = $2)`, userID, assetID).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		_, err = tx.Exec(ctx, `DELETE FROM watchlists WHERE user_id = $1 AND asset_id = $2`, userID, assetID)
	} else {
		_, err = tx.Exec(ctx, `INSERT INTO watchlists (user_id, asset_id) VALUES ($1, $2)`, userID, assetID)
	}

	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// GetAllTransactions fetches paginated transactions for all users (for admin panel).
func (r *WalletRepo) GetAllTransactions(ctx context.Context, limit, offset int) ([]models.Transaction, int, error) {
	countQuery := `SELECT COUNT(*) FROM transactions`
	
	var total int
	if err := r.pool.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting transactions: %w", err)
	}

	query := `
		SELECT id, user_id, asset_id, type, status, amount, fee, tx_hash, from_address, to_address, note, sweep_status, created_at, confirmed_at
		FROM transactions
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	
	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("querying all transactions: %w", err)
	}
	defer rows.Close()

	var txs []models.Transaction
	for rows.Next() {
		var tx models.Transaction
		if err := rows.Scan(
			&tx.ID, &tx.UserID, &tx.AssetID, &tx.Type, &tx.Status, &tx.Amount, &tx.Fee,
			&tx.TxHash, &tx.FromAddress, &tx.ToAddress, &tx.Note, &tx.SweepStatus,
			&tx.CreatedAt, &tx.ConfirmedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scanning transaction: %w", err)
		}
		txs = append(txs, tx)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return txs, total, nil
}
