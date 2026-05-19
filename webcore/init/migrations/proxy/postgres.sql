CREATE TABLE IF NOT EXISTS transactions (
    id VARCHAR(36) PRIMARY KEY,
    env VARCHAR(6) NOT NULL,
    resource_type VARCHAR(64) NOT NULL,
    patient_id VARCHAR(36),
    payload JSONB,
    error_message TEXT,
    status VARCHAR(20) DEFAULT 'PENDING',
    retry_count INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Index untuk mempercepat filter status dan retry oleh Cron
CREATE INDEX idx_transactions_status_retry ON transactions (status, retry_count) 
WHERE status = 'PENDING';