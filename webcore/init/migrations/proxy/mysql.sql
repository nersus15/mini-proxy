CREATE TABLE IF NOT EXISTS transactions (
    id VARCHAR(36) PRIMARY KEY,
    env VARCHAR(6) NOT NULL,
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