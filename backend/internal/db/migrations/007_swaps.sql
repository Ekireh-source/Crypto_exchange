-- 007_swaps.sql

CREATE TABLE IF NOT EXISTS swaps (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    from_asset_id INTEGER NOT NULL REFERENCES assets(id),
    to_asset_id INTEGER NOT NULL REFERENCES assets(id),
    from_amount NUMERIC(36, 18) NOT NULL,
    to_amount NUMERIC(36, 18) NOT NULL,
    exchange_rate NUMERIC(36, 18) NOT NULL,
    fee_amount NUMERIC(36, 18) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for quick lookups by user (e.g. for user transaction history)
CREATE INDEX IF NOT EXISTS idx_swaps_user_id ON swaps(user_id);
