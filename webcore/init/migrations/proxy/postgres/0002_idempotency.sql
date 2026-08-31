-- +goose Up

-- ==========================================================
-- REQUEST IDEMPOTENCY TABLE & INDEX
-- ==========================================================
CREATE TABLE IF NOT EXISTS request_idempotency (
    fingerprint VARCHAR(64) PRIMARY KEY,
    client VARCHAR(255) NOT NULL,
    env VARCHAR(32) NOT NULL,
    method VARCHAR(16) NOT NULL,
    resource_type VARCHAR(128) NOT NULL,
    url TEXT NOT NULL,
    resource_id VARCHAR(255) NULL,
    response_body TEXT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expired_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_request_idempotency_expired_at ON request_idempotency(expired_at);
CREATE INDEX IF NOT EXISTS idx_request_idempotency_client ON request_idempotency(client, env);

-- +goose Down

DROP INDEX IF EXISTS idx_request_idempotency_client;
DROP INDEX IF EXISTS idx_request_idempotency_expired_at;
DROP TABLE IF EXISTS request_idempotency;
