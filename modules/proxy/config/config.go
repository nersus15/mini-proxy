package config

import (
	config2 "github.com/semanggilab/lib-go-fhir/config"
	"github.com/webcore-go/webcore/infra/config"
)

type ModuleConfig struct {
	Hapi                   config2.HapiConfig         `mapstructure:"hapi" json:"hapi"`
	Production             config2.SatuSehatConfig    `mapstructure:"production" json:"production"`
	Development            config2.SatuSehatConfig    `mapstructure:"development" json:"development"`
	Propagation            SatusehatPropagationConfig `mapstructure:"propagation" json:"propagation"`
	ResourceUseMasterToken []string                   `mapstructure:"use_master_token" json:"use_master_token"`
	FhirSource             FhirSource                 `mapstructure:"source" json:"source"`
	Kafka                  config.KafkaConfig         `mapstructure:"kafka" json:"kafka"`
	Cron                   CronConfig                 `mapstructru:"cron" json:"cron"`
	Ildki                  IldkiConfig                `mapstructure:"hapi" json:"json"`
}

type SatusehatPropagationConfig struct {
}

type IldkiConfig struct {
	Enabled        bool   `mapstructure:"enabled" json:"enabled"`
	DevAuthURL     string `mapstructure:"development_authurl" json:"development_authurl"`
	ProdAuthURL    string `mapstructure:"production_authurl" json:"production_authurl"`
	ProductionURL  string `mapstructure:"production_url" json:"production_url"`
	DevelopmentURL string `mapstructure:"development_url" json:"development_url"`
}

type FhirSource struct {
	Priority string `mapstructure:"priority" json:"priority"`
	Header   string `mapstructure:"header" json:"header"`
}

type CronConfig struct {
	Enabled  bool   `mapstructure:"enabled" json:"enabled"`
	Schedule string `mapstructure:"schedule" json:"schedule"`
}

func (c *ModuleConfig) SetEnvBindings() map[string]string {
	return map[string]string{
		"module.fhir.ildki.enabled":             "MODULE_FHIR_ILDKI_ENABLED",
		"module.fhir.ildki.production_url":      "MODULE_FHIR_ILDKI_PRODUCTION_URL",
		"module.fhir.ildki.development_url":     "MODULE_FHIR_ILDKI_DEVELOPMENT_URL",
		"module.fhir.ildki.development_authurl": "MODULE_FHIR_ILDKI_DEVELOPMENT_AUTHURL",
		"module.fhir.ildki.production_authurl":  "MODULE_FHIR_ILDKI_PRODUCTION_AUTHURL",

		"module.fhir.production.baseurl":       "MODULE_FHIR_PRODUCTION_BASEURL",
		"module.fhir.production.authurl":       "MODULE_FHIR_PRODUCTION_AUTHURL",
		"module.fhir.production.client_id":     "MODULE_FHIR_PRODUCTION_CLIENT_ID",
		"module.fhir.production.client_secret": "MODULE_FHIR_PRODUCTION_CLIENT_SECRET",
		"module.fhir.production.http_proxy":    "MODULE_FHIR_PRODUCTION_HTTP_PROXY",

		"module.fhir.development.baseurl":       "MODULE_FHIR_DEVELOPMENT_BASEURL",
		"module.fhir.development.authurl":       "MODULE_FHIR_DEVELOPMENT_AUTHURL",
		"module.fhir.development.client_id":     "MODULE_FHIR_DEVELOPMENT_CLIENT_ID",
		"module.fhir.development.client_secret": "MODULE_FHIR_DEVELOPMENT_CLIENT_SECRET",
		"module.fhir.development.http_proxy":    "MODULE_FHIR_DEVELOPMENT_HTTP_PROXY",

		"module.fhir.use_master_token": "MODULE_FHIR_RESOURCE_USE_MASTER_TOKEN",

		"module.fhir.propagation.auto":   "MODULE_FHIR_PROPAGATION_AUTO",
		"module.fhir.propagation.header": "MODULE_FHIR_PROPAGATION_HEADER",

		"module.fhir.source.priority": "MODULE_FHIR_SOURCE_PRIORITY",
		"module.fhir.source.header":   "MODULE_FHIR_SOURCE_PRIORITY_HEADER",

		// Kafka
		"kafka.enabled": "KAFKA_ENABLED",
		"kafka.brokers": "KAFKA_BROKERS",
		"kafka.group":   "KAFKA_GROUP_ID",
		"kafka.offset":  "KAFKA_AUTO_OFFSET_RESET",
		"kafka.topics":  "KAFKA_TOPICS",
		"cron.enabled":  "CRON_ENABLED",
		"cron.schedule": "CRON_SCHEDULE",
	}
}

func (c *ModuleConfig) SetDefaults() map[string]any {
	return map[string]any{
		"module.fhir.hapi.enabled":             true,
		"module.fhir.hapi.production_url":      "https://api.ildki.appgo.my.id/prod/fhir/r4/v1",
		"module.fhir.hapi.development_url":     "https://api.ildki.appgo.my.id/dev/fhir/r4/v1",
		"module.fhir.hapi.development_authurl": "https://api.ildki.appgo.my.id/dev/oauth2/v1/accesstoken?grant_type=client_credentials",
		"module.fhir.hapi.production_authurl":  "https://api.ildki.appgo.my.id/prod/oauth2/v1/accesstoken?grant_type=client_credentials",

		"module.fhir.production.baseurl":       "https://api-satusehat.kemkes.go.id/fhir-r4/v1",
		"module.fhir.production.authurl":       "https://api-satusehat.kemkes.go.id/oauth2/v1/accesstoken?grant_type=client_credentials",
		"module.fhir.production.client_id":     "your-production-client-id",
		"module.fhir.production.client_secret": "your-production-client-secret",

		"module.fhir.development.baseurl":       "https://api-satusehat-stg.dto.kemkes.go.id/fhir-r4/v1",
		"module.fhir.development.authurl":       "https://api-satusehat-stg.dto.kemkes.go.id/oauth2/v1/accesstoken?grant_type=client_credentials",
		"module.fhir.development.client_id":     "your-development-client-id",
		"module.fhir.development.client_secret": "your-development-client-secret",

		"module.fhir.use_master_token": []string{},

		"module.fhir.propagation.auto":   false,
		"module.fhir.propagation.header": "X-Propagation-Strategy", // pilihan ['satusehat-failover', 'local-only']

		"module.fhir.source.priority": "satusehat-first",
		"module.fhir.source.header":   "X-FHIR-Source-Priority",
		"cron.enabled":                true,
		"cron.schedule":               "59 23 * * *",
	}
}

func (c *ModuleConfig) ToFhirConfig() *config2.FhirTransactionConfig {
	config := config2.FhirTransactionConfig{
		Hapi:                   c.Hapi,
		Production:             c.Production,
		Development:            c.Development,
		ResourceUseMasterToken: c.ResourceUseMasterToken,
	}

	return &config
}
