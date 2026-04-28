-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS proxy;
SET search_path TO proxy;

-- Enable UUID extension for better performance
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- SCHEMA: proxy

-- Create transactions table with optimized PostgreSQL data types
CREATE TABLE transactions (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    transaction_id VARCHAR(64) NOT NULL UNIQUE,
    env VARCHAR(5) NOT NULL DEFAULT 'prod',
    method VARCHAR(8),
    path VARCHAR(255),
    request JSONB NOT NULL,
    response JSONB,
    error JSONB,
    mediator JSONB,
    retry BOOLEAN NOT NULL DEFAULT false,
    retry_attempt INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(12) NOT NULL CHECK (status IN ('Processing', 'Failed', 'Completed', 'Successful', 'Completed with error(s)')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT "UQ_transaction_id" UNIQUE (transaction_id)
);

-- Create indexes for better query performance
CREATE INDEX idx_transactions_created_at ON transactions(created_at DESC);
CREATE INDEX idx_transactions_env ON transactions(env);
CREATE INDEX idx_transactions_method ON transactions("method");
CREATE INDEX idx_transactions_path ON transactions("path");
CREATE INDEX idx_transactions_status ON transactions(status);
CREATE INDEX idx_transactions_transaction_id ON transactions(transaction_id);

-- Create function to automatically update updatedAt timestamp
CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for automatic timestamp updates
CREATE TRIGGER set_timestamp_transactions
    BEFORE UPDATE ON transactions
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_timestamp();

-- Add comments for documentation
COMMENT ON TABLE transactions IS 'Stores all proxy transaction data with detailed request/response information';

COMMENT ON COLUMN transactions.transaction_id IS 'Unique identifier for the transaction';
COMMENT ON COLUMN transactions.request IS 'JSONB containing the complete request details';
COMMENT ON COLUMN transactions.response IS 'JSONB containing the complete response details';
COMMENT ON COLUMN transactions.error IS 'JSONB containing error information if transaction failed';
COMMENT ON COLUMN transactions.mediator IS 'JSONB containing mediator-specific data';
COMMENT ON COLUMN transactions.retry IS 'Flag indicating if this is a retry attempt';
COMMENT ON COLUMN transactions.retry_attempt IS 'Number of retry attempts made';
COMMENT ON COLUMN transactions.status IS 'Current status of the transaction';
-- +goose StatementEnd