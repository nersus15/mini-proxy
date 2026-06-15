CREATE TABLE IF NOT EXISTS transactions (
    id VARCHAR(36) PRIMARY KEY,
    env VARCHAR(6) NOT NULL,
    type VARCHAR(10) NOT NULL,
    url VARCHAR(64) NOT NULL,
    resource_type VARCHAR(64) NOT NULL,
    patient_id VARCHAR(36),
    payload JSON,
    error_message TEXT,
    status VARCHAR(20) DEFAULT 'PENDING',
    retry_count INT DEFAULT 0,
    created_at DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6)
);

-- Index untuk mempercepat proses Cron retry
CREATE INDEX idx_transactions_status_retry ON transactions (status, retry_count);


-- Create client_credentials table with optimized MySQL data types
CREATE TABLE client_credentials (
    client_id VARCHAR(64) NOT NULL, -- saat ini client_id di satusehat panjangnya 48 karakter
    env VARCHAR(8) NOT NULL,
    organization_id VARCHAR(40) NULL,
    access_token VARCHAR(256) NOT NULL,
    -- token_type VARCHAR(50) DEFAULT 'Bearer',
    -- expires_in INT,
    -- scope VARCHAR(255),
    -- token_issued_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    expired_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    PRIMARY KEY (client_id),
    CONSTRAINT CHK_expired_at CHECK (expired_at > created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Di MySQL, Primary Key otomatis membuat index, jadi idx_client_credentials_client_id tidak perlu dibuat lagi secara eksplisit.
CREATE INDEX idx_client_credentials_access_token ON client_credentials(access_token);
CREATE INDEX idx_client_credentials_env ON client_credentials(env);
CREATE INDEX idx_client_credentials_organization_id ON client_credentials(organization_id);
CREATE INDEX idx_client_credentials_expired_at ON client_credentials(expired_at DESC);
-- CREATE INDEX idx_client_credentials_token_issued_at ON client_credentials(token_issued_at);

-- +goose StatementEnd