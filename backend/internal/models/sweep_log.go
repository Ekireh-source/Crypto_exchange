package models

import (
	"time"

	"github.com/google/uuid"
)

type SweepLogType string

const (
	SweepTypeGasFunding SweepLogType = "GAS_FUNDING"
	SweepTypeToken      SweepLogType = "TOKEN_SWEEP"
	SweepTypeNative     SweepLogType = "NATIVE_SWEEP"
)

type SweepLogStatus string

const (
	SweepLogStatusPending SweepLogStatus = "PENDING"
	SweepLogStatusSuccess SweepLogStatus = "SUCCESS"
	SweepLogStatusFailed  SweepLogStatus = "FAILED"
)

type SweepLog struct {
	ID            int            `db:"id"             json:"id"`
	TransactionID uuid.UUID      `db:"transaction_id" json:"transaction_id"`
	UserID        uuid.UUID      `db:"user_id"        json:"user_id"`
	AssetID       int            `db:"asset_id"       json:"asset_id"`
	Type          SweepLogType   `db:"type"           json:"type"`
	Status        SweepLogStatus `db:"status"         json:"status"`
	Amount        string         `db:"amount"         json:"amount"`
	TxHash        *string        `db:"tx_hash"        json:"tx_hash"`
	FromAddress   string         `db:"from_address"   json:"from_address"`
	ToAddress     string         `db:"to_address"     json:"to_address"`
	ErrorMessage  *string        `db:"error_message"  json:"error_message"`
	CreatedAt     time.Time      `db:"created_at"     json:"created_at"`
	UpdatedAt     time.Time      `db:"updated_at"     json:"updated_at"`
}
