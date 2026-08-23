-- 010_webhooks.sql

CREATE TABLE IF NOT EXISTS webhooks (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    app_id     UUID        NOT NULL REFERENCES api_applications(id) ON DELETE CASCADE,
    url        TEXT        NOT NULL,
    events     TEXT[]      NOT NULL DEFAULT '{}',
    secret     TEXT        NOT NULL,
    is_active  BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_webhooks_app_id ON webhooks(app_id);
