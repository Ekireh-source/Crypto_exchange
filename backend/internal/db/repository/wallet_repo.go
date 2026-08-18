package repository

import (
	"context"
	"fmt"

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
		SELECT id, user_id, asset_id, type, status, amount, fee, tx_hash, from_address, to_address, note, created_at, confirmed_at
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
			&tx.Note, &tx.CreatedAt, &tx.ConfirmedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scanning transaction: %w", err)
		}
		txs = append(txs, tx)
	}

	return txs, total, nil
}

// CreateTransaction inserts a new transaction record into the database.
func (r *WalletRepo) CreateTransaction(ctx context.Context, tx *models.Transaction) error {
	query := `
		INSERT INTO transactions (id, user_id, asset_id, type, status, amount, fee, tx_hash, from_address, to_address, note, created_at, confirmed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err := r.pool.Exec(ctx, query,
		tx.ID, tx.UserID, tx.AssetID, tx.Type, tx.Status,
		tx.Amount, tx.Fee, tx.TxHash, tx.FromAddress, tx.ToAddress,
		tx.Note, tx.CreatedAt, tx.ConfirmedAt,
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
