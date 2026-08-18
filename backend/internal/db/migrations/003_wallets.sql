-- 003_wallets.sql
-- Two tables:
--   deposit_addresses  — the on-chain address generated for each user per asset
--   balances           — internal ledger (available + locked amount per user/asset)
--
-- Why separate from the user table?
--   A user may have one deposit address per network (BSC address, TRON address).
--   The private key is AES-256 encrypted before storage and NEVER returned by
--   the API — it is only decrypted server-side to sign withdrawals.

CREATE TABLE IF NOT EXISTS deposit_addresses (
    id                    UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id               UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    asset_id              INT         NOT NULL REFERENCES assets(id),
    address               TEXT        NOT NULL,
    encrypted_private_key TEXT        NOT NULL,  -- AES-256-GCM, base64-encoded
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, asset_id)
);

CREATE INDEX IF NOT EXISTS idx_deposit_addresses_user    ON deposit_addresses(user_id);
CREATE INDEX IF NOT EXISTS idx_deposit_addresses_address ON deposit_addresses(address);

-- ─────────────────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS balances (
    user_id   UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    asset_id  INT          NOT NULL REFERENCES assets(id),
    available NUMERIC(36, 18) NOT NULL DEFAULT 0 CHECK (available >= 0),
    locked    NUMERIC(36, 18) NOT NULL DEFAULT 0 CHECK (locked >= 0),
    PRIMARY KEY (user_id, asset_id)
);

-- Ensure every new asset gets a zero-balance row for each user when inserted.
-- (Optional convenience — services can also upsert on first credit.)
