package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"exchange/internal/db"
	"exchange/internal/models"
)

type P2PRepo struct {
	pool *db.Pool
}

func NewP2PRepo(pool *db.Pool) *P2PRepo {
	return &P2PRepo{pool: pool}
}

// CreateOrder creates a new P2P sell order and locks the seller's asset balance in escrow.
func (r *P2PRepo) CreateOrder(ctx context.Context, order *models.P2POrder) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Lock the balance
	// We deduct from available and add to locked.
	lockQuery := `
		UPDATE balances 
		SET available = available - $1::NUMERIC, 
		    locked = locked + $1::NUMERIC
		WHERE user_id = $2 AND asset_id = $3 AND available >= $1::NUMERIC
	`
	cmdTag, err := tx.Exec(ctx, lockQuery, order.AvailableAmount, order.SellerID, order.AssetID)
	if err != nil {
		return fmt.Errorf("locking balance: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("insufficient available balance")
	}

	// 2. Insert the order
	insertQuery := `
		INSERT INTO p2p_orders (id, seller_id, asset_id, fiat_currency, rate, min_amount, max_amount, available_amount, payment_method, status, completion_rate, total_orders, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err = tx.Exec(ctx, insertQuery,
		order.ID, order.SellerID, order.AssetID, order.FiatCurrency, order.Rate,
		order.MinAmount, order.MaxAmount, order.AvailableAmount, order.PaymentMethod,
		order.Status, order.CompletionRate, order.TotalOrders, order.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting order: %w", err)
	}

	return tx.Commit(ctx)
}

// ListActiveOrders returns all active P2P orders.
func (r *P2PRepo) ListActiveOrders(ctx context.Context) ([]*models.P2POrder, error) {
	query := `
		SELECT id, seller_id, asset_id, fiat_currency, rate, min_amount, max_amount, available_amount, payment_method, status, completion_rate, total_orders, created_at
		FROM p2p_orders
		WHERE status = 'active' AND available_amount > 0
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("querying active orders: %w", err)
	}
	defer rows.Close()

	orders := make([]*models.P2POrder, 0)
	for rows.Next() {
		o := &models.P2POrder{}
		if err := rows.Scan(
			&o.ID, &o.SellerID, &o.AssetID, &o.FiatCurrency, &o.Rate,
			&o.MinAmount, &o.MaxAmount, &o.AvailableAmount, &o.PaymentMethod,
			&o.Status, &o.CompletionRate, &o.TotalOrders, &o.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning order: %w", err)
		}
		orders = append(orders, o)
	}
	return orders, nil
}

// GetOrder fetches a specific order by ID.
func (r *P2PRepo) GetOrder(ctx context.Context, orderID uuid.UUID) (*models.P2POrder, error) {
	query := `
		SELECT id, seller_id, asset_id, fiat_currency, rate, min_amount, max_amount, available_amount, payment_method, status, completion_rate, total_orders, created_at
		FROM p2p_orders
		WHERE id = $1
	`
	o := &models.P2POrder{}
	err := r.pool.QueryRow(ctx, query, orderID).Scan(
		&o.ID, &o.SellerID, &o.AssetID, &o.FiatCurrency, &o.Rate,
		&o.MinAmount, &o.MaxAmount, &o.AvailableAmount, &o.PaymentMethod,
		&o.Status, &o.CompletionRate, &o.TotalOrders, &o.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("querying order: %w", err)
	}
	return o, nil
}

// CreateTrade initiates a trade and reserves the crypto from the order's available amount.
func (r *P2PRepo) CreateTrade(ctx context.Context, trade *models.P2PTrade) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Deduct from order's available amount
	updateOrderQuery := `
		UPDATE p2p_orders
		SET available_amount = available_amount - $1::NUMERIC
		WHERE id = $2 AND available_amount >= $1::NUMERIC AND status = 'active'
	`
	cmdTag, err := tx.Exec(ctx, updateOrderQuery, trade.Amount, trade.OrderID)
	if err != nil {
		return fmt.Errorf("updating order available amount: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("insufficient order available amount or order not active")
	}

	// 2. Insert the trade
	insertTradeQuery := `
		INSERT INTO p2p_trades (id, order_id, buyer_id, seller_id, asset_id, amount, fiat_amount, rate, status, escrow_locked, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err = tx.Exec(ctx, insertTradeQuery,
		trade.ID, trade.OrderID, trade.BuyerID, trade.SellerID, trade.AssetID,
		trade.Amount, trade.FiatAmount, trade.Rate, trade.Status, trade.EscrowLocked, trade.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting trade: %w", err)
	}

	return tx.Commit(ctx)
}

// GetTrade fetches a trade by ID.
func (r *P2PRepo) GetTrade(ctx context.Context, tradeID uuid.UUID) (*models.P2PTrade, error) {
	query := `
		SELECT id, order_id, buyer_id, seller_id, asset_id, amount, fiat_amount, rate, status, escrow_locked, created_at, paid_at, released_at
		FROM p2p_trades
		WHERE id = $1
	`
	t := &models.P2PTrade{}
	err := r.pool.QueryRow(ctx, query, tradeID).Scan(
		&t.ID, &t.OrderID, &t.BuyerID, &t.SellerID, &t.AssetID,
		&t.Amount, &t.FiatAmount, &t.Rate, &t.Status, &t.EscrowLocked,
		&t.CreatedAt, &t.PaidAt, &t.ReleasedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("querying trade: %w", err)
	}
	return t, nil
}

// ListUserTrades lists all trades for a given user (either as buyer or seller).
func (r *P2PRepo) ListUserTrades(ctx context.Context, userID uuid.UUID) ([]*models.P2PTrade, error) {
	query := `
		SELECT id, order_id, buyer_id, seller_id, asset_id, amount, fiat_amount, rate, status, escrow_locked, created_at, paid_at, released_at
		FROM p2p_trades
		WHERE buyer_id = $1 OR seller_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("querying user trades: %w", err)
	}
	defer rows.Close()

	trades := make([]*models.P2PTrade, 0)
	for rows.Next() {
		t := &models.P2PTrade{}
		if err := rows.Scan(
			&t.ID, &t.OrderID, &t.BuyerID, &t.SellerID, &t.AssetID,
			&t.Amount, &t.FiatAmount, &t.Rate, &t.Status, &t.EscrowLocked,
			&t.CreatedAt, &t.PaidAt, &t.ReleasedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning trade: %w", err)
		}
		trades = append(trades, t)
	}
	return trades, nil
}

// MarkTradePaid updates trade status to paid.
func (r *P2PRepo) MarkTradePaid(ctx context.Context, tradeID uuid.UUID) error {
	query := `
		UPDATE p2p_trades
		SET status = 'paid', paid_at = NOW()
		WHERE id = $1 AND status = 'waiting_payment'
	`
	cmdTag, err := r.pool.Exec(ctx, query, tradeID)
	if err != nil {
		return fmt.Errorf("marking trade paid: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("trade not found or not in waiting_payment status")
	}
	return nil
}

// ReleaseTrade releases the escrowed crypto to the buyer.
func (r *P2PRepo) ReleaseTrade(ctx context.Context, tradeID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Get trade info and lock the row
	var sellerID, buyerID uuid.UUID
	var assetID int
	var amount string
	
	err = tx.QueryRow(ctx, `
		SELECT seller_id, buyer_id, asset_id, amount
		FROM p2p_trades
		WHERE id = $1 AND status = 'paid' AND escrow_locked = true
		FOR UPDATE
	`, tradeID).Scan(&sellerID, &buyerID, &assetID, &amount)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("trade not found or cannot be released")
		}
		return fmt.Errorf("fetching trade for release: %w", err)
	}

	// 2. Deduct from seller's locked balance
	_, err = tx.Exec(ctx, `
		UPDATE balances
		SET locked = locked - $1::NUMERIC
		WHERE user_id = $2 AND asset_id = $3
	`, amount, sellerID, assetID)
	if err != nil {
		return fmt.Errorf("deducting seller locked balance: %w", err)
	}

	// 3. Add to buyer's available balance
	// Insert or update
	_, err = tx.Exec(ctx, `
		INSERT INTO balances (user_id, asset_id, available, locked)
		VALUES ($1, $2, $3::NUMERIC, 0)
		ON CONFLICT (user_id, asset_id) 
		DO UPDATE SET available = balances.available + EXCLUDED.available
	`, buyerID, assetID, amount)
	if err != nil {
		return fmt.Errorf("crediting buyer available balance: %w", err)
	}

	// 4. Mark trade as released
	_, err = tx.Exec(ctx, `
		UPDATE p2p_trades
		SET status = 'released', escrow_locked = false, released_at = NOW()
		WHERE id = $1
	`, tradeID)
	if err != nil {
		return fmt.Errorf("updating trade status: %w", err)
	}

	// 5. Update seller's total_orders and completion rate (naive increment for now)
	_, err = tx.Exec(ctx, `
		UPDATE p2p_orders
		SET total_orders = total_orders + 1
		WHERE id = (SELECT order_id FROM p2p_trades WHERE id = $1)
	`, tradeID)
	if err != nil {
		return fmt.Errorf("updating order stats: %w", err)
	}

	return tx.Commit(ctx)
}

// CancelTrade cancels a trade and returns the locked amount back to the order's available amount.
func (r *P2PRepo) CancelTrade(ctx context.Context, tradeID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var orderID uuid.UUID
	var amount string
	err = tx.QueryRow(ctx, `
		SELECT order_id, amount
		FROM p2p_trades
		WHERE id = $1 AND status IN ('waiting_payment', 'paid') AND escrow_locked = true
		FOR UPDATE
	`, tradeID).Scan(&orderID, &amount)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("trade cannot be cancelled in current state")
		}
		return fmt.Errorf("fetching trade for cancellation: %w", err)
	}

	// 1. Update trade status
	_, err = tx.Exec(ctx, `
		UPDATE p2p_trades
		SET status = 'cancelled'
		WHERE id = $1
	`, tradeID)
	if err != nil {
		return fmt.Errorf("updating trade status to cancelled: %w", err)
	}

	// 2. Return the amount to the order's available pool
	_, err = tx.Exec(ctx, `
		UPDATE p2p_orders
		SET available_amount = available_amount + $1::NUMERIC
		WHERE id = $2
	`, amount, orderID)
	if err != nil {
		return fmt.Errorf("returning amount to order: %w", err)
	}

	return tx.Commit(ctx)
}
