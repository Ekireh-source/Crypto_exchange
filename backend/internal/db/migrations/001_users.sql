-- 001_users.sql
-- Users table: one row per exchange account.
-- referral_code is a short unique code shown to the user (e.g. "7252").
-- referred_by points to the user who introduced them (via referral link).

CREATE EXTENSION IF NOT EXISTS "pgcrypto";  -- for gen_random_uuid()

CREATE TABLE IF NOT EXISTS users (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    email         TEXT        UNIQUE NOT NULL,
    phone         TEXT,
    password_hash TEXT        NOT NULL,
    referral_code TEXT        UNIQUE NOT NULL,
    referred_by   UUID        REFERENCES users(id) ON DELETE SET NULL,
    kyc_status    TEXT        NOT NULL DEFAULT 'pending'
                              CHECK (kyc_status IN ('pending','verified','rejected')),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Keep updated_at in sync automatically.
CREATE OR REPLACE FUNCTION touch_updated_at()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS users_updated_at ON users;
CREATE TRIGGER users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

-- Fast lookup for auth (login by email) and referral link resolution.
CREATE INDEX IF NOT EXISTS idx_users_email         ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_referral_code ON users(referral_code);
CREATE INDEX IF NOT EXISTS idx_users_referred_by   ON users(referred_by);
