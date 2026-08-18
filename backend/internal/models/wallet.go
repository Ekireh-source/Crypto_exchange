package models

import (
	"time"

	"github.com/google/uuid"
)

// DepositAddress holds the on-chain address generated per user per asset.
// The private key is stored AES-256 encrypted; it is decrypted only when
// signing a withdrawal transaction.
type DepositAddress struct {
	ID                 uuid.UUID `db:"id"                  json:"id"`
	UserID             uuid.UUID `db:"user_id"             json:"-"`
	AssetID            int       `db:"asset_id"            json:"asset_id"`
	Address            string    `db:"address"             json:"address"`
	EncryptedPrivateKey string   `db:"encrypted_private_key" json:"-"` // never exposed via API
	CreatedAt          time.Time `db:"created_at"          json:"created_at"`
}

// Balance tracks a user's available and locked amount for a single asset.
// locked is used for P2P escrow and open limit orders.
type Balance struct {
	UserID    uuid.UUID `db:"user_id"  json:"user_id"`
	AssetID   int       `db:"asset_id" json:"asset_id"`
	Available string    `db:"available" json:"available"` // NUMERIC stored as string to avoid float precision loss
	Locked    string    `db:"locked"    json:"locked"`
}

// AssetBalance combines asset metadata with the user's live balance —
// this is what the /wallet endpoint returns.
type AssetBalance struct {
	Asset      Asset   `json:"asset"`
	Available  string  `json:"available"`
	Locked     string  `json:"locked"`
	USDValue   float64 `json:"usd_value"`
}

// DepositAddressResponse is the safe, API-facing shape for a deposit address.
type DepositAddressResponse struct {
	AssetID   int     `json:"asset_id"`
	Symbol    string  `json:"symbol"`
	Network   Network `json:"network"`
	Standard  string  `json:"standard"`
	Address   string  `json:"address"`
	// QRData is a URI the frontend can feed directly into a QR generator:
	// e.g. "ethereum:0xABC..." or just the plain address.
	QRData    string  `json:"qr_data"`
	// Warning to show the user so they don't send the wrong network.
	NetworkWarning string `json:"network_warning"`
}
