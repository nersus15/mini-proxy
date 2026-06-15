-- +goose Up

-- ==========================================================
-- 1. TRANSACTIONS TABLE & INDEX
-- ==========================================================
CREATE TABLE IF NOT EXISTS transactions (
    id TEXT PRIMARY KEY,
    env TEXT NOT NULL,
    type TEXT NOT NULL,
    url TEXT NOT NULL,
    resource_type TEXT NOT NULL,
    patient_id TEXT,
    payload TEXT, 
    error_message TEXT,
    status TEXT DEFAULT 'PENDING',
    retry_count INTEGER DEFAULT 0,
    created_at TEXT DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
    updated_at TEXT DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW'))
);

CREATE INDEX IF NOT EXISTS idx_transactions_status_retry ON transactions (status, retry_count);


-- ==========================================================
-- 2. CLIENT CREDENTIALS TABLE & INDEX
-- ==========================================================
CREATE TABLE IF NOT EXISTS client_credentials (
    client_id TEXT NOT NULL,
    env TEXT NOT NULL,
    organization_id TEXT NULL,
    access_token TEXT NOT NULL,
    expired_at TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
    updated_at TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
    
    PRIMARY KEY (client_id),
    CONSTRAINT CHK_expired_at CHECK (expired_at > created_at)
);

CREATE INDEX IF NOT EXISTS idx_client_credentials_access_token ON client_credentials(access_token);
CREATE INDEX IF NOT EXISTS idx_client_credentials_env ON client_credentials(env);
CREATE INDEX IF NOT EXISTS idx_client_credentials_organization_id ON client_credentials(organization_id);
CREATE INDEX IF NOT EXISTS idx_client_credentials_expired_at ON client_credentials(expired_at DESC);