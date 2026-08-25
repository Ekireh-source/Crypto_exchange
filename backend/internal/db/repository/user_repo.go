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
		SELECT id, email, phone, password_hash, referral_code, referred_by, kyc_status, role, created_at
		FROM users WHERE email = $1
	`
	u := &models.User{}
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&u.ID, &u.Email, &u.Phone, &u.PasswordHash,
		&u.ReferralCode, &u.ReferredBy, &u.KYCStatus, &u.Role, &u.CreatedAt,
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
		SELECT id, email, phone, password_hash, referral_code, referred_by, kyc_status, role, created_at
		FROM users WHERE id = $1
	`
	u := &models.User{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.Email, &u.Phone, &u.PasswordHash,
		&u.ReferralCode, &u.ReferredBy, &u.KYCStatus, &u.Role, &u.CreatedAt,
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
		SELECT id, email, phone, password_hash, referral_code, referred_by, kyc_status, role, created_at
		FROM users WHERE referral_code = $1
	`
	u := &models.User{}
	err := r.pool.QueryRow(ctx, query, code).Scan(
		&u.ID, &u.Email, &u.Phone, &u.PasswordHash,
		&u.ReferralCode, &u.ReferredBy, &u.KYCStatus, &u.Role, &u.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("getting user by referral code: %w", err)
	}
	return u, nil
}

// GetReferralStats calculates basic referral stats for a user.
func (r *UserRepo) GetReferralStats(ctx context.Context, userID uuid.UUID, referralCode string) (*models.ReferralStats, error) {
	query := `
		SELECT count(*)
		FROM users
		WHERE referred_by = $1
	`
	var teamCount int
	err := r.pool.QueryRow(ctx, query, userID).Scan(&teamCount)
	if err != nil {
		return nil, fmt.Errorf("counting referrals: %w", err)
	}

	return &models.ReferralStats{
		ReferralCode: referralCode,
		ReferralLink: fmt.Sprintf("https://xxhange.com/register?ref=%s", referralCode),
		TeamCount:    teamCount,
		FeeEarnings:  0,
		TotalVolume:  0,
	}, nil
}

// GetReferredUsers returns a list of users referred by a specific user.
func (r *UserRepo) GetReferredUsers(ctx context.Context, userID uuid.UUID) ([]*models.User, error) {
	query := `
		SELECT id, email, phone, kyc_status, created_at
		FROM users
		WHERE referred_by = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("querying referred users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		u := &models.User{}
		// Scan only the fields we selected. Other fields will be zero/empty.
		if err := rows.Scan(&u.ID, &u.Email, &u.Phone, &u.KYCStatus, &u.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning referred user: %w", err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error for referred users: %w", err)
	}

	return users, nil
}

// UpdateRole updates a user's role.
func (r *UserRepo) UpdateRole(ctx context.Context, userID uuid.UUID, role string) error {
	query := `UPDATE users SET role = $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, role, userID)
	return err
}

// ListUsers returns a list of all users (for admin panel).
func (r *UserRepo) ListUsers(ctx context.Context) ([]*models.User, error) {
	query := `
		SELECT id, email, phone, referral_code, kyc_status, role, created_at
		FROM users
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("querying users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		u := &models.User{}
		if err := rows.Scan(&u.ID, &u.Email, &u.Phone, &u.ReferralCode, &u.KYCStatus, &u.Role, &u.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning user: %w", err)
		}
		users = append(users, u)
	}
	return users, nil
}

