package types

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/nersus15/mini-proxy/mod-proxy/entity"
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
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

// Struct Dari Library Fhir
const (
	NOTYET = iota
	FAILED
	DONE
)

type SetReferenceEntry struct {
	Index                int
	Priority             int
	Temporary            bool
	ReferredByTemp       []string
	ResourceType         string
	URN                  *string
	Id                   *string
	Entry                *BundleEntry
	InDB                 int
	IsDBChecked          bool
	ErrorCodeDB          *int
	ErrorStringDB        *string
	InHapi               int
	IsHapiChecked        bool
	ErrorCodeHapi        *int
	ErrorStringHapi      *string
	InSatusehat          int
	IsSatusehatChecked   bool
	ErrorCodeSatusehat   *int
	ErrorStringSatusehat *string
}
type SetReference map[string]*SetReferenceEntry

type NewPost struct {
	Index        int
	ResourceType string
	NewID        string
	TempID       string
}

type BaseResource struct {
	fhir.Resource

	// Tambahkan resourceType secara explisit
	ResourceType *string `bson:"resourceType" json:"resourceType"`
	ResourceReal any
}

type BundleEntryInterface interface {
	GetTemporaryID() string
}

type BundleEntry struct {
	fhir.BundleEntry

	Base *BaseResource
}

type BundleEntryResponse struct {
	fhir.BundleEntryResponse

	// Override
	Id           *string `bson:"resourceID" json:"resourceID"`
	ResourceType *string `bson:"resourceType" json:"resourceType"`
}

type Bundle struct {
	fhir.Bundle

	// Override
	Entry []BundleEntry `bson:"entry,omitempty" json:"entry,omitempty"`
}

func (e *BaseResource) MarshalJSON() ([]byte, error) {
	if e.ResourceReal != nil {
		return json.Marshal(e.ResourceReal)
	}

	return json.Marshal(BaseResource{
		Resource:     e.Resource,
		ResourceType: e.ResourceType,
	})
}

func (e *BundleEntry) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		fhir.BundleEntry
		Resource *BaseResource `bson:"resource,omitempty" json:"resource,omitempty"`
	}{
		BundleEntry: e.BundleEntry,
		Resource:    e.Base,
	})
}

func (e *Bundle) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ResourceType string          `bson:"resourceType" json:"resourceType"`
		Type         fhir.BundleType `bson:"type" json:"type"`
		Entry        []BundleEntry   `bson:"entry,omitempty" json:"entry,omitempty"`
	}{
		ResourceType: "Bundle",
		Type:         e.Bundle.Type,
		Entry:        e.Entry,
	})
}

func CastToBundle(old fhir.Bundle) Bundle {
	return Bundle{
		Bundle: old,
		Entry:  make([]BundleEntry, len(old.Entry)),
	}
}

func GetTemporaryReference(reference *string) *string {
	if IsTemporaryReference(reference) {
		url := *reference
		id := url[9:]
		return &id
	}

	return nil
}

func IsTemporaryReference(reference *string) bool {
	return reference != nil && strings.HasPrefix(*reference, "urn:uuid:")
}

func GetReferencePath(resourceType string, id string) string {
	return resourceType + "/" + id
}

func GetReferenceID(reference *fhir.Reference) *string {
	if reference == nil || reference.Reference == nil {
		return nil
	}
	if IsTemporaryReference(reference.Reference) {
		return nil
	}

	refParts := strings.Split(*reference.Reference, "/")
	if len(refParts) != 2 {
		return nil
	}

	return &refParts[1]
}

func UpdateReferenceID(reference *fhir.Reference, register *SetReference, baru NewPost, entry *BundleEntry) bool {
	tmpId := GetTemporaryReference(reference.Reference)
	if tmpId != nil && *tmpId == baru.TempID {
		ref := GetReferencePath(baru.ResourceType, baru.NewID)
		reference.Reference = &ref

		// cek jika entry adalah NewPost, maka tambahkan ke daftar ReferredByTemp
		sre, ok := (*register)[*entry.Base.Id]
		if ok && sre.Temporary {
			sre.ReferredByTemp = append(sre.ReferredByTemp, baru.NewID)
		}

		return true
	}

	return false
}

func UpdateReferenceArrayID(reference *([]fhir.Reference), register *SetReference, baru NewPost, entry *BundleEntry) bool {
	for i := range *reference {
		// UpdateReferenceID(&(*reference)[i], baru)
		tmpId := GetTemporaryReference((*reference)[i].Reference)
		if tmpId != nil && *tmpId == baru.TempID {
			ref := GetReferencePath(baru.ResourceType, baru.NewID)
			(*reference)[i].Reference = &ref

			// cek jika entry adalah NewPost, maka tambahkan ke daftar ReferredByTemp
			sre, ok := (*register)[*entry.Base.Id]
			if ok && sre.Temporary {
				sre.ReferredByTemp = append(sre.ReferredByTemp, baru.NewID)
			}

			return true
		}
	}

	return false
}

// ProxyHeaders adalah alias untuk map[string]string, merepresentasikan header HTTP.
type ProxyHeaders map[string]string

// ProxyRequestDefinition mendefinisikan struktur dari sebuah request yang di-proxy.
type ProxyRequestDefinition struct {
	Host        string       `json:"host"`
	Port        string       `json:"port"`
	Path        string       `json:"path"`
	Headers     ProxyHeaders `json:"headers"`
	Querystring string       `json:"querystring"`
	Body        string       `json:"body"`
	Method      string       `json:"method"`
	Timestamp   time.Time    `json:"timestamp"`
}

// ProxyResponseDefinition mendefinisikan struktur dari sebuah response dari layanan yang di-proxy.
type ProxyResponseDefinition struct {
	Status    int          `json:"status"`
	Headers   ProxyHeaders `json:"headers"`
	Body      string       `json:"body"`
	Timestamp time.Time    `json:"timestamp"`
}

// ProxyStatus merepresentasikan status dari sebuah transaksi proxy.
type ProxyStatus string

// Mendefinisikan nilai-nilai yang mungkin untuk ProxyStatus.
const (
	StatusProcessing          ProxyStatus = "Processing"
	StatusFailed              ProxyStatus = "Failed"
	StatusCompleted           ProxyStatus = "Completed"
	StatusSuccessful          ProxyStatus = "Successful"
	StatusCompletedWithErrors ProxyStatus = "Completed with error(s)"
)

// ProxyError mendefinisikan struktur dari sebuah error yang terjadi selama pemrosesan.
type ProxyError struct {
	Message string `json:"message"`
	Stack   string `json:"stack"`
}

// ProxyMediatorResponseDefinition mendefinisikan struktur dari response mediator.
type ProxyMediatorResponseDefinition struct {
	Status    int       `json:"status"`
	Body      string    `json:"body"`
	Timestamp time.Time `json:"timestamp"`
}

// ProxyMediator menyimpan detail dari proses mediasi.
type ProxyMediator struct {
	Status      ProxyStatus                      `json:"status"`
	RequestBody any                              `json:"requestBody,omitempty"`
	Response    *ProxyMediatorResponseDefinition `json:"response,omitempty"`
	Error       *ProxyError                      `json:"error,omitempty"`
}

// ProxyTransaction merepresentasikan sebuah transaksi lengkap melalui proxy.
type ProxyTransaction struct {
	TransactionID string                   `json:"transactionId"`
	Request       ProxyRequestDefinition   `json:"request"`
	Retry         bool                     `json:"retry"`
	RetryAttempt  int                      `json:"retryAttempt"`
	Status        ProxyStatus              `json:"status"`
	Response      *ProxyResponseDefinition `json:"response,omitempty"`
	Error         *ProxyError              `json:"error,omitempty"`
	Mediator      *ProxyMediator           `json:"mediator,omitempty"`
}

type PackMediator struct {
	TransactionID *string               `json:"transactionId,omitempty"`
	Method        *string               `json:"method"`
	Env           *string               `json:"env"`
	Authorization *string               `json:"authorization,omitempty"`
	MetaProfile   *[]string             `json:"metaProfile"`
	Patient       *string               `json:"patient"`
	Input         *fhir.Bundle          `json:"input"`
	Response      []BundleEntryResponse `json:"response,omitempty"` // bisa []fhir.BundleEntry | fhir.OperationOutcome
	// Bundle        *fhir.Bundle
}

// BackupPayload adalah objek utama yang dikirim ke NestJS
type BackupPayload struct {
	Request  interface{}   `json:"request"`  // Sesuai kebutuhan request awal kamu
	Response AxiosResponse `json:"response"` // Dipetakan mirip Axios
}

// AxiosResponse merepresentasikan interface AxiosResponse di TypeScript
type AxiosResponse struct {
	Data       interface{}         `json:"data"`              // Body response dari SATUSEHAT (JSON/Object/Array)
	Status     int                 `json:"status"`            // Contoh: 200, 400, 401
	StatusText string              `json:"statusText"`        // Contoh: "OK", "Bad Request"
	Headers    map[string][]string `json:"headers"`           // Header dari SATUSEHAT
	Config     AxiosRequestConfig  `json:"config"`            // Metadata internal request
	Request    interface{}         `json:"request,omitempty"` // Opsional jika ingin melacak detail raw request
}

// AxiosRequestConfig mencerminkan InternalAxiosRequestConfig untuk melengkapi struktur Axios
type AxiosRequestConfig struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Data    interface{}       `json:"data,omitempty"` // Body yang dikirim saat request ke SATUSEHAT
}
