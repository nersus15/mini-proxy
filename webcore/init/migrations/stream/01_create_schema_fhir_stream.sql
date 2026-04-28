-- SCHEMA: logs

-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS logs;
SET search_path TO logs;

-- Enable UUID extension for better performance
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- failsafe definition

-- Drop table

-- DROP TABLE failsafe;

CREATE TABLE failsafe (
  id varchar(50) NOT NULL PRIMARY KEY,
  resource_type varchar(50) NOT NULL,
  env varchar(10) NOT NULL,
  entry jsonb DEFAULT '{}'::jsonb NOT NULL,
  err_code_db varchar(50) NULL,
  err_msg_db text NULL,
  err_code_hapi varchar(50) NULL,
  err_msg_hapi text NULL,
  err_code_satusehat varchar(50) NULL,
  err_msg_satusehat text NULL,
  created_at timestamp(0) NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp(0) NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_failsafe_resource_type ON failsafe (resource_type);
CREATE INDEX idx_failsafe_env ON failsafe (env);
CREATE INDEX idx_failsafe_created_at ON failsafe (created_at);
-- fasyankes definition

-- Drop table

-- DROP TABLE fasyankes;

CREATE TABLE fasyankes (
  id10 VARCHAR(16) NOT NULL PRIMARY KEY,
  id9 VARCHAR(16) NOT NULL,
  kode_sarana VARCHAR(16) DEFAULT NULL,
  kode_bpjs VARCHAR(16) DEFAULT NULL,
  kode_sisdmk VARCHAR(16) DEFAULT NULL,
  kode_pusdatin_baru VARCHAR(16) DEFAULT NULL,
  kode_pusdatin_lama VARCHAR(16) DEFAULT NULL,
  nama VARCHAR(128) DEFAULT NULL,
  longitude VARCHAR(28) DEFAULT NULL,
  latitude VARCHAR(28) DEFAULT NULL,
  alamat text DEFAULT NULL,
  provinsi_code VARCHAR(16) DEFAULT NULL,
  kabkota_code VARCHAR(16) DEFAULT NULL,
  kecamatan_code VARCHAR(16) DEFAULT NULL,
  kelurahan_code VARCHAR(16) DEFAULT NULL,
  jenis_sarana VARCHAR(40) DEFAULT NULL,
  subjenis VARCHAR(40) DEFAULT NULL,
  kelas_sarana VARCHAR(40) DEFAULT NULL,
  status_sarana VARCHAR(20) DEFAULT NULL,
  status_aktif BOOLEAN DEFAULT TRUE
);

CREATE INDEX idx_faskes_nama ON fasyankes (nama);
CREATE INDEX idx_faskes_provinsi_code ON fasyankes (provinsi_code);
CREATE INDEX idx_faskes_kabkota_code ON fasyankes (kabkota_code);
CREATE INDEX idx_faskes_kecamatan_code ON fasyankes (kecamatan_code);
CREATE INDEX idx_faskes_kelurahan_code ON fasyankes (kelurahan_code);
CREATE INDEX idx_faskes_organization_id ON fasyankes (id9);
CREATE INDEX idx_faskes_organization_id10 ON fasyankes (id10);
CREATE INDEX idx_faskes_kode_sarana ON fasyankes (kode_sarana);
CREATE INDEX idx_faskes_kode_bpjs ON fasyankes (kode_bpjs);
CREATE INDEX idx_faskes_kode_pusdatin_baru ON fasyankes (kode_pusdatin_baru);
CREATE INDEX idx_faskes_kode_pusdatin_lama ON fasyankes (kode_pusdatin_lama);
CREATE INDEX idx_faskes_jenis_sarana ON fasyankes (jenis_sarana);
CREATE INDEX idx_faskes_subjenis ON fasyankes (subjenis);
CREATE INDEX idx_faskes_kelas_sarana ON fasyankes (kelas_sarana);
CREATE INDEX idx_faskes_kode_sisdmk ON fasyankes (kode_sisdmk);

-- organizations definition

-- Drop table

-- DROP TABLE organizations;

CREATE TABLE organizations (
	id varchar(50) NOT NULL,
	satusehat_id varchar(50) NOT NULL,
	hapi_id varchar(50) NULL,
	"name" varchar(255) NOT NULL,
	body jsonb DEFAULT '{}'::jsonb NOT NULL,
	created_at timestamp(0) NULL,
	updated_at timestamp(0) NULL,
	deleted_at timestamp(0) NULL,
	dhis2_uid varchar(20) NULL,
	CONSTRAINT organizations_pkey PRIMARY KEY (satusehat_id),
	CONSTRAINT organizations_id_unique UNIQUE (id)
);

-- patients definition

-- Drop table

-- DROP TABLE patients;

CREATE TABLE patients (
	id varchar(50) NOT NULL,
	satusehat_id varchar(50) NOT NULL,
	hapi_id varchar(50) NULL,
	organization_id varchar(50) NULL,
	name varchar(255) NULL,
	telecom varchar(255) NULL,
	gender varchar(50) NULL,
	birth_date varchar(50) NULL,
	body jsonb DEFAULT '{}'::jsonb NOT NULL,
	deleted_at timestamp(0) NULL,
	created_at timestamp(0) NULL,
	updated_at timestamp(0) NULL,
	CONSTRAINT patients_pkey PRIMARY KEY (satusehat_id),
    CONSTRAINT patients_id_unique UNIQUE (id)
    -- CONSTRAINT patients_organization_id_foreign FOREIGN KEY (organization_id) REFERENCES organizations(satusehat_id)
);

-- practitioners definition

-- Drop table

-- DROP TABLE practitioners;

CREATE TABLE practitioners (
	id varchar(50) NOT NULL,
	satusehat_id varchar(50) NOT NULL,
	hapi_id varchar(50) NULL,
	organization_id varchar(50) NULL,
	name varchar(255) NULL,
	telecom varchar(255) NULL,
	gender varchar(50) NULL,
	birth_date varchar(50) NULL,
	qualitifation varchar(255) NULL,
	body jsonb DEFAULT '{}'::jsonb NOT NULL,
	deleted_at timestamp(0) NULL,
	created_at timestamp(0) NULL,
	updated_at timestamp(0) NULL,
	CONSTRAINT practitioners_pkey PRIMARY KEY (satusehat_id),
    CONSTRAINT practitioners_id_unique UNIQUE (id)
    -- CONSTRAINT practitioners_organization_id_foreign FOREIGN KEY (organization_id) REFERENCES organizations(satusehat_id)
);

-- practitioner_roles definition

-- Drop table

-- DROP TABLE practitioner_roles;

CREATE TABLE practitioner_roles (
	id varchar(50) NOT NULL,
	satusehat_id varchar(50) NOT NULL,
	hapi_id varchar(50) NULL,
	practitioner_id varchar(50) NULL,
	organization_id varchar(50) NULL,
	code varchar(50) NULL,
	period jsonb DEFAULT '{}'::jsonb NOT NULL,
	body jsonb DEFAULT '{}'::jsonb NOT NULL,
	deleted_at timestamp(0) NULL,
	created_at timestamp(0) NULL,
	updated_at timestamp(0) NULL,
	CONSTRAINT practitioner_roles_pkey PRIMARY KEY (satusehat_id),
    CONSTRAINT practitioner_roles_id_unique UNIQUE (id)
    -- CONSTRAINT practitioner_roles_practitioner_id_foreign FOREIGN KEY (practitioner_id) REFERENCES practitioners(satusehat_id),
    -- CONSTRAINT practitioner_roles_organization_id_foreign FOREIGN KEY (organization_id) REFERENCES organizations(satusehat_id)
);

-- locations definition

-- Drop table

-- DROP TABLE locations;

CREATE TABLE locations (
	id varchar(50) NOT NULL,
	satusehat_id varchar(50) NOT NULL,
	hapi_id varchar(50) NULL,
	organization_id varchar(50) NULL,
	part_of_id varchar(50) NULL,
	status varchar(50) NULL,
	name varchar(255) NULL,
	alias varchar(255) NULL,
	description text NULL,
	mode varchar(50) NULL,
	type varchar(50) NULL,
	telecom varchar(255) NULL,
	physical_type varchar(50) NULL,
	latitude varchar(50) NULL,
	longitude varchar(50) NULL,
	body jsonb DEFAULT '{}'::jsonb NOT NULL,
	deleted_at timestamp(0) NULL,
	created_at timestamp(0) NULL,
	updated_at timestamp(0) NULL,
	CONSTRAINT locations_pkey PRIMARY KEY (satusehat_id),
    CONSTRAINT locations_id_unique UNIQUE (id)
    -- CONSTRAINT locations_organization_id_foreign FOREIGN KEY (organization_id) REFERENCES organizations(satusehat_id),
    -- CONSTRAINT locations_part_of_id_foreign FOREIGN KEY (part_of_id) REFERENCES locations(satusehat_id)
);

-- healthcare_services definition

-- Drop table

-- DROP TABLE healthcare_services;

CREATE TABLE healthcare_services (
	id varchar(50) NOT NULL,
	satusehat_id varchar(50) NOT NULL,
	hapi_id varchar(50) NULL,
	organization_id varchar(50) NULL,
	location_id varchar(50) NULL,
	name varchar(255) NULL,
	comment varchar(255) NULL,
	provided_by varchar(255) NULL,
	body jsonb DEFAULT '{}'::jsonb NOT NULL,
	deleted_at timestamp(0) NULL,
	created_at timestamp(0) NULL,
	updated_at timestamp(0) NULL,
	CONSTRAINT healthcare_services_pkey PRIMARY KEY (satusehat_id),
    CONSTRAINT healthcare_services_id_unique UNIQUE (id)
    -- CONSTRAINT healthcare_services_organization_id_foreign FOREIGN KEY (organization_id) REFERENCES organizations(satusehat_id),
    -- CONSTRAINT healthcare_services_location_id_foreign FOREIGN KEY (location_id) REFERENCES locations(satusehat_id)
);

-- episode_of_cares definition

-- Drop table

-- DROP TABLE episode_of_cares;

CREATE TABLE episode_of_cares (
	id varchar(50) NOT NULL,
	satusehat_id varchar(50) NOT NULL,
	hapi_id varchar(50) NULL,
	organization_id varchar(50) NULL,
	-- encounter_id varchar(50) NULL, -- Many to Many
	patient_id varchar(50) NULL,
	"type" varchar(50) NULL,
	"status" varchar(16) NULL,
	body jsonb DEFAULT '{}'::jsonb NOT NULL,
	deleted_at timestamp(0) NULL,
	created_at timestamp(0) NULL,
	updated_at timestamp(0) NULL,
	CONSTRAINT episode_of_cares_pkey PRIMARY KEY (satusehat_id),
    CONSTRAINT episode_of_cares_id_unique UNIQUE (id)
	-- CONSTRAINT episode_of_cares_organization_id_foreign FOREIGN KEY (organization_id) REFERENCES organizations(satusehat_id)
);

-- encounters definition

-- Drop table

-- DROP TABLE encounters;

CREATE TABLE encounters (
	id varchar(50) NOT NULL,
	satusehat_id varchar(50) NOT NULL,
	hapi_id varchar(50) NULL,
	organization_id varchar(50) NOT NULL,
	-- episode_of_care_id varchar(50) NULL, -- Many to Many
	patient_id varchar(50) NULL,
	body jsonb DEFAULT '{}'::jsonb NOT NULL,
	encounter_diagnosis_condition varchar(255) NULL,
	encounter_diagnosis_procedure varchar(255) NULL,
	"basedOn" varchar(255) NULL,
	reason_condition varchar(255) NULL,
	reason_procedure varchar(255) NULL,
	reason_observation varchar(255) NULL,
	reason_immunization_recommendation varchar(255) NULL,
	created_at timestamp(0) NULL,
	updated_at timestamp(0) NULL,
	deleted_at timestamp(0) NULL,
	CONSTRAINT encounters_pkey PRIMARY KEY (satusehat_id),
    CONSTRAINT encounters_id_unique UNIQUE (id)
    -- CONSTRAINT encounters_episode_of_care_id_foreign FOREIGN KEY (episode_of_care_id) REFERENCES episode_of_cares(satusehat_id),
	-- CONSTRAINT encounters_organization_id_foreign FOREIGN KEY (organization_id) REFERENCES organizations(satusehat_id),
	-- CONSTRAINT encounters_patient_id_foreign FOREIGN KEY (patient_id) REFERENCES patients(satusehat_id)
);

-- conditions definition

-- Drop table

-- DROP TABLE conditions;

CREATE TABLE conditions (
	id varchar(50) NOT NULL,
	satusehat_id varchar(50) NOT NULL,
	hapi_id varchar(50) NULL,
	organization_id varchar(50) NULL,
	encounter_id varchar(50) NULL,
	patient_id varchar(50) NULL,
	code varchar(50) NULL,
	category varchar(20) NULL,
	body jsonb DEFAULT '{}'::jsonb NOT NULL,
	deleted_at timestamp(0) NULL,
	created_at timestamp(0) NULL,
	updated_at timestamp(0) NULL,
	CONSTRAINT conditions_pkey PRIMARY KEY (satusehat_id),
    CONSTRAINT conditions_id_unique UNIQUE (id)
    -- CONSTRAINT conditions_encounter_id_foreign FOREIGN KEY (encounter_id) REFERENCES encounters(satusehat_id),
	-- CONSTRAINT conditions_organization_id_foreign FOREIGN KEY (organization_id) REFERENCES organizations(satusehat_id)
);

-- "procedures" definition

-- Drop table

-- DROP TABLE "procedures";

CREATE TABLE "procedures" (
	id varchar(50) NOT NULL,
	satusehat_id varchar(50) NOT NULL,
	hapi_id varchar(50) NULL,
	organization_id varchar(50) NULL,
	encounter_id varchar(50) NULL,
	patient_id varchar(50) NULL,
	code varchar(50) NULL,
	code_system varchar(128) NULL,
	category varchar(20) NULL,
	body jsonb DEFAULT '{}'::jsonb NOT NULL,
	deleted_at timestamp(0) NULL,
	created_at timestamp(0) NULL,
	updated_at timestamp(0) NULL,
	CONSTRAINT procedures_pkey PRIMARY KEY (satusehat_id),
    CONSTRAINT procedures_id_unique UNIQUE (id)
    -- CONSTRAINT procedures_encounter_id_foreign FOREIGN KEY (encounter_id) REFERENCES encounters(satusehat_id),
	-- CONSTRAINT procedures_organization_id_foreign FOREIGN KEY (organization_id) REFERENCES organizations(satusehat_id)
);

-- questionnaire_responses definition

-- Drop table

-- DROP TABLE questionnaire_responses;

CREATE TABLE questionnaire_responses (
	id varchar(50) NOT NULL,
	satusehat_id varchar(50) NOT NULL,
	hapi_id varchar(50) NULL,
	organization_id varchar(50) NULL,
	encounter_id varchar(50) NULL,
	patient_id varchar(50) NULL,
	questionnaire varchar(100) NULL,
	body jsonb DEFAULT '{}'::jsonb NOT NULL,
	deleted_at timestamp(0) NULL,
	created_at timestamp(0) NULL,
	updated_at timestamp(0) NULL,
	CONSTRAINT questionnaire_responses_pkey PRIMARY KEY (satusehat_id),
    CONSTRAINT questionnaire_responses_id_unique UNIQUE (id)
    -- CONSTRAINT questionnaire_responses_encounter_id_foreign FOREIGN KEY (encounter_id) REFERENCES encounters(satusehat_id),
	-- CONSTRAINT questionnaire_responses_organization_id_foreign FOREIGN KEY (organization_id) REFERENCES organizations(satusehat_id)
);

-- immunizations definition

-- Drop table

-- DROP TABLE immunizations;

CREATE TABLE immunizations (
	id varchar(50) NOT NULL,
	satusehat_id varchar(50) NOT NULL,
	hapi_id varchar(50) NULL,
	organization_id varchar(50) NULL,
	encounter_id varchar(50) NULL,
	patient_id varchar(50) NULL,
	vaccine_code varchar(50) NULL,
	reason_code varchar(50) NULL,
	protocol_dose_number int NULL,
	protocol_dose_string varchar(50) NULL,
	body jsonb DEFAULT '{}'::jsonb NOT NULL,
	deleted_at timestamp(0) NULL,
	created_at timestamp(0) NULL,
	updated_at timestamp(0) NULL,
	CONSTRAINT immunizations_pkey PRIMARY KEY (satusehat_id),
    CONSTRAINT immunizations_id_unique UNIQUE (id)
    -- CONSTRAINT immunizations_encounter_id_foreign FOREIGN KEY (encounter_id) REFERENCES encounters(satusehat_id),
	-- CONSTRAINT immunizations_organization_id_foreign FOREIGN KEY (organization_id) REFERENCES organizations(satusehat_id)
);

-- allergyintolerances definition

-- Drop table

-- DROP TABLE allergyintolerances;

CREATE TABLE allergyintolerances (
	id varchar(50) NOT NULL,
	satusehat_id varchar(50) NOT NULL,
	hapi_id varchar(50) NULL,
	organization_id varchar(50) NULL,
	encounter_id varchar(50) NULL,
	patient_id varchar(50) NULL,
	"type" varchar(50) NULL,
    criticality varchar(50) NULL,
    code varchar(50) NULL,
	category varchar(20) NULL,
	body jsonb DEFAULT '{}'::jsonb NOT NULL,
	deleted_at timestamp(0) NULL,
	created_at timestamp(0) NULL,
	updated_at timestamp(0) NULL,
	CONSTRAINT allergyintolerances_pkey PRIMARY KEY (satusehat_id),
    CONSTRAINT allergyintolerances_id_unique UNIQUE (id)
    -- CONSTRAINT allergyintolerances_encounter_id_foreign FOREIGN KEY (encounter_id) REFERENCES encounters(satusehat_id),
	-- CONSTRAINT allergyintolerances_organization_id_foreign FOREIGN KEY (organization_id) REFERENCES organizations(satusehat_id)
);

-- family_member_histories definition

-- Drop table

-- DROP TABLE family_member_histories;

CREATE TABLE family_member_histories (
	id varchar(50) NOT NULL,
	satusehat_id varchar(50) NOT NULL,
	hapi_id varchar(50) NULL,
	organization_id varchar(50) NULL,
	patient_id varchar(50) NULL,
	"status" varchar(50) NULL,
	"name" varchar(20) NULL,
	relationship varchar(20) NULL,
	sex varchar(20) NULL,
	body jsonb DEFAULT '{}'::jsonb NOT NULL,
	deleted_at timestamp(0) NULL,
	created_at timestamp(0) NULL,
	updated_at timestamp(0) NULL,
	CONSTRAINT family_member_histories_pkey PRIMARY KEY (satusehat_id),
    CONSTRAINT family_member_histories_id_unique UNIQUE (id)
	-- CONSTRAINT family_member_histories_organization_id_foreign FOREIGN KEY (organization_id) REFERENCES organizations(satusehat_id)
);

-- compositions definition

-- Drop table

-- DROP TABLE compositions;

CREATE TABLE compositions (
	id varchar(50) NOT NULL,
	satusehat_id varchar(50) NOT NULL,
	hapi_id varchar(50) NULL,
	organization_id varchar(50) NULL,
	body jsonb DEFAULT '{}'::jsonb NOT NULL,
	deleted_at timestamp(0) NULL,
	created_at timestamp(0) NULL,
	updated_at timestamp(0) NULL,
	CONSTRAINT compositions_pkey PRIMARY KEY (satusehat_id),
    CONSTRAINT compositions_id_unique UNIQUE (id)
	-- CONSTRAINT compositions_organization_id_foreign FOREIGN KEY (organization_id) REFERENCES organizations(satusehat_id)
);

-- goals definition

-- Drop table

-- DROP TABLE goals;

CREATE TABLE goals (
	id varchar(50) NOT NULL,
	satusehat_id varchar(50) NOT NULL,
	hapi_id varchar(50) NULL,
	organization_id varchar(50) NULL,
	patient_id varchar(50) NULL,
	"priority" varchar(50) NULL,
	category varchar(20) NULL,
	body jsonb DEFAULT '{}'::jsonb NOT NULL,
	deleted_at timestamp(0) NULL,
	created_at timestamp(0) NULL,
	updated_at timestamp(0) NULL,
	CONSTRAINT goals_pkey PRIMARY KEY (satusehat_id),
    CONSTRAINT goals_id_unique UNIQUE (id)
	-- CONSTRAINT goals_organization_id_foreign FOREIGN KEY (organization_id) REFERENCES organizations(satusehat_id)
);

-- substances definition

-- Drop table

-- DROP TABLE substances;

CREATE TABLE substances (
	id varchar(50) NOT NULL,
	satusehat_id varchar(50) NOT NULL,
	hapi_id varchar(50) NULL,
	"status" varchar(50) NULL,
	code varchar(50) NULL,
	category varchar(20) NULL,
	body jsonb DEFAULT '{}'::jsonb NOT NULL,
	deleted_at timestamp(0) NULL,
	created_at timestamp(0) NULL,
	updated_at timestamp(0) NULL,
	CONSTRAINT substances_pkey PRIMARY KEY (satusehat_id),
    CONSTRAINT substances_id_unique UNIQUE (id)
);

-- clinical_impressions definition

-- Drop table

-- DROP TABLE clinical_impressions;

CREATE TABLE clinical_impressions (
	id varchar(50) NOT NULL,
	satusehat_id varchar(50) NOT NULL,
	hapi_id varchar(50) NULL,
	organization_id varchar(50) NULL,
	encounter_id varchar(50) NULL,
	patient_id varchar(50) NULL,
	"status" varchar(50) NULL,
    code varchar(50) NULL,
	body jsonb DEFAULT '{}'::jsonb NOT NULL,
	deleted_at timestamp(0) NULL,
	created_at timestamp(0) NULL,
	updated_at timestamp(0) NULL,
	CONSTRAINT clinical_impressions_pkey PRIMARY KEY (satusehat_id),
    CONSTRAINT clinical_impressions_id_unique UNIQUE (id)
    -- CONSTRAINT clinical_impressions_encounter_id_foreign FOREIGN KEY (encounter_id) REFERENCES encounters(satusehat_id),
	-- CONSTRAINT clinical_impressions_organization_id_foreign FOREIGN KEY (organization_id) REFERENCES organizations(satusehat_id)
);

-- careplans definition

-- Drop table

-- DROP TABLE careplans;

CREATE TABLE careplans (
	id varchar(50) NOT NULL,
	satusehat_id varchar(50) NOT NULL,
	hapi_id varchar(50) NULL,
	organization_id varchar(50) NULL,
	encounter_id varchar(50) NULL,
	patient_id varchar(50) NULL,
	goal_id varchar(50) NULL,
	"status" varchar(50) NULL,
	category varchar(20) NULL,
	body jsonb DEFAULT '{}'::jsonb NOT NULL,
	deleted_at timestamp(0) NULL,
	created_at timestamp(0) NULL,
	updated_at timestamp(0) NULL,
	CONSTRAINT careplans_pkey PRIMARY KEY (satusehat_id),
    CONSTRAINT careplans_id_unique UNIQUE (id),
    CONSTRAINT careplans_encounter_id_foreign FOREIGN KEY (encounter_id) REFERENCES encounters(satusehat_id)
	-- CONSTRAINT careplans_organization_id_foreign FOREIGN KEY (organization_id) REFERENCES organizations(satusehat_id),
	-- CONSTRAINT careplans_goal_id_foreign FOREIGN KEY (goal_id) REFERENCES goals(satusehat_id)
);

-- nutrition_orders definition

-- Drop table

-- DROP TABLE nutrition_orders;

CREATE TABLE nutrition_orders (
	id varchar(50) NOT NULL,
	satusehat_id varchar(50) NOT NULL,
	hapi_id varchar(50) NULL,
	organization_id varchar(50) NULL,
	encounter_id varchar(50) NULL,
	patient_id varchar(50) NULL,
	"status" varchar(50) NULL,
	-- allergyintolerance_id varchar(20) NULL, -- Many to Many
	body jsonb DEFAULT '{}'::jsonb NOT NULL,
	deleted_at timestamp(0) NULL,
	created_at timestamp(0) NULL,
	updated_at timestamp(0) NULL,
	CONSTRAINT nutrition_orders_pkey PRIMARY KEY (satusehat_id),
    CONSTRAINT nutrition_orders_id_unique UNIQUE (id)
    -- CONSTRAINT nutrition_orders_encounter_id_foreign FOREIGN KEY (encounter_id) REFERENCES encounters(satusehat_id),
	-- CONSTRAINT nutrition_orders_organization_id_foreign FOREIGN KEY (organization_id) REFERENCES organizations(satusehat_id)
);

-- service_requests definition

-- Drop table

-- DROP TABLE service_requests;

CREATE TABLE service_requests (
	id varchar(50) NOT NULL,
	satusehat_id varchar(50) NOT NULL,
	hapi_id varchar(50) NULL,
	organization_id varchar(50) NULL,
	encounter_id varchar(50) NULL,
	patient_id varchar(50) NULL,
	code varchar(50) NULL,
	category varchar(20) NULL,
	"priority" varchar(20) NULL,
	body jsonb DEFAULT '{}'::jsonb NOT NULL,
	deleted_at timestamp(0) NULL,
	created_at timestamp(0) NULL,
	updated_at timestamp(0) NULL,
	CONSTRAINT service_requests_pkey PRIMARY KEY (satusehat_id),
    CONSTRAINT service_requests_id_unique UNIQUE (id)
    -- CONSTRAINT service_requests_encounter_id_foreign FOREIGN KEY (encounter_id) REFERENCES encounters(satusehat_id),
	-- CONSTRAINT service_requests_organization_id_foreign FOREIGN KEY (organization_id) REFERENCES organizations(satusehat_id)
);

-- diagnostic_reports definition

-- Drop table

-- DROP TABLE diagnostic_reports;

CREATE TABLE diagnostic_reports (
	id varchar(50) NOT NULL,
	satusehat_id varchar(50) NOT NULL,
	hapi_id varchar(50) NULL,
	organization_id varchar(50) NULL,
	encounter_id varchar(50) NULL,
	patient_id varchar(50) NULL,
	code varchar(50) NULL,
	category varchar(20) NULL,
	body jsonb DEFAULT '{}'::jsonb NOT NULL,
	deleted_at timestamp(0) NULL,
	created_at timestamp(0) NULL,
	updated_at timestamp(0) NULL,
	CONSTRAINT diagnostic_reports_pkey PRIMARY KEY (satusehat_id),
    CONSTRAINT diagnostic_reports_id_unique UNIQUE (id)
    -- CONSTRAINT diagnostic_reports_encounter_id_foreign FOREIGN KEY (encounter_id) REFERENCES encounters(satusehat_id),
	-- CONSTRAINT diagnostic_reports_organization_id_foreign FOREIGN KEY (organization_id) REFERENCES organizations(satusehat_id)
);

-- specimens definition

-- Drop table

-- DROP TABLE specimens;

CREATE TABLE specimens (
	id varchar(50) NOT NULL,
	satusehat_id varchar(50) NOT NULL,
	hapi_id varchar(50) NULL,
	organization_id varchar(50) NULL,
	encounter_id varchar(50) NULL,
	patient_id varchar(50) NULL,
	diagnostic_report_id varchar(50) NULL,
	request_id varchar(50) NULL,
	"type" varchar(50) NULL,
	"status" varchar(20) NULL,
	body jsonb DEFAULT '{}'::jsonb NOT NULL,
	deleted_at timestamp(0) NULL,
	created_at timestamp(0) NULL,
	updated_at timestamp(0) NULL,
	CONSTRAINT specimens_pkey PRIMARY KEY (satusehat_id),
    CONSTRAINT specimens_id_unique UNIQUE (id),
    CONSTRAINT specimens_encounter_id_foreign FOREIGN KEY (encounter_id) REFERENCES encounters(satusehat_id)
	-- CONSTRAINT specimens_organization_id_foreign FOREIGN KEY (organization_id) REFERENCES organizations(satusehat_id),
	-- CONSTRAINT specimens_diagnostic_report_id_foreign FOREIGN KEY (diagnostic_report_id) REFERENCES diagnostic_reports(satusehat_id),
	-- CONSTRAINT specimens_patient_id_foreign FOREIGN KEY (patient_id) REFERENCES patients(satusehat_id)
);

-- observations definition

-- Drop table

-- DROP TABLE observations;

CREATE TABLE observations (
	id varchar(50) NOT NULL,
	satusehat_id varchar(50) NOT NULL,
	hapi_id varchar(50) NULL,
	organization_id varchar(50) NULL,
	encounter_id varchar(50) NULL,
	patient_id varchar(50) NULL,
    diagnostic_report_id varchar(50) NULL,
	code varchar(50) NULL,
	code_system varchar(128) NULL,
	category varchar(20) NULL,
	body jsonb DEFAULT '{}'::jsonb NOT NULL,
	value_data jsonb DEFAULT '{}'::jsonb NOT NULL,
	deleted_at timestamp(0) NULL,
	created_at timestamp(0) NULL,
	updated_at timestamp(0) NULL,
	CONSTRAINT observations_pkey PRIMARY KEY (satusehat_id),
    CONSTRAINT observations_id_unique UNIQUE (id)
    -- CONSTRAINT observations_encounter_id_foreign FOREIGN KEY (encounter_id) REFERENCES encounters(satusehat_id),
	-- CONSTRAINT observations_organization_id_foreign FOREIGN KEY (organization_id) REFERENCES organizations(satusehat_id),
	-- CONSTRAINT observations_diagnostic_report_id_foreign FOREIGN KEY (diagnostic_report_id) REFERENCES diagnostic_reports(satusehat_id),
	-- CONSTRAINT observations_patient_id_foreign FOREIGN KEY (patient_id) REFERENCES patients(satusehat_id)
);

-- imagingstudies definition

-- Drop table

-- DROP TABLE imagingstudies;

CREATE TABLE imagingstudies (
	id varchar(50) NOT NULL,
	satusehat_id varchar(50) NOT NULL,
	hapi_id varchar(50) NULL,
	organization_id varchar(50) NULL,
	encounter_id varchar(50) NULL,
	patient_id varchar(50) NULL,
    diagnostic_report_id varchar(50) NULL,
	modality varchar(50) NULL,
	"status" varchar(20) NULL,
	body jsonb DEFAULT '{}'::jsonb NOT NULL,
	deleted_at timestamp(0) NULL,
	created_at timestamp(0) NULL,
	updated_at timestamp(0) NULL,
	CONSTRAINT imagingstudies_pkey PRIMARY KEY (satusehat_id),
    CONSTRAINT imagingstudies_id_unique UNIQUE (id)
    -- CONSTRAINT imagingstudies_encounter_id_foreign FOREIGN KEY (encounter_id) REFERENCES encounters(satusehat_id),
	-- CONSTRAINT imagingstudies_organization_id_foreign FOREIGN KEY (organization_id) REFERENCES organizations(satusehat_id),
	-- CONSTRAINT imagingstudies_diagnostic_report_id_foreign FOREIGN KEY (diagnostic_report_id) REFERENCES diagnostic_reports(satusehat_id),
	-- CONSTRAINT imagingstudies_patient_id_foreign FOREIGN KEY (patient_id) REFERENCES patients(satusehat_id)
);

-- medications definition

-- Drop table

-- DROP TABLE medications;

CREATE TABLE medications (
	id varchar(50) NOT NULL,
	satusehat_id varchar(50) NOT NULL,
	hapi_id varchar(50) NULL,
	organization_id varchar(50) NULL,
	code varchar(50) NULL,
	body jsonb DEFAULT '{}'::jsonb NOT NULL,
	deleted_at timestamp(0) NULL,
	created_at timestamp(0) NULL,
	updated_at timestamp(0) NULL,
	CONSTRAINT medications_pkey PRIMARY KEY (satusehat_id),
    CONSTRAINT medications_id_unique UNIQUE (id)
	-- CONSTRAINT medications_organization_id_foreign FOREIGN KEY (organization_id) REFERENCES organizations(satusehat_id)
);

-- medication_requests definition

-- Drop table

-- DROP TABLE medication_requests;

CREATE TABLE medication_requests (
	id varchar(50) NOT NULL,
	satusehat_id varchar(50) NOT NULL,
	hapi_id varchar(50) NULL,
	organization_id varchar(50) NULL,
	encounter_id varchar(50) NULL,
	patient_id varchar(50) NULL,
	medication_id varchar(50) NULL,
	body jsonb DEFAULT '{}'::jsonb NOT NULL,
	deleted_at timestamp(0) NULL,
	created_at timestamp(0) NULL,
	updated_at timestamp(0) NULL,
	CONSTRAINT medication_requests_pkey PRIMARY KEY (satusehat_id),
    CONSTRAINT medication_requests_id_unique UNIQUE (id)
    -- CONSTRAINT medication_requests_encounter_id_foreign FOREIGN KEY (encounter_id) REFERENCES encounters(satusehat_id),
	-- CONSTRAINT medication_requests_organization_id_foreign FOREIGN KEY (organization_id) REFERENCES organizations(satusehat_id),
	-- CONSTRAINT medication_requests_medication_id_foreign FOREIGN KEY (medication_id) REFERENCES medications(satusehat_id)
);

-- medication_statements definition

-- Drop table

-- DROP TABLE medication_statements;

CREATE TABLE medication_statements (
	id varchar(50) NOT NULL,
	satusehat_id varchar(50) NOT NULL,
	hapi_id varchar(50) NULL,
	organization_id varchar(50) NULL,
	encounter_id varchar(50) NULL,
	patient_id varchar(50) NULL,
	medication_id varchar(50) NULL,
	body jsonb DEFAULT '{}'::jsonb NOT NULL,
	deleted_at timestamp(0) NULL,
	created_at timestamp(0) NULL,
	updated_at timestamp(0) NULL,
	CONSTRAINT medication_statements_pkey PRIMARY KEY (satusehat_id),
    CONSTRAINT medication_statements_id_unique UNIQUE (id)
    -- CONSTRAINT medication_statements_encounter_id_foreign FOREIGN KEY (encounter_id) REFERENCES encounters(satusehat_id),
	-- CONSTRAINT medication_statements_organization_id_foreign FOREIGN KEY (organization_id) REFERENCES organizations(satusehat_id),
	-- CONSTRAINT medication_statements_medication_id_foreign FOREIGN KEY (medication_id) REFERENCES medications(satusehat_id)
);

-- medication_administrations definition

-- Drop table

-- DROP TABLE medication_administrations;

CREATE TABLE medication_administrations (
	id varchar(50) NOT NULL,
	satusehat_id varchar(50) NOT NULL,
	hapi_id varchar(50) NULL,
	organization_id varchar(50) NULL,
	encounter_id varchar(50) NULL,
	patient_id varchar(50) NULL,
	"status" varchar(50) NULL,
	status_reason varchar(50) NULL,
	category varchar(50) NULL,
	medication_id varchar(50) NULL,
	medication_request_id varchar(50) NULL,
	body jsonb DEFAULT '{}'::jsonb NOT NULL,
	deleted_at timestamp(0) NULL,
	created_at timestamp(0) NULL,
	updated_at timestamp(0) NULL,
	CONSTRAINT medication_administrations_pkey PRIMARY KEY (satusehat_id),
    CONSTRAINT medication_administrations_id_unique UNIQUE (id)
    -- CONSTRAINT medication_administrations_encounter_id_foreign FOREIGN KEY (encounter_id) REFERENCES encounters(satusehat_id),
	-- CONSTRAINT medication_administrations_organization_id_foreign FOREIGN KEY (organization_id) REFERENCES organizations(satusehat_id),
	-- CONSTRAINT medication_administrations_medication_id_foreign FOREIGN KEY (medication_id) REFERENCES medications(satusehat_id),
	-- CONSTRAINT medication_administrations_medication_request_id_foreign FOREIGN KEY (medication_request_id) REFERENCES medication_requests(satusehat_id)
);

-- medication_dispenses definition

-- Drop table

-- DROP TABLE medication_dispenses;

CREATE TABLE medication_dispenses (
	id varchar(50) NOT NULL,
	satusehat_id varchar(50) NOT NULL,
	hapi_id varchar(50) NULL,
	organization_id varchar(50) NULL,
	encounter_id varchar(50) NULL,
	patient_id varchar(50) NULL,
	medication_id varchar(50) NULL,
	medication_request_id varchar(50) NULL,
	body jsonb DEFAULT '{}'::jsonb NOT NULL,
	deleted_at timestamp(0) NULL,
	created_at timestamp(0) NULL,
	updated_at timestamp(0) NULL,
	CONSTRAINT medication_dispenses_pkey PRIMARY KEY (satusehat_id),
    CONSTRAINT medication_dispenses_id_unique UNIQUE (id)
    -- CONSTRAINT medication_dispenses_encounter_id_foreign FOREIGN KEY (encounter_id) REFERENCES encounters(satusehat_id),
	-- CONSTRAINT medication_dispenses_organization_id_foreign FOREIGN KEY (organization_id) REFERENCES organizations(satusehat_id),
	-- CONSTRAINT medication_dispenses_medication_id_foreign FOREIGN KEY (medication_id) REFERENCES medications(satusehat_id),
	-- CONSTRAINT medication_dispenses_medication_request_id_foreign FOREIGN KEY (medication_request_id) REFERENCES medication_requests(satusehat_id)
);

-- risk_assessments definition

-- Drop table

-- DROP TABLE risk_assessments;

CREATE TABLE risk_assessments (
	id varchar(50) NOT NULL,
	satusehat_id varchar(50) NOT NULL,
	hapi_id varchar(50) NULL,
	organization_id varchar(50) NULL,
	encounter_id varchar(50) NULL,
	patient_id varchar(50) NULL,
	"status" varchar(50) NULL,
	method varchar(20) NULL,
	code varchar(20) NULL,
	body jsonb DEFAULT '{}'::jsonb NOT NULL,
	deleted_at timestamp(0) NULL,
	created_at timestamp(0) NULL,
	updated_at timestamp(0) NULL,
	CONSTRAINT risk_assessments_pkey PRIMARY KEY (satusehat_id),
    CONSTRAINT risk_assessments_id_unique UNIQUE (id)
    -- CONSTRAINT risk_assessments_encounter_id_foreign FOREIGN KEY (encounter_id) REFERENCES encounters(satusehat_id),
	-- CONSTRAINT risk_assessments_organization_id_foreign FOREIGN KEY (organization_id) REFERENCES organizations(satusehat_id)
);

CREATE TABLE related_person (
    id            varchar(50) NOT NULL,
    satusehat_id  varchar(50) NOT NULL,
    hapi_id       varchar(50) NOT NULL,
    patient_id    varchar(50) NOT NULL,
    name          varchar(255) NOT NULL,
	telecom 	varchar(255) NULL,
	gender varchar(50) NULL,
	birth_date varchar(50) NULL,
	body jsonb DEFAULT '{}'::jsonb NOT NULL,
	deleted_at timestamp(0) NULL,
	created_at timestamp(0) NULL,
	updated_at timestamp(0) NULL,
	CONSTRAINT related_person_pkey PRIMARY KEY (satusehat_id),
	CONSTRAINT related_person_id_unique UNIQUE (id)
)
-- +goose StatementEnd