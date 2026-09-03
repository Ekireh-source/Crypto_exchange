package models

import (
	"time"

	"github.com/google/uuid"
)

// KYCStatus represents the KYC verification state of a user.
type KYCStatus string

const (
	KYCPending  KYCStatus = "pending"
	KYCVerified KYCStatus = "verified"
	KYCRejected KYCStatus = "rejected"
)

// User represents an exchange account holder.
type User struct {
	ID           uuid.UUID  `db:"id"            json:"id"`
	Email        string     `db:"email"         json:"email"`
	Phone        *string    `db:"phone"         json:"phone,omitempty"`
	PasswordHash string     `db:"password_hash" json:"-"`
	ReferralCode string     `db:"referral_code" json:"referral_code"`
	ReferredBy   *uuid.UUID `db:"referred_by"   json:"referred_by,omitempty"`
	KYCStatus    KYCStatus  `db:"kyc_status"    json:"kyc_status"`
	IsEmailVerified bool       `db:"is_email_verified" json:"is_email_verified"`
	Role         string     `db:"role"          json:"role"`
	CreatedAt    time.Time  `db:"created_at"    json:"created_at"`
}

// EmailVerificationToken represents a token used for verifying a user's email address.
type EmailVerificationToken struct {
	ID        uuid.UUID `db:"id"         json:"id"`
	UserID    uuid.UUID `db:"user_id"    json:"user_id"`
	Token     string    `db:"token"      json:"token"`
	ExpiresAt time.Time `db:"expires_at" json:"expires_at"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// ReferralStats summarises a user's referral performance.
type ReferralStats struct {
	ReferralCode string  `json:"referral_code"`
	ReferralLink string  `json:"referral_link"`
	TeamCount    int     `json:"team_count"`
	FeeEarnings  float64 `json:"fee_earnings"`  // in USD
	TotalVolume  float64 `json:"total_volume"`  // referred users' volume in USD
}
