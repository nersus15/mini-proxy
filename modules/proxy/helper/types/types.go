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

type Authorization struct {
	ClientID       string    `json:"client_id"`
	Env            string    `json:"env"`
	OrganizationID *string   `json:"organization_id,omitempty"`
	AccessToken    string    `json:"access_token"`
	ExpiredAt      time.Time `json:"expired_at"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// SatuSehatTokenResponse mewakili payload JSON response saat melakukan generate token.
type SatuSehatTokenResponse struct {
	RefreshTokenExpiresIn string   `json:"refresh_token_expires_in"`
	ApiProductList        string   `json:"api_product_list"` // Format string berstruktur array dari API
	ApiProductListJson    []string `json:"api_product_list_json"`
	OrganizationName      string   `json:"organization_name"`
	DeveloperEmail        string   `json:"developer.email"` // Menggunakan dot notation sesuai key JSON
	TokenType             string   `json:"token_type"`
	IssuedAt              string   `json:"issued_at"` // Epoch milidetik dalam bentuk string
	ClientID              string   `json:"client_id"`
	AccessToken           string   `json:"access_token"`
	ApplicationName       string   `json:"application_name"`
	Scope                 string   `json:"scope"`
	ExpiresIn             string   `json:"expires_in"` // Durasi detik dalam bentuk string
	RefreshCount          string   `json:"refresh_count"`
	Status                string   `json:"status"`
}

type TokenType struct {
	Env         string `json:"env"`
	AccessToken string `json:"access_token"`
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

func (r *Authorization) ToEntity() *entity.ClinetCredential {
	return &entity.ClinetCredential{
		ClientID:       r.ClientID,
		Env:            r.Env,
		OrganizationID: r.OrganizationID,
		AccessToken:    r.AccessToken,
		ExpiredAt:      r.ExpiredAt,
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
	}
}

func (r *SatuSehatTokenResponse) ToEntity() *entity.ClinetCredential {
	return &entity.ClinetCredential{
		ClientID:    r.ClientID,
		AccessToken: r.AccessToken,
	}
}
