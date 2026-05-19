package types

import (
	"encoding/json"
	"time"

	"github.com/nersus15/mini-proxy/mod-proxy/entity"
)

type TransactionError struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"`
	Env          string    `json:"env"`
	Url          string    `json:"url"`
	ResourceType string    `json:"resource_type"`
	PatientId    string    `json:"patient_id"`
	Payload      []byte    `json:"payload"`
	ErrorMessage string    `json:"error_message"`
	Status       string    `json:"status"` // "PENDING", "FAILED", "RETRYING" "COMPLETE"
	RetryCount   int       `json:"retry_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (r *TransactionError) ToEntity() *entity.Transactions {
	return &entity.Transactions{
		ID:           r.ID,
		Type:         r.Type,
		Env:          r.Env,
		Url:          r.Url,
		ResourceType: r.ResourceType,
		PatientId:    r.PatientId,
		Payload:      json.RawMessage(r.Payload),
		ErrorMessage: r.ErrorMessage,
		Status:       r.Status,
		RetryCount:   r.RetryCount,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
	}
}
