-- 005_p2p_orders.sql
-- P2P marketplace tables.
-- p2p_orders  — sell listings created by sellers (crypto locked in escrow)
-- p2p_trades  — individual buyer–seller transactions against a listing

CREATE TABLE IF NOT EXISTS p2p_orders (
    id               UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    seller_id        UUID           NOT NULL REFERENCES users(id),
    asset_id         INT            NOT NULL REFERENCES assets(id),
    fiat_currency    TEXT           NOT NULL DEFAULT 'USD',
    rate             NUMERIC(18, 6) NOT NULL CHECK (rate > 0),  -- fiat per 1 coin unit
    min_amount       NUMERIC(36,18) NOT NULL CHECK (min_amount > 0),
    max_amount       NUMERIC(36,18) NOT NULL CHECK (max_amount >= min_amount),
    available_amount NUMERIC(36,18) NOT NULL CHECK (available_amount >= 0),
    payment_method   TEXT           NOT NULL DEFAULT 'Bank Transfer',
    status           TEXT           NOT NULL DEFAULT 'active'
                                    CHECK (status IN ('active','paused','completed','cancelled')),
    completion_rate  NUMERIC(5,2)   NOT NULL DEFAULT 0,
    total_orders     INT            NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_p2p_orders_asset_status ON p2p_orders(asset_id, status);
CREATE INDEX IF NOT EXISTS idx_p2p_orders_seller       ON p2p_orders(seller_id);

-- ─────────────────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS p2p_trades (
    id            UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id      UUID           NOT NULL REFERENCES p2p_orders(id),
    buyer_id      UUID           NOT NULL REFERENCES users(id),
    seller_id     UUID           NOT NULL REFERENCES users(id),
    asset_id      INT            NOT NULL REFERENCES assets(id),
    amount        NUMERIC(36,18) NOT NULL CHECK (amount > 0),
    fiat_amount   NUMERIC(18, 2) NOT NULL CHECK (fiat_amount > 0),
    rate          NUMERIC(18, 6) NOT NULL,
    status        TEXT           NOT NULL DEFAULT 'waiting_payment'
                                 CHECK (status IN (
                                     'waiting_payment','paid','released',
                                     'disputed','cancelled'
                                 )),
    escrow_locked BOOLEAN        NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    paid_at       TIMESTAMPTZ,
    released_at   TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_p2p_trades_buyer  ON p2p_trades(buyer_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_p2p_trades_seller ON p2p_trades(seller_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_p2p_trades_order  ON p2p_trades(order_id);
