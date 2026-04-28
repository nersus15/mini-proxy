-- SCHEMA: access

-- +goose Up
-- +goose StatementBegin
-- consents definition

-- Drop table

-- DROP TABLE consents;

CREATE SCHEMA IF NOT EXISTS access;
SET search_path TO access, public;

CREATE TABLE consents (
  id varchar(50) NOT NULL PRIMARY KEY,
  organization_id varchar(50) NOT NULL,
  env varchar(10) NOT NULL,
  status varchar(10) NULL,
  patient_id varchar(50) NULL,
  provision_id varchar(50) NULL,
  created_at timestamp(0) NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp(0) NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_consents_organization_id ON consents (organization_id);
CREATE INDEX idx_consents_env ON consents (env);
CREATE INDEX idx_consents_status ON consents (status);
CREATE INDEX idx_consents_patient_id ON consents (patient_id);
CREATE INDEX idx_consents_provision_id ON consents (provision_id);
CREATE INDEX idx_consents_created_at ON consents (created_at);
CREATE INDEX idx_consents_updated_at ON consents (updated_at);

-- provisions definition

-- Drop table

-- DROP TABLE provisions;

CREATE TABLE provisions (
  id varchar(50) NOT NULL PRIMARY KEY,
  consent_id varchar(50) NOT NULL,
  parent_id varchar(50) NULL,
  provision_type varchar(8) NULL,
  provision_action varchar(8) NULL,
  period_start timestamp(0) NULL,
  period_end timestamp(0) NULL,
  data_period_start timestamp(0) NULL,
  data_period_end timestamp(0) NULL
);

CREATE INDEX idx_provisions_consent_id ON provisions (consent_id);
CREATE INDEX idx_provisions_parent_id ON provisions (parent_id);
CREATE INDEX idx_provisions_provision_type ON provisions (provision_type);
CREATE INDEX idx_provisions_provision_action ON provisions (provision_action);


-- provision_items definition

-- Drop table

-- DROP TABLE provision_items;

-- Create ENUM type for provision_item_type
CREATE TYPE provision_item_type AS ENUM ('actor_rule', 'action', 'security_label', 'purpose', 'class', 'code');

CREATE TABLE provision_items (
  id varchar(50) NOT NULL PRIMARY KEY,
  provision_id varchar(50) NOT NULL,
  "type" provision_item_type NOT NULL,
  code varchar(50) NULL,
  "system" varchar(100) NULL,
  display varchar(100) NULL
);

CREATE INDEX idx_provision_items_provision_id ON provision_items (provision_id);
CREATE INDEX idx_provision_items_type ON provision_items ("type");
CREATE INDEX idx_provision_items_code ON provision_items (code);
CREATE INDEX idx_provision_items_system ON provision_items ("system");
-- +goose StatementEnd