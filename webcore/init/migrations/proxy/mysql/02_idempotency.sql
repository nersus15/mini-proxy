-- +goose Up

-- ==========================================================
-- REQUEST IDEMPOTENCY TABLE & INDEX
-- ==========================================================
CREATE TABLE IF NOT EXISTS request_idempotency (
    fingerprint VARCHAR(64) NOT NULL,
    client VARCHAR(255) NOT NULL,
    env VARCHAR(32) NOT NULL,
    method VARCHAR(16) NOT NULL,
    resource_type VARCHAR(128) NOT NULL,
    url TEXT NOT NULL,
    resource_id VARCHAR(255) NULL,
    response_body LONGTEXT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expired_at DATETIME NOT NULL,

    PRIMARY KEY (fingerprint),
    INDEX idx_request_idempotency_expired_at (expired_at),
    INDEX idx_request_idempotency_client (client, env)
);

-- +goose Down

DROP TABLE IF EXISTS request_idempotency;
