package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"exchange/internal/db"
	"exchange/internal/models"
)

// UserRepo provides database operations for Users.
type UserRepo struct {
	pool *db.Pool
}

// NewUserRepo creates a new UserRepo.
func NewUserRepo(pool *db.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

// Create inserts a new user into the database.
func (r *UserRepo) Create(ctx context.Context, u *models.User) error {
	query := `
		INSERT INTO users (id, email, phone, password_hash, referral_code, referred_by, kyc_status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.pool.Exec(ctx, query,
		u.ID, u.Email, u.Phone, u.PasswordHash, u.ReferralCode, u.ReferredBy, u.KYCStatus, u.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting user: %w", err)
	}
	return nil
}

// GetByEmail retrieves a user by their email address.
func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, email, phone, password_hash, referral_code, referred_by, kyc_status, created_at
		FROM users
		WHERE email = $1
	`
	u := &models.User{}
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&u.ID, &u.Email, &u.Phone, &u.PasswordHash, &u.ReferralCode, &u.ReferredBy, &u.KYCStatus, &u.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("getting user by email: %w", err)
	}
	return u, nil
}

// GetByID retrieves a user by their ID.
func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, email, phone, password_hash, referral_code, referred_by, kyc_status, created_at
		FROM users
		WHERE id = $1
	`
	u := &models.User{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.Email, &u.Phone, &u.PasswordHash, &u.ReferralCode, &u.ReferredBy, &u.KYCStatus, &u.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("getting user by ID: %w", err)
	}
	return u, nil
}

// GetByReferralCode retrieves a user by their referral code.
func (r *UserRepo) GetByReferralCode(ctx context.Context, code string) (*models.User, error) {
	query := `
		SELECT id, email, phone, password_hash, referral_code, referred_by, kyc_status, created_at
		FROM users
		WHERE referral_code = $1
	`
	u := &models.User{}
	err := r.pool.QueryRow(ctx, query, code).Scan(
		&u.ID, &u.Email, &u.Phone, &u.PasswordHash, &u.ReferralCode, &u.ReferredBy, &u.KYCStatus, &u.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("getting user by referral code: %w", err)
	}
	return u, nil
}
