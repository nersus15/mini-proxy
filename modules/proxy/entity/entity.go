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
	Url          string          `bun:"url,notnull"`
	Type         string          `bun:"type,notnull,default:kafka"` // kafka | forward
	ResourceType string          `bun:"resource_type,notnull"`
	PatientId    string          `bun:"patient_id"`
	Payload      json.RawMessage `bun:"payload"`
	ErrorMessage string          `bun:"error_message,type:text"`
	Status       string          `bun:"status,default:'PENDING'"` // PENDING | COMPLETE
	RetryCount   int             `bun:"retry_count,default:0"`
	CreatedAt    time.Time       `bun:"created_at,default:current_timestamp"`
	UpdatedAt    time.Time       `bun:"updated_at,default:current_timestamp"`
}

func (t Transactions) TableName() string {
	return "transactions"
}
