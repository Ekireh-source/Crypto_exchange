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
