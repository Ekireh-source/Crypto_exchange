package models

import (
	"time"

	"github.com/google/uuid"
)

// P2POrderStatus represents the lifecycle of a sell listing.
type P2POrderStatus string

const (
	P2POrderActive    P2POrderStatus = "active"
	P2POrderPaused    P2POrderStatus = "paused"
	P2POrderCompleted P2POrderStatus = "completed"
	P2POrderCancelled P2POrderStatus = "cancelled"
)

// P2POrder is a sell listing created by a seller.
// The seller's crypto is locked in escrow (balances.locked) for the duration.
type P2POrder struct {
	ID              uuid.UUID      `db:"id"               json:"id"`
	SellerID        uuid.UUID      `db:"seller_id"        json:"seller_id"`
	AssetID         int            `db:"asset_id"         json:"asset_id"`
	FiatCurrency    string         `db:"fiat_currency"    json:"fiat_currency"`  // e.g. "USD"
	Rate            string         `db:"rate"             json:"rate"`           // price per coin in fiat
	MinAmount       string         `db:"min_amount"       json:"min_amount"`
	MaxAmount       string         `db:"max_amount"       json:"max_amount"`
	AvailableAmount string         `db:"available_amount" json:"available_amount"`
	PaymentMethod   string         `db:"payment_method"   json:"payment_method"` // e.g. "Bank Transfer"
	Status          P2POrderStatus `db:"status"           json:"status"`
	CompletionRate  float64        `db:"completion_rate"  json:"completion_rate"` // 0–100
	TotalOrders     int            `db:"total_orders"     json:"total_orders"`
	CreatedAt       time.Time      `db:"created_at"       json:"created_at"`
}

// P2PTradeStatus tracks each individual trade initiated from a P2POrder.
type P2PTradeStatus string

const (
	TradeWaitingPayment P2PTradeStatus = "waiting_payment"
	TradePaid           P2PTradeStatus = "paid"
	TradeReleased       P2PTradeStatus = "released"
	TradeDisputed       P2PTradeStatus = "disputed"
	TradeCancelled      P2PTradeStatus = "cancelled"
)

// P2PTrade represents one buyer-seller trade against a P2POrder.
type P2PTrade struct {
	ID           uuid.UUID      `db:"id"            json:"id"`
	OrderID      uuid.UUID      `db:"order_id"      json:"order_id"`
	BuyerID      uuid.UUID      `db:"buyer_id"      json:"buyer_id"`
	SellerID     uuid.UUID      `db:"seller_id"     json:"seller_id"`
	AssetID      int            `db:"asset_id"      json:"asset_id"`
	Amount       string         `db:"amount"        json:"amount"`       // crypto amount
	FiatAmount   string         `db:"fiat_amount"   json:"fiat_amount"`  // buyer pays this in fiat
	Rate         string         `db:"rate"          json:"rate"`
	Status       P2PTradeStatus `db:"status"        json:"status"`
	EscrowLocked bool           `db:"escrow_locked" json:"escrow_locked"`
	CreatedAt    time.Time      `db:"created_at"    json:"created_at"`
	PaidAt       *time.Time     `db:"paid_at"       json:"paid_at"`
	ReleasedAt   *time.Time     `db:"released_at"   json:"released_at"`
}
