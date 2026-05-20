package service

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/nersus15/mini-proxy/mod-proxy/config"
	types2 "github.com/nersus15/mini-proxy/mod-proxy/helper"
	"github.com/nersus15/mini-proxy/mod-proxy/repository"
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
	"github.com/semanggilab/lib-go-fhir/helper/types"
	"github.com/semanggilab/lib-go-fhir/helper/utils"
	"github.com/semanggilab/lib-go-fhir/processor"
	kafka "github.com/webcore-go/lib-kafka"
	"github.com/webcore-go/webcore/app/core"
	"github.com/webcore-go/webcore/app/helper"
	"github.com/webcore-go/webcore/infra/logger"
)

type ProxyService struct {
	Context        *core.AppContext
	Config         *config.ModuleConfig
	Repository     *repository.ProxyRepository
	Token          *string
	DevHttpClient  *http.Client
	ProdHttpClient *http.Client
	kafka          *kafka.KafkaProducer
}

func NewProxyService(wctx *core.AppContext, cfg *config.ModuleConfig, repository *repository.ProxyRepository, kafkaProducer *kafka.KafkaProducer) *ProxyService {
	return &ProxyService{
		Context:        wctx,
		Config:         cfg,
		Repository:     repository,
		DevHttpClient:  createHttpClient(cfg.Development.HttpProxy),
		ProdHttpClient: createHttpClient(cfg.Development.HttpProxy),
		kafka:          kafkaProducer,
	}
}

func (s *ProxyService) PostResource(env string, resourceType string, ctx *fiber.Ctx) (*types.BaseResource, any, int, string) {
	auth := ctx.Get("Authorization")

	mainUrl, forwardUrl, priority := s.GetUrl(env, resourceType, ctx)
	target, forward := s.GetTarget(priority)

	if (mainUrl == "" || mainUrl == "/") || (forwardUrl == "" || forwardUrl == "/") {
		return nil, nil, 400, "priority must be 'local-first' or 'satusehat-first'"
	}

	logger.Info("Send POST Request Resource ==> ", "ENV", env, "Resource Type", resourceType, "To", mainUrl, "Forward", forwardUrl)

	resource, raw, errcode, errstr := s.sendRequest("POST", mainUrl, env, target, auth, ctx.Body())

	if target == "hapi" {
		// Jika local-first tidak perlu forwad request ke SatuSehat melalui mini proxy, karena sudah di handle oleh ILDKI
		return resource, raw, 0, ""
	}
	if errcode == 0 && errstr == "" {
		// Sebelum di forward tambahkan id nya, ambil dari response
		newBody, err := json.Marshal(resource)
		if err != nil {
			logger.Error("json.Marshal", helper.ToLogJSON(err))
		} else {
			// forward ke internal hapi via api (bukan kafka)
			s.forwardRequest("PUT", forwardUrl, env, forward, auth, newBody)
		}
	}

	return resource, raw, 0, ""
}

func (s *ProxyService) PostBundle(env string, resourceType string, ctx *fiber.Ctx) (*types.BaseResource, any, int, string) {
	auth := ctx.Get("Authorization")

	mainUrl, forwardUrl, priority := s.GetUrl(env, resourceType, ctx)
	target, _ := s.GetTarget(priority)

	if (mainUrl == "" || mainUrl == "/") || (forwardUrl == "" || forwardUrl == "/") {
		return nil, nil, 400, "priority must be 'local-first' or 'satusehat-first'"
	}

	logger.Info("Send POST Request Bundle ==> ", "ENV", env, "Resource Type", resourceType, "To", mainUrl, "Forward", forwardUrl)

	resource, raw, errcode, errstr := s.sendRequest("POST", mainUrl, env, target, auth, ctx.Body())

	if target == "hapi" {
		// Jika local-first tidak perlu forwad request ke SatuSehat melalui mini proxy, karena sudah di handle oleh ILDKI
		return resource, raw, 0, ""
	}
	if errcode == 0 && errstr == "" {
		// Untuk Bundle Forwardnya harus ke kafka, karena jika melalui gateway ildki pasti akan error ketika kirim ke satusehat (duplikat), sehingga tidak diteruskan ke kafka
		var input fhir.Bundle
		err := ctx.BodyParser(&input)
		if err != nil {
			logger.ErrorJson("Gagal Parse Body", err)
		}
		var transactionId *string
		xtransId := ctx.Get("X-Transaction-Id", "")
		if xtransId != "" {
			transactionId = helper.StringPtr(xtransId)
		} else {
			transactionId = helper.StringPtr(uuid.New().String())
		}

		s.sendToKafka(transactionId, env, auth, input, resource, raw, "Bundle", nil)
	}

	return resource, raw, 0, ""
}

func (s *ProxyService) PutResource(env string, resourceType string, id string, ctx *fiber.Ctx) (*types.BaseResource, any, int, string) {
	auth := ctx.Get("Authorization")

	mainUrl, forwardUrl, priority := s.GetUrl(env, resourceType, ctx)
	target, forward := s.GetTarget(priority)

	if (mainUrl == "" || mainUrl == "/") || (forwardUrl == "" || forwardUrl == "/") {
		return nil, nil, 400, "priority must be 'local-first' or 'satusehat-first'"
	}
	mainUrl += "/" + id
	forwardUrl += "/" + id
	logger.Info("Send PUT Request ==> ", "ENV", env, "Resource Type", resourceType, "To", mainUrl, "Forward", forwardUrl)

	resource, raw, errcode, errstr := s.sendRequest("PUT", mainUrl, env, target, auth, ctx.Body())

	if target == "hapi" {
		// Jika local-first tidak perlu forwad request ke SatuSehat melalui mini proxy, karena sudah di handle oleh ILDKI
		return resource, raw, 0, ""
	}

	if errcode == 0 && errstr == "" {
		// Sebelum di forward tambahkan id nya, ambil dari response
		newBody, err := json.Marshal(resource)
		if err != nil {
			logger.Error("json.Marshal", helper.ToLogJSON(err))
		} else {
			// forward ke internal hapi via api (bukan kafka)
			s.forwardRequest("PUT", forwardUrl, env, forward, auth, newBody)
		}
	}

	return resource, raw, 0, ""
}

func (s *ProxyService) GenerateToken(env string, target string, clientId string, clientSecret string) (types2.SatuSehatTokenResponse, int, string) {
	var endpoint string
	var errstr string
	var errcode int
	var satsetRes types2.SatuSehatTokenResponse

	if env == "dev" {
		if target == "hapi" {
			endpoint = s.Config.Ildki.DevAuthURL
		} else {
			endpoint = s.Config.Development.AuthURL
		}
	} else {
		if target == "hapi" {
			endpoint = s.Config.Ildki.ProdAuthURL
		} else {
			endpoint = s.Config.Production.AuthURL
		}
	}
	logger.Info("Request Token", "ENV", env, "Target", target, "Endpoint", endpoint)

	data := url.Values{}
	data.Set("client_id", clientId)
	data.Set("client_secret", clientSecret)

	payload := strings.NewReader(data.Encode())

	req, err := http.NewRequest("POST", endpoint, payload)
	if err != nil {
		errcode = utils.ERR_SATUSEHAT_FORMAT
		if target == "hapi" {
			logger.Warn("Failed creating request to ILDKI, falling back to SATUSEHAT")
			return s.GenerateToken(env, "satusehat", clientId, clientSecret)
		}
		return satsetRes, errcode, fmt.Sprintf("gagal membuat request: %s", err)
	}
	req.Header.Add("content-type", "application/x-www-form-urlencoded")

	response, err := s.httpClient(&env).Do(req)
	if err != nil {
		errcode, errstr = utils.HttpError("satusehat", req, nil, err)
		logger.Error("Connection Error", "To", endpoint, "env", env, "Err", errstr, "code", errcode)

		if target == "hapi" {
			logger.Warn("Failed connecting to ILDKI, falling back to SATUSEHAT")
			return s.GenerateToken(env, "satusehat", clientId, clientSecret)
		}
		return satsetRes, errcode, errstr
	}
	defer response.Body.Close()

	bodyBytes, _ := io.ReadAll(response.Body)

	if response.StatusCode != http.StatusOK {
		if target == "hapi" {
			logger.Warn(fmt.Sprintf("HAPI returned status %d, falling back to SATUSEHAT", response.StatusCode))
			return s.GenerateToken(env, "satusehat", clientId, clientSecret)
		}
		return satsetRes, response.StatusCode, fmt.Sprintf("upstream returned error status: %d body: %s", response.StatusCode, string(bodyBytes))
	}

	err = json.Unmarshal(bodyBytes, &satsetRes)
	if err != nil {
		if target == "hapi" {
			logger.Warn("Failed to unmarshal ILDKI response, falling back to SATUSEHAT")
			return s.GenerateToken(env, "satusehat", clientId, clientSecret)
		}
		return satsetRes, errcode, fmt.Sprintf("gagal unmarshall Response request: %s", err)
	}

	logger.InfoJson("ClientCredential Dari "+target, satsetRes)
	return satsetRes, errcode, errstr
}

func (s *ProxyService) SendCredentialToProxyIL() {}

func (s *ProxyService) SaveCredential(env string, credential *types2.SatuSehatTokenResponse) {
	logger.Info("Save User Credential In Background Process")
	go func() {
		entityData := credential.ToEntity()

		var issuedAtUnixMilli int64
		if credential.IssuedAt != "" {
			var err error
			issuedAtUnixMilli, err = strconv.ParseInt(credential.IssuedAt, 10, 64)
			if err != nil {
				logger.Error("Failed to parse issued_at, fallback to time.Now()", "error", err)
				issuedAtUnixMilli = time.Now().UnixMilli()
			}
		} else {
			issuedAtUnixMilli = time.Now().UnixMilli()
		}

		expiresInSeconds, err := strconv.ParseInt(credential.ExpiresIn, 10, 64)
		if err != nil {
			logger.Error("Failed to parse expires_in", "error", err)
			return
		}

		issuedAtTime := time.UnixMilli(issuedAtUnixMilli)
		expiredAt := issuedAtTime.Add(time.Duration(expiresInSeconds) * time.Second)

		entityData.ExpiredAt = expiredAt
		entityData.Env = env

		err = s.Repository.SaveClientCredentials(entityData)

		if err != nil {
			logger.ErrorJson("Error saving client credential", err)
			return
		}

		logger.Info("Successfully saved client credential", "client_id", entityData.ClientID, "env", env)
	}()
}

func (s *ProxyService) sendRequest(method string, url string, env string, target string, auth string, body []byte) (*types.BaseResource, any, int, string) {
	var errcode int
	var errstr string
	logger.Info("===== Forward Request di Background =====", "Method", method, "Url", url, "Env", env, "Target", target)

	req, err := http.NewRequest(method, url, strings.NewReader(string(body)))
	if err != nil {
		errcode = utils.ERR_SATUSEHAT_FORMAT
		if target == "hapi" {
			errcode = utils.ERR_HAPI_FORMAT
		}

		return nil, nil, errcode, fmt.Sprintf("gagal membuat request: %s", err)
	}
	// Set Request Header
	req.Header.Add("Authorization", auth)
	req.Header.Add("content-type", "application/json")

	if target == "hapi" {
		req.Header.Add(s.Config.FhirSource.Header, "local-first")
	}

	response, err := s.httpClient(&env).Do(req)
	if err != nil {
		errcode, errstr = utils.HttpError(target, req, response, nil)

		return nil, nil, errcode, errstr
	}
	defer response.Body.Close()

	bodyBytes, _ := io.ReadAll(response.Body)
	resource, err := processor.UnmarshalResource(bodyBytes, nil)
	logger.InfoJson("Response "+url, resource)

	if err != nil {
		errcode = utils.ERR_SATUSEHAT_FORMAT
		if target == "hapi" {
			errcode = utils.ERR_HAPI_FORMAT
		}
		return nil, nil, errcode, fmt.Sprintf("gagal mem-parsing JSON: %s", err)
	}

	if resource.ResourceType == nil {
		errcode = utils.ERR_SATUSEHAT_FORMAT
		if target == "hapi" {
			errcode = utils.ERR_HAPI_FORMAT
		}
		return nil, nil, errcode, "resourceType is nil"
	} else if *resource.ResourceType == "OperationOutcome" {
		var oo fhir.OperationOutcome = resource.ResourceReal.(fhir.OperationOutcome)
		if len(oo.Issue) > 0 && oo.Issue[0].Diagnostics != nil {
			errcode, errstr = utils.OperationOutcomeError(target, *oo.Issue[0].Diagnostics)
			return nil, resource, errcode, errstr
		} else {
			errcode = utils.ERR_SATUSEHAT_UNDEFINED
			if target == "hapi" {
				errcode = utils.ERR_HAPI_UNDEFINED
			}
			return nil, resource, errcode, fmt.Sprintf("response OperationOutcome dari %s-%s: %s \n Detail: %s", target, env, helper.ToLogJSON(oo), helper.ToLogJSON(resource))
		}
	}

	return resource, resource, 0, ""

}

/*
*
target disini adalah forwad
*/
func (s *ProxyService) forwardRequest(method string, url string, env string, target string, auth string, body []byte) {
	logger.Info("===== Forward Request di Background =====", "Method", method, "Url", url, "Env", env, "Target", target)
	var bodyResource *types.BaseResource
	err := json.Unmarshal(body, &bodyResource)

	if err != nil {

	}
	patientId := s.getFHIRPatientReference(bodyResource)

	go func() {
		var errcode int
		var errstr string

		req, err := http.NewRequest(method, url, strings.NewReader(string(body)))
		if err != nil {
			s.logForwardError(target, url, env, "gagal membuat request", err)
			return
		}

		// Set Request Headers
		req.Header.Add("Authorization", auth)
		req.Header.Add("content-type", "application/json")

		if target == "hapi" {
			req.Header.Add(s.Config.FhirSource.Header, "local-first")
		}

		response, err := s.httpClient(&env).Do(req)
		if err != nil {
			errcode, errstr = utils.HttpError(target, req, nil, err)
			logger.Error("Forward Request Connection Error", "To", url, "env", env, "Err", errstr, "code", errcode)

			s.saveErrorTransaction(*bodyResource.Id, "forward", *bodyResource.ResourceType, env, url, *patientId, body, errstr)
			return
		}

		defer response.Body.Close()

		bodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			s.logForwardError(target, url, env, "gagal membaca response body", err)
			return
		}
		base := map[string]any{}
		err1 := json.Unmarshal(bodyBytes, &base)

		if err1 != nil {
			logger.Error("Error Marshal Response ", err1)
		}

		logger.InfoJson("Response Raw Forward Request", base)

		if response.StatusCode != fiber.StatusOK && response.StatusCode != fiber.StatusCreated {
			_, msg := utils.HttpError(target, req, response, nil)
			s.saveErrorTransaction(*bodyResource.Id, "forward", *bodyResource.ResourceType, env, url, *patientId, body, msg)
		}

		resource, err := processor.UnmarshalResource(bodyBytes, nil)
		if err != nil {
			errcode = utils.ERR_SATUSEHAT_FORMAT
			if target == "hapi" {
				errcode = utils.ERR_HAPI_FORMAT
			}
			logger.Error("Forward Request Parsing Error", "To", url, "Err", err.Error(), "code", errcode)
			return
		}

		// Logic check for OperationOutcome
		if resource.ResourceType == nil {
			logger.Error("Forward Request", "To", url, "Err", "resourceType is nil")
		} else if *resource.ResourceType == "OperationOutcome" {
			// Handle specific FHIR Error types
			if oo, ok := resource.ResourceReal.(fhir.OperationOutcome); ok {
				if len(oo.Issue) > 0 && oo.Issue[0].Diagnostics != nil {
					errcode, errstr = utils.OperationOutcomeError(target, *oo.Issue[0].Diagnostics)
					logger.Error("Forward Request Outcome", "To", url, "Err", errstr, "code", errcode)
				}
			}
		} else {
			logger.InfoJson("Response Forward Success "+url, resource)
		}
	}()
}

func (s *ProxyService) sendToKafka(transactionId *string, env string, auth string, input fhir.Bundle, satusehatResource *types.BaseResource, rawSatusehat any, resourceType string, resourceId *string) {
	logger.InfoJson("SEND TO KAFKA ====>", satusehatResource)
	logger.Info("TransactionId", transactionId)
	logger.InfoJson("Bundle Entry", input)

	go func() {
		rawResource, err := json.Marshal(satusehatResource)
		if err != nil {
			return
		}

		method := fhir.HTTPVerbPOST
		endpoint := resourceType
		if resourceId != nil {
			method = fhir.HTTPVerbPUT
			endpoint = endpoint + "/" + *resourceId
		}

		// buat object types.MediatorPack untuk diproses lewat h.fhirService.ProcessForDatabase()
		var pack types.PackMediator
		token := strings.ReplaceAll(auth, "Bearer ", "")
		if resourceType != "Bundle" {
			pack = types.PackMediator{
				TransactionID: transactionId,
				Env:           &env,
				Authorization: &token,
				Method:        helper.StringPtr("PUT"),
				Patient:       s.getFHIRPatientReference(satusehatResource),
				Input: &fhir.Bundle{
					Entry: []fhir.BundleEntry{
						{
							Resource: rawResource,
							Request: &fhir.BundleEntryRequest{
								Method: method,
								Url:    endpoint,
							},
						},
					},
				},
				Response: []types.BundleEntryResponse{
					{
						Id:           satusehatResource.Id,
						ResourceType: satusehatResource.ResourceType,
					},
				},
			}

		} else {
			pack = types.PackMediator{
				TransactionID: transactionId,
				Env:           &env,
				Authorization: &token,
				Method:        helper.StringPtr("PUT"),
			}

			var response []types.BundleEntryResponse
			var responseBundle types.Bundle

			// Cari id patient dari payload bukan dari response satusehat, karena response untuk resource bundle adalah array of entry
			pack.Patient = s.getFHIRPatientReferenceFromBundleEntry(input.Entry)

			satusehatResponseByte, err := json.Marshal(rawSatusehat)

			if err != nil {
				logger.ErrorJson("Gagal marshall Response Satusehat", err)
				return
			}

			if err := json.Unmarshal(satusehatResponseByte, &responseBundle); err != nil {
				logger.ErrorJson("Gagal Unmarshall Response Bundle Satusehat", err)
				return
			}
			response = make([]types.BundleEntryResponse, len(responseBundle.Entry))

			for i, res := range responseBundle.Entry {
				var resourceType string
				var resourceId string
				response[i].BundleEntryResponse = *res.Response

				segments := strings.Split(*res.Response.Location, "/")
				for i, segment := range segments {
					if segment == "v1" && len(segments) > i+2 {
						resourceType = segments[i+1]
						resourceId = segments[i+2]
						break
					}
				}

				response[i].ResourceType = &resourceType
				response[i].Id = &resourceId
			}
			pack.Input = &input
			pack.Response = response
		}

		// err = s.kafka.SendJSONMessage(c.Context(), s.Config.Kafka.Topics[0], "", pack)
		// if err != nil {
		// 	logger.Debug("Error Kirim Kafka Message", err)
		// }

		// logger.InfoJson("Berhasil Kirim Kafka Message", pack)

		// Gunakan proxy untuk mengirim pesan kafka
		ur, err := url.Parse(s.Config.Hapi.DevelopmentURL)
		if err != nil {
			logger.ErrorJson("Gagal Parse Url", err)
			return
		}
		kafkaProxyUrl := ur.Hostname() + "/kafka/" + env
		var errcode int
		var errstr string
		body, err := json.Marshal(pack)
		if err != nil {

		}
		req, err := http.NewRequest("POST", kafkaProxyUrl, strings.NewReader(string(body)))
		if err != nil {
			s.logForwardError("kafka", kafkaProxyUrl, env, "gagal membuat request", err)
			return
		}

		// Set Request Headers
		req.Header.Add("Authorization", auth)
		req.Header.Add("content-type", "application/json")
		if transactionId != nil {
			req.Header.Add("x-request-id", *transactionId)
		}

		response, err := s.httpClient(&env).Do(req)
		if err != nil {
			errcode, errstr = utils.HttpError("kafka", req, nil, err)
			logger.Error("Forward Request Connection Error", "To", kafkaProxyUrl, "env", env, "Err", errstr, "code", errcode)

			// simpan transaction ke database untuk dikirim ulang melalui cronjob
			s.saveErrorTransaction(*transactionId, "kafka", resourceType, env, kafkaProxyUrl, *pack.Patient, body, errstr)
			return
		}

		defer response.Body.Close()

		bodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			s.logForwardError("kafka", kafkaProxyUrl, env, "gagal membaca response body", err)
			return
		}
		base := map[string]any{}
		err1 := json.Unmarshal(bodyBytes, &base)

		if err1 != nil {
			logger.Error("Error Marshal Response ", err1)
		}

		if response.StatusCode != fiber.StatusOK && response.StatusCode != fiber.StatusCreated {
			_, msg := utils.HttpError("kafka", req, response, nil)
			s.saveErrorTransaction(*transactionId, "kafka", resourceType, env, kafkaProxyUrl, *pack.Patient, body, msg)
		}

		logger.InfoJson("Response Raw Forward Request", base)
	}()
}

func (s *ProxyService) logForwardError(target, url, env, msg string, err error) {
	errcode := utils.ERR_SATUSEHAT_FORMAT
	if target == "hapi" {
		errcode = utils.ERR_HAPI_FORMAT
	}
	logger.Error("Forward Request", "To", url, "env", env, "Msg", msg, "Err", err.Error(), "code", errcode)
}

func (h *ProxyService) httpClient(env *string) *http.Client {
	if *env == "prod" {
		return h.ProdHttpClient
	}
	return h.DevHttpClient
}

func createHttpClient(proxyUrl string) *http.Client {
	if proxyUrl == "" {
		if os.Getenv("HTTP_PROXY") != "" {
			proxyUrl = os.Getenv("HTTP_PROXY")
		} else if os.Getenv("http_proxy") != "" {
			proxyUrl = os.Getenv("http_proxy")
		}
	}
	var httpClient *http.Client
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	if proxyUrl != "" {
		proxyUrl, err := url.Parse(proxyUrl)
		if err != nil {
			logger.Fatal("Proxy URL Development", "error", err)
		}
		tr.Proxy = http.ProxyURL(proxyUrl)
		logger.Debug("HTTP Client", "Proxy", proxyUrl)
	}

	httpClient = &http.Client{
		Transport: tr,
		Timeout:   60 * time.Second, // Timeout dinaikkan untuk mengakomodasi beberapa request
	}

	return httpClient
}

func (s *ProxyService) GetUrl(env string, resourceType string, ctx *fiber.Ctx) (string, string, string) {
	// Get Priority Header Value
	priority := s.Config.FhirSource.Priority
	header := ctx.Get(s.Config.FhirSource.Header, "")

	satusehatUrl := ""
	localUrl := ""
	if header != "" {
		priority = header
	}

	if env == "dev" {
		satusehatUrl = s.Config.Development.BaseURL + "/" + resourceType
		localUrl = s.Config.Hapi.DevelopmentURL + "/" + resourceType
	} else {
		satusehatUrl = s.Config.Production.BaseURL + "/" + resourceType
		localUrl = s.Config.Hapi.ProductionURL + "/" + resourceType
	}

	switch priority {
	case "local-first":
		return localUrl, satusehatUrl, priority
	case "satusehat-first":
		return satusehatUrl, localUrl, priority
	default:
		return "", "", priority
	}
}

func (s *ProxyService) GetTarget(priority string) (string, string) {
	target := "satusehat"
	forward := "hapi"
	if priority == "local-first" {
		target = "hapi"
		forward = "satusehat"
	}

	return target, forward
}

func (s *ProxyService) getFHIRPatientReference(satusehatResource *types.BaseResource) *string {
	// satusehatResource.ResourceReal berisi object Resource FHIR
	if satusehatResource.ResourceReal == nil {
		return nil
	}

	// Gunakan reflect untuk mencari field dengan nama 'subject' atau 'patient'
	val := reflect.ValueOf(satusehatResource.ResourceReal)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return nil
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldName := strings.ToLower(field.Name)

		// Cari field bernama 'subject' atau 'patient'
		if fieldName == "subject" || fieldName == "patient" {
			fieldValue := val.Field(i)

			// Pastikan field bertipe fhir.Reference atau *fhir.Reference
			if fieldValue.Kind() == reflect.Ptr {
				if fieldValue.IsNil() {
					continue
				}
				fieldValue = fieldValue.Elem()
			}

			// Cek apakah field bertipe fhir.Reference
			if fieldValue.Type() == reflect.TypeOf(fhir.Reference{}) {
				// Ambil nilai dari field .Reference
				refField := fieldValue.FieldByName("Reference")
				if refField.IsValid() && refField.Kind() == reflect.Ptr && !refField.IsNil() {
					refStr := refField.Interface().(*string)
					if refStr != nil {
						// Hilangkan prefix "Patient/"
						patientID := strings.ReplaceAll(*refStr, "Patient/", "")
						return &patientID
					}
				}
			}
		}
	}

	return nil
}

func (s *ProxyService) getFHIRPatientReferenceFromBundleEntry(bundleEntry []fhir.BundleEntry) *string {
	if len(bundleEntry) == 0 {
		return nil
	}
	var patientId *string
	for _, entry := range bundleEntry {
		var baseResource *types.BaseResource
		entryBytes, err := json.Marshal(entry.Resource)
		if err != nil {
			continue
		}
		baseResource, err = processor.UnmarshalResource(entryBytes, nil)
		// err = json.Unmarshal(entryBytes, &baseResource)

		if err != nil {
			logger.Error("Gagal Unmarshal Entry", err)
			continue
		}

		patientId = s.getFHIRPatientReference(baseResource)

		if patientId != nil {
			break
		}
	}
	return patientId
}

func (s *ProxyService) saveErrorTransaction(id string, forwardType string, resourceType string, env string, url string, patientId string, payload []byte, errstr string) {
	transaction := types2.TransactionError{
		ID:           id,
		Type:         forwardType,
		ResourceType: resourceType,
		Env:          env,
		Url:          url,
		PatientId:    patientId,
		Payload:      payload,
		ErrorMessage: errstr,
		Status:       "PENDING",
		RetryCount:   0,
	}

	err := s.Repository.SaveTransactionError(transaction.ToEntity())

	if err != nil {
		logger.ErrorJson("Gagal Simpan Transaction Error", err)
	}
}
