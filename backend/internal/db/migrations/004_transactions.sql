-- 004_transactions.sql
-- The canonical ledger for all value movements:
--   deposits      — on-chain inbound transfer, credited after N confirmations
--   withdrawals   — user-initiated on-chain send
--   swaps         — internal ledger swap, no on-chain tx (tx_hash IS NULL)
--   p2p_buy/sell  — completed P2P trades

CREATE TABLE IF NOT EXISTS transactions (
    id           UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID           NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    asset_id     INT            NOT NULL REFERENCES assets(id),
    type         TEXT           NOT NULL
                                CHECK (type IN ('deposit','withdrawal','swap','p2p_buy','p2p_sell')),
    status       TEXT           NOT NULL DEFAULT 'pending'
                                CHECK (status IN ('pending','confirmed','failed')),
    amount       NUMERIC(36,18) NOT NULL CHECK (amount > 0),
    fee          NUMERIC(36,18) NOT NULL DEFAULT 0 CHECK (fee >= 0),
    tx_hash      TEXT,           -- NULL for internal swaps
    from_address TEXT,
    to_address   TEXT,
    note         TEXT,
    created_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    confirmed_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_transactions_user       ON transactions(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_transactions_tx_hash    ON transactions(tx_hash) WHERE tx_hash IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_transactions_status     ON transactions(status) WHERE status = 'pending';

-- ─────────────────────────────────────────────────────────────────────────────

-- swap_records links the two transaction rows (debit + credit) produced by one
-- internal swap, giving a clean view of From/To for display purposes.
CREATE TABLE IF NOT EXISTS swap_records (
    id            UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID           NOT NULL REFERENCES users(id),
    from_asset_id INT            NOT NULL REFERENCES assets(id),
    to_asset_id   INT            NOT NULL REFERENCES assets(id),
    from_amount   NUMERIC(36,18) NOT NULL,
    to_amount     NUMERIC(36,18) NOT NULL,
    rate          NUMERIC(24,12) NOT NULL,  -- from_price / to_price at execution time
    fee           NUMERIC(36,18) NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);
