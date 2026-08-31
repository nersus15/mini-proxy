package entity

import (
	"encoding/json"
	"time"

	"github.com/uptrace/bun"
)

type Entity interface {
	TableName() string
}

type Transactions struct {
	bun.BaseModel `bun:"table:transactions,alias:tr"`

	ID           string          `bun:"id,pk"` // Id Transaction (bundle) Id Resource (jika resource type != bundle)
	Env          string          `bun:"env,notnull"`
	Client       string          `bun:"client,notnull"`
	Url          string          `bun:"url,notnull"`
	Type         string          `bun:"type,notnull,default:kafka"` // kafka | forward
	ResourceType string          `bun:"resource_type,notnull"`
	PatientId    string          `bun:"patient_id"`
	Payload      json.RawMessage `bun:"payload"`
	ErrorMessage string          `bun:"error_message,nullzero,type:text"`
	Status       string          `bun:"status,default:'PENDING'"` // PENDING | COMPLETE
	RetryCount   int             `bun:"retry_count,default:0"`
	CreatedAt    time.Time       `bun:"created_at,default:current_timestamp"`
	UpdatedAt    time.Time       `bun:"updated_at,default:current_timestamp"`
}

type ClinetCredential struct {
	bun.BaseModel `bun:"table:client_credentials,alias:cc"`

	ClientID       string    `bun:"client_id,pk" json:"client_id"`
	Env            string    `bun:"env,notnull" json:"env"`
	OrganizationID *string   `bun:"organization_id,nullzero" json:"organization_id,omitempty"`
	AccessToken    string    `bun:"access_token,notnull" json:"access_token"`
	ExpiredAt      time.Time `bun:"expired_at,notnull" json:"expired_at"`
	CreatedAt      time.Time `bun:"created_at,default:current_timestamp" json:"created_at"`
	UpdatedAt      time.Time `bun:"updated_at,default:current_timestamp" json:"updated_at"`
}

// RequestIdempotency menyimpan respons sukses dari SatuSehat, dikunci dengan
// sidik jari request, supaya retry klien tidak membuat data ganda.
type RequestIdempotency struct {
	bun.BaseModel `bun:"table:request_idempotency,alias:ri"`

	Fingerprint  string          `bun:"fingerprint,pk"`
	Client       string          `bun:"client,notnull"`
	Env          string          `bun:"env,notnull"`
	Method       string          `bun:"method,notnull"`
	ResourceType string          `bun:"resource_type,notnull"`
	Url          string          `bun:"url,notnull"`
	ResourceID   string          `bun:"resource_id,nullzero"`
	ResponseBody json.RawMessage `bun:"response_body"`
	CreatedAt    time.Time       `bun:"created_at,default:current_timestamp"`
	ExpiredAt    time.Time       `bun:"expired_at,notnull"`
}

func (t Transactions) TableName() string {
	return "transactions"
}

func (t RequestIdempotency) TableName() string {
	return "request_idempotency"
}

func (t ClinetCredential) TableName() string {
	return "client_credentials"
}
