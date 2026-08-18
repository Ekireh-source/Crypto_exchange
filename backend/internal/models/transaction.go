package models

import (
	"time"

	"github.com/google/uuid"
)

// TxType classifies a transaction record.
type TxType string

const (
	TxDeposit    TxType = "deposit"
	TxWithdrawal TxType = "withdrawal"
	TxSwap       TxType = "swap"
	TxP2PBuy     TxType = "p2p_buy"
	TxP2PSell    TxType = "p2p_sell"
)

// TxStatus describes the lifecycle state of a transaction.
type TxStatus string

const (
	TxPending   TxStatus = "pending"
	TxConfirmed TxStatus = "confirmed"
	TxFailed    TxStatus = "failed"
)

// Transaction is the canonical ledger record for any value movement.
// On-chain deposits/withdrawals have a TxHash; internal swaps do not.
type Transaction struct {
	ID          uuid.UUID  `db:"id"           json:"id"`
	UserID      uuid.UUID  `db:"user_id"      json:"user_id"`
	AssetID     int        `db:"asset_id"     json:"asset_id"`
	Type        TxType     `db:"type"         json:"type"`
	Status      TxStatus   `db:"status"       json:"status"`
	Amount      string     `db:"amount"       json:"amount"`      // NUMERIC as string
	Fee         string     `db:"fee"          json:"fee"`         // NUMERIC as string
	TxHash      *string    `db:"tx_hash"      json:"tx_hash"`     // nil for internal swaps
	FromAddress *string    `db:"from_address" json:"from_address"`
	ToAddress   *string    `db:"to_address"   json:"to_address"`
	Note        *string    `db:"note"         json:"note"`
	CreatedAt   time.Time  `db:"created_at"   json:"created_at"`
	ConfirmedAt *time.Time `db:"confirmed_at" json:"confirmed_at"`
}

// SwapRecord links two transaction rows (debit + credit) for a swap operation.
type SwapRecord struct {
	ID          uuid.UUID `db:"id"           json:"id"`
	UserID      uuid.UUID `db:"user_id"      json:"user_id"`
	FromAssetID int       `db:"from_asset_id" json:"from_asset_id"`
	ToAssetID   int       `db:"to_asset_id"  json:"to_asset_id"`
	FromAmount  string    `db:"from_amount"  json:"from_amount"`
	ToAmount    string    `db:"to_amount"    json:"to_amount"`
	Rate        float64   `db:"rate"         json:"rate"`
	Fee         string    `db:"fee"          json:"fee"`
	CreatedAt   time.Time `db:"created_at"   json:"created_at"`
}
