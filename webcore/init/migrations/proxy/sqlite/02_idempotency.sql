-- +goose Up

-- ==========================================================
-- REQUEST IDEMPOTENCY TABLE & INDEX
-- ==========================================================
-- Respons sukses dari SatuSehat, dikunci hash body + client + env + method + url.
CREATE TABLE IF NOT EXISTS request_idempotency (
    fingerprint TEXT PRIMARY KEY,
    client TEXT NOT NULL,
    env TEXT NOT NULL,
    method TEXT NOT NULL,
    resource_type TEXT NOT NULL,
    url TEXT NOT NULL,
    resource_id TEXT,
    response_body TEXT,
    created_at TEXT DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
    expired_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_request_idempotency_expired_at ON request_idempotency(expired_at);
CREATE INDEX IF NOT EXISTS idx_request_idempotency_client ON request_idempotency(client, env);

-- +goose Down

DROP INDEX IF EXISTS idx_request_idempotency_client;
DROP INDEX IF EXISTS idx_request_idempotency_expired_at;
DROP TABLE IF EXISTS request_idempotency;
