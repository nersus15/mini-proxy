-- SCHEMA: access

-- +goose Up
-- +goose StatementBegin

CREATE SCHEMA IF NOT EXISTS "access";
SET search_path TO "access";

-- Enable UUID extension for better performance
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create client_credentials table with optimized PostgreSQL data types
CREATE TABLE client_credentials (
    client_id VARCHAR(64) NOT NULL PRIMARY KEY, -- saat ini client_id di satusehat panjangnya 48 karakter
    env VARCHAR(8) NOT NULL,
    organization_id VARCHAR(40) NULL,
    access_token VARCHAR(256) NOT NULL,
    -- token_type VARCHAR(50) DEFAULT 'Bearer',
    -- expires_in INTEGER,
    -- scope VARCHAR(255),
    -- token_issued_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expired_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT "CHK_expired_at" CHECK (expired_at > NOW())
    -- , CONSTRAINT "CHK_token_issued_at" CHECK (token_issued_at <= expired_at)
);

CREATE INDEX idx_client_credentials_client_id ON client_credentials(client_id);
CREATE INDEX idx_client_credentials_access_token ON client_credentials(access_token);
CREATE INDEX idx_client_credentials_env ON client_credentials(env);
CREATE INDEX idx_client_credentials_organization_id ON client_credentials(organization_id);
CREATE INDEX idx_client_credentials_expired_at ON client_credentials(expired_at DESC);
-- CREATE INDEX idx_client_credentials_token_issued_at ON client_credentials(token_issued_at);

-- Create function to automatically update updatedAt timestamp
CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for automatic timestamp updates
CREATE TRIGGER set_timestamp_client_credentials
    BEFORE UPDATE ON client_credentials
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_timestamp();

-- +goose StatementEnd