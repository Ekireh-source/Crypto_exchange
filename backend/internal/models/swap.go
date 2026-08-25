package models

import (
	"time"

	"github.com/google/uuid"
)

// Swap represents a completed asset swap operation.
type Swap struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	FromAssetID  int       `json:"from_asset_id"`
	ToAssetID    int       `json:"to_asset_id"`
	FromAmount   string    `json:"from_amount"`
	ToAmount     string    `json:"to_amount"`
	ExchangeRate string    `json:"exchange_rate"`
	FeeAmount    string    `json:"fee_amount"`
	CreatedAt    time.Time `json:"created_at"`
}

// SwapRequest represents the incoming API payload for a swap.
type SwapRequest struct {
	FromAssetID int    `json:"from_asset_id"`
	ToAssetID   int    `json:"to_asset_id"`
	Amount      string `json:"amount"` // The amount of from_asset to swap
}

// SwapResponse is the API response after a successful swap.
type SwapResponse struct {
	Swap Swap `json:"swap"`
}
