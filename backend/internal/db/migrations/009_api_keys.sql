-- 009_api_keys.sql

CREATE TABLE IF NOT EXISTS api_applications (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        TEXT        NOT NULL,
    is_live     BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Store both secret and publishable key hashes in the same row
CREATE TABLE IF NOT EXISTS api_keys (
    id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    app_id            UUID        NOT NULL REFERENCES api_applications(id) ON DELETE CASCADE,
    secret_hash       TEXT        NOT NULL UNIQUE, -- SHA-256 hash of sk_
    secret_hint       TEXT        NOT NULL,        -- Last 6 chars of sk_
    publishable_hash  TEXT        NOT NULL UNIQUE, -- SHA-256 hash of pk_
    publishable_hint  TEXT        NOT NULL,        -- Last 6 chars of pk_
    is_active         BOOLEAN     NOT NULL DEFAULT TRUE,
    last_used_at      TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS api_request_logs (
    id          BIGSERIAL   PRIMARY KEY,
    key_id      UUID        NOT NULL REFERENCES api_keys(id) ON DELETE CASCADE,
    method      TEXT        NOT NULL,
    path        TEXT        NOT NULL,
    status_code INTEGER     NOT NULL,
    latency_ms  INTEGER     NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_api_keys_app_id ON api_keys(app_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_secret_hash ON api_keys(secret_hash);
CREATE INDEX IF NOT EXISTS idx_api_keys_pub_hash ON api_keys(publishable_hash);
CREATE INDEX IF NOT EXISTS idx_api_request_logs_key ON api_request_logs(key_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_api_applications_user ON api_applications(user_id);
