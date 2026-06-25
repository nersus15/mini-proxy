//go:build ignore
// +build ignore

package service

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/nersus15/mini-proxy/mod-proxy/config"
	"github.com/nersus15/mini-proxy/mod-proxy/helper/processor"
	"github.com/nersus15/mini-proxy/mod-proxy/helper/types"
	"github.com/nersus15/mini-proxy/mod-proxy/helper/utils"
	utils2 "github.com/nersus15/mini-proxy/mod-proxy/helper/utils"
	"github.com/nersus15/mini-proxy/mod-proxy/repository"
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
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
		DevHttpClient:  createHttpClient(cfg.Ildki.HttpProxy),
		ProdHttpClient: createHttpClient(cfg.Ildki.HttpProxy),
		kafka:          kafkaProducer,
	}
}

func (s *ProxyService) GetResource(env string, resourceType string, resid string, params map[string]string, ctx *fiber.Ctx) (*types.BaseResource, any, int, string) {
	auth := ctx.Get("Authorization")
	mainUrl, _, priority := s.GetUrl(env, resourceType, ctx)
	target, _ := s.GetTarget(priority)

	if mainUrl == "" || mainUrl == "/" {
		return nil, nil, 400, "priority must be 'local-first' or 'satusehat-first'"
	}
	if resid != "" {
		mainUrl = mainUrl + "/" + resid
	}

	if len(params) > 0 {
		mainUrl = fmt.Sprintf("%s?%s", mainUrl, queryParams(params))
	}

	logger.Info("Send GET Request Resource ==> ", "ENV", env, "Resource Type", resourceType, "Resource Id", resid, "Params: "+helper.ToLogJSON(params), "To", mainUrl)

	res, raw, _, _, errcode, errstr := s.sendRequest("GET", mainUrl, env, target, auth, nil)
	return res, raw, errcode, errstr
}

func (s *ProxyService) PostResource(env string, resourceType string, ctx *fiber.Ctx) (*types.BaseResource, any, int, string) {
	auth := ctx.Get("Authorization")
	noForward := []string{
		"Patient",
		"Location",
		"Organization",
	}
	mainUrl, forwardUrl, priority := s.GetUrl(env, resourceType, ctx)
	target, _ := s.GetTarget(priority)

	if (mainUrl == "" || mainUrl == "/") || (forwardUrl == "" || forwardUrl == "/") {
		return nil, nil, 400, "priority must be 'local-first' or 'satusehat-first'"
	}

	logger.Info("Send POST Request Resource ==> ", "ENV", env, "Resource Type", resourceType, "To", mainUrl, "Forward", forwardUrl)

	resource, raw, _, httpResponse, errcode, errstr := s.sendRequest("POST", mainUrl, env, target, auth, ctx.Body())

	if target == "ildki" {
		// Jika local-first tidak perlu forwad request ke SatuSehat melalui mini proxy, karena sudah di handle oleh ILDKI
		return resource, raw, 0, ""
	}
	if errcode == 0 && errstr == "" && !slices.Contains(noForward, resourceType) {
		var requestBodyJson interface{}
		if errBody := json.Unmarshal(ctx.Body(), &requestBodyJson); errBody != nil {
			// Jika bukan json valid, biarkan fallback ke string mentah
			requestBodyJson = string(ctx.Body())
		}

		temp := forwardPayload(ctx.GetReqHeaders(), httpResponse, mainUrl, requestBodyJson, raw)
		newBody, err := json.Marshal(temp)

		if err != nil {
			logger.Error("json.Marshal", helper.ToLogJSON(err))
		} else {
			// forward ke internal hapi via api (bukan kafka)
			hostname := getHostName(s.Config.Ildki.ProductionURL)

			forwardUrl = fmt.Sprintf("https://%s/%s/backup", hostname, env)
			s.forwardRequestToIldki("POST", forwardUrl, env, auth, resourceType, newBody)

		}
	}

	return resource, raw, errcode, errstr
}

func (s *ProxyService) PutResource(env string, resourceType string, id string, ctx *fiber.Ctx) (*types.BaseResource, any, int, string) {
	auth := ctx.Get("Authorization")

	mainUrl, forwardUrl, priority := s.GetUrl(env, resourceType, ctx)
	target, _ := s.GetTarget(priority)

	if (mainUrl == "" || mainUrl == "/") || (forwardUrl == "" || forwardUrl == "/") {
		return nil, nil, 400, "priority must be 'local-first' or 'satusehat-first'"
	}
	mainUrl += "/" + id
	forwardUrl += "/" + id
	logger.Info("Send PUT Request ==> ", "ENV", env, "Resource Type", resourceType, "To", mainUrl, "Forward", forwardUrl)

	resource, raw, _, httpResponse, errcode, errstr := s.sendRequest("PUT", mainUrl, env, target, auth, ctx.Body())

	if target == "ildki" {
		// Jika local-first tidak perlu forwad request ke SatuSehat melalui mini proxy, karena sudah di handle oleh ILDKI
		return resource, raw, 0, ""
	}

	if errcode == 0 && errstr == "" {
		var requestBodyJson interface{}
		if errBody := json.Unmarshal(ctx.Body(), &requestBodyJson); errBody != nil {
			// Jika bukan json valid, biarkan fallback ke string mentah
			requestBodyJson = string(ctx.Body())
		}

		temp := forwardPayload(ctx.GetReqHeaders(), httpResponse, mainUrl, requestBodyJson, raw)
		newBody, err := json.Marshal(temp)

		if err != nil {
			logger.Error("json.Marshal", helper.ToLogJSON(err))
		} else {
			// forward ke internal hapi via api (bukan kafka)
			hostname := getHostName(s.Config.Ildki.ProductionURL)

			forwardUrl = fmt.Sprintf("https://%s/%s/backup", hostname, env)
			s.forwardRequestToIldki("POST", forwardUrl, env, auth, resourceType, newBody)

		}
	}

	return resource, raw, errcode, errstr
}

func (s *ProxyService) PatchResource(env string, resourceType string, id string, ctx *fiber.Ctx) (*types.BaseResource, any, int, string) {
	auth := ctx.Get("Authorization")

	mainUrl, forwardUrl, priority := s.GetUrl(env, resourceType, ctx)
	target, _ := s.GetTarget(priority)

	if (mainUrl == "" || mainUrl == "/") || (forwardUrl == "" || forwardUrl == "/") {
		return nil, nil, 400, "priority must be 'local-first' or 'satusehat-first'"
	}
	mainUrl += "/" + id
	forwardUrl += "/" + id
	logger.Info("Send PUT Request ==> ", "ENV", env, "Resource Type", resourceType, "To", mainUrl, "Forward", forwardUrl)

	resource, raw, _, httpResponse, errcode, errstr := s.sendRequest("PATCH", mainUrl, env, target, auth, ctx.Body())

	if target == "ildki" {
		// Jika local-first tidak perlu forwad request ke SatuSehat melalui mini proxy, karena sudah di handle oleh ILDKI
		return resource, raw, errcode, errstr
	}

	if errcode == 0 && errstr == "" {
		var requestBodyJson interface{}
		if errBody := json.Unmarshal(ctx.Body(), &requestBodyJson); errBody != nil {
			// Jika bukan json valid, biarkan fallback ke string mentah
			requestBodyJson = string(ctx.Body())
		}

		temp := forwardPayload(ctx.GetReqHeaders(), httpResponse, mainUrl, requestBodyJson, raw)
		newBody, err := json.Marshal(temp)

		if err != nil {
			logger.Error("json.Marshal", helper.ToLogJSON(err))
		} else {
			// forward ke internal hapi via api (bukan kafka)
			hostname := getHostName(s.Config.Ildki.ProductionURL)

			forwardUrl = fmt.Sprintf("https://%s/%s/backup", hostname, env)
			s.forwardRequestToIldki("POST", forwardUrl, env, auth, resourceType, newBody)

		}
	}

	return resource, raw, errcode, errstr

}

func (s *ProxyService) GenerateToken(env string, target string, clientId string, clientSecret string) (types.SatuSehatTokenResponse, int, string) {
	var endpoint string
	var errstr string
	var errcode int
	var satsetRes types.SatuSehatTokenResponse

	if env == "dev" {
		if target == "ildki" {
			endpoint = s.Config.Ildki.DevAuthUrl
		} else {
			endpoint = s.Config.SatSetDev.AuthURL
		}
	} else {
		if target == "ildki" {
			endpoint = s.Config.Ildki.ProductionURL
		} else {
			endpoint = s.Config.SatSetProd.AuthURL
		}
	}
	logger.Info("Request Token", "ENV", env, "Target", target, "Endpoint", endpoint)

	data := url.Values{}
	data.Set("client_id", clientId)
	data.Set("client_secret", clientSecret)
	dataEncode := data.Encode()
	payload := strings.NewReader(dataEncode)

	req, err := http.NewRequest("POST", endpoint, payload)
	if err != nil {
		errcode = utils.ERR_SATUSEHAT_FORMAT
		if target == "ildki" {
			logger.Warn("Failed creating request to ILDKI, falling back to SATUSEHAT")
			return s.GenerateToken(env, "satusehat", clientId, clientSecret)
		}
		return satsetRes, errcode, fmt.Sprintf("gagal membuat request: %s", err)
	}
	req.Header.Add("content-type", "application/x-www-form-urlencoded")

	if target == "ildki" && s.Config.Ildki.ZeroTrust {
		errSign := utils2.AddSignatureToRequest(s.Config.Ildki.Faskes, req, []byte(dataEncode))
		if errSign != nil {
			logger.Error("Gagal menambahkan signature pada request", "Err", errSign.Error())
			return s.GenerateToken(env, "satusehat", clientId, clientSecret)
		}
	}

	response, err := s.httpClient(&env).Do(req)
	if err != nil {
		errcode, errstr = utils.HttpError("satusehat", req, nil, err)
		logger.Error("Connection Error", "To", endpoint, "env", env, "Err", errstr, "code", errcode)

		if target == "ildki" {
			logger.Warn("Failed connecting to ILDKI, falling back to SATUSEHAT")
			return s.GenerateToken(env, "satusehat", clientId, clientSecret)
		}
		return satsetRes, errcode, errstr
	}
	defer response.Body.Close()

	bodyBytes, _ := io.ReadAll(response.Body)

	if response.StatusCode != fiber.StatusOK && response.StatusCode != fiber.StatusCreated {
		if target == "ildki" {
			logger.Warn(fmt.Sprintf("HAPI returned status %d, falling back to SATUSEHAT", response.StatusCode))
			return s.GenerateToken(env, "satusehat", clientId, clientSecret)
		}
		return satsetRes, response.StatusCode, fmt.Sprintf("upstream returned error status: %d body: %s", response.StatusCode, string(bodyBytes))
	}

	err = json.Unmarshal(bodyBytes, &satsetRes)
	if err != nil {
		if target == "ildki" {
			logger.Warn("Failed to unmarshal ILDKI response, falling back to SATUSEHAT")
			return s.GenerateToken(env, "satusehat", clientId, clientSecret)
		}
		return satsetRes, errcode, fmt.Sprintf("gagal unmarshall Response request: %s", err)
	}

	logger.InfoJson("ClientCredential Dari "+target, satsetRes)

	if target == "satusehat" {
		s.SendCredentialToProxyIL(env, &satsetRes)
	}
	return satsetRes, errcode, errstr
}

func (s *ProxyService) SendCredentialToProxyIL(env string, credential *types.SatuSehatTokenResponse) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error(fmt.Sprintf("Panic terjadi di background SendCredentialToProxyIL: %v", r))
			}
		}()

		logger.Info("Send Credential To ILDKI In Background Process")
		body, err := json.Marshal(credential)
		if err != nil {
			logger.DebugJson("Failed To Marshal Credential", err)
			return
		}

		// Hapus Transaksi sebelumnya (where id != credential.AccessToken)
		_, err = s.Repository.DeleteOldCredTransactions(credential.AccessToken)
		if err != nil {
			logger.ErrorJson("Gagal Hapus Transaction", err)
		}

		ur, err := url.Parse(s.Config.Ildki.DevelopmentURL)
		endpoint := fmt.Sprintf("https://%s/api/%s/credential/sync", ur.Hostname(), env)
		logger.Info(fmt.Sprintf("Sending Credential To %s", endpoint))

		req, err := http.NewRequest("POST", endpoint, bytes.NewReader(body))
		if err != nil {
			logger.DebugJson("Failed To Create Request", err)
			s.saveErrorTransaction(credential.AccessToken, "forward", "credential", env, endpoint, "", body, err.Error())
			return
		}

		if s.Config.Ildki.ZeroTrust {
			errSign := utils2.AddSignatureToRequest(s.Config.Ildki.Faskes, req, body)
			if errSign != nil {
				logger.Error("Gagal menambahkan signature pada request", "Err", errSign.Error())
				s.saveErrorTransaction(credential.AccessToken, "forward", "credential", env, endpoint, "", body, errSign.Error())
				return
			}
		}
		req.Header.Add("Content-Type", "application/json")

		response, err := s.httpClient(&env).Do(req)
		if err != nil {
			logger.DebugJson("Request Error", err)
			s.saveErrorTransaction(credential.AccessToken, "forward", "credential", env, endpoint, "", body, err.Error())
			return
		}
		defer response.Body.Close()

		if response.StatusCode == fiber.StatusOK || response.StatusCode == fiber.StatusCreated {
			logger.Info("Successfully To Sent Credential To ILDKI")
		} else {
			s.saveErrorTransaction(credential.AccessToken, "forward", "credential", env, endpoint, "", body, "Response Error")
			logger.ErrorJson("Failed To Sent Credential To ILDKI", response)
		}

	}()
}

func (s *ProxyService) SaveCredential(env string, credential *types.SatuSehatTokenResponse) {
	logger.Info("Save User Credential In Background Process")
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error(fmt.Sprintf("Panic terjadi di SaveCredential: %v", r))
			}
		}()

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

func (s *ProxyService) sendRequest(method string, url string, env string, target string, auth string, body []byte) (*types.BaseResource, any, *http.Request, *http.Response, int, string) {
	var errcode int
	var errstr string
	logger.Info("===== Send Request =====", "Method", method, "Url", url, "Env", env, "Target", target)

	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		errcode = utils.ERR_SATUSEHAT_FORMAT
		if target == "ildki" {
			errcode = utils.ERR_HAPI_FORMAT
		}

		return nil, nil, req, nil, errcode, fmt.Sprintf("gagal membuat request: %s", err)
	}
	// Set Request Header
	req.Header.Add("Authorization", auth)
	req.Header.Add("content-type", "application/json")

	if target == "ildki" {
		req.Header.Add(s.Config.FhirSource.Header, "local-first")
		if s.Config.Ildki.ZeroTrust {
			errSign := utils2.AddSignatureToRequest(s.Config.Ildki.Faskes, req, body)
			if errSign != nil {
				logger.Error("Gagal menambahkan signature pada request", "Err", errSign.Error())
				return nil, nil, req, nil, 500, errSign.Error()
			}
		}
	}

	response, err := s.httpClient(&env).Do(req)
	if err != nil {
		errcode, errstr = utils.HttpError(target, req, response, nil)

		return nil, nil, req, nil, errcode, errstr
	}
	defer response.Body.Close()

	bodyBytes, _ := io.ReadAll(response.Body)
	var rawResponse any
	err1 := json.Unmarshal(bodyBytes, &rawResponse)

	resource, err := processor.UnmarshalResource(bodyBytes, nil)
	logger.InfoJson("Response "+url, resource)
	logger.InfoJson("Raw Response "+url, rawResponse)

	if err != nil {
		errcode = utils.ERR_SATUSEHAT_FORMAT
		if target == "ildki" {
			errcode = utils.ERR_HAPI_FORMAT
		}
		return resource, rawResponse, req, response, errcode, fmt.Sprintf("gagal mem-parsing JSON: %s", err)
	}

	if err1 != nil {
		rawResponse = resource
		logger.InfoJson("Gagal Unmarshall Body", err1)
	}

	if resource.ResourceType == nil {
		errcode = utils.ERR_SATUSEHAT_FORMAT
		if target == "ildki" {
			errcode = utils.ERR_HAPI_FORMAT
		}
		return resource, rawResponse, req, response, errcode, "resourceType is Nul"
	} else if *resource.ResourceType == "OperationOutcome" {
		var oo fhir.OperationOutcome = resource.ResourceReal.(fhir.OperationOutcome)
		if len(oo.Issue) > 0 && oo.Issue[0].Diagnostics != nil {
			errcode, errstr = utils.OperationOutcomeError(target, *oo.Issue[0].Diagnostics)
			return resource, resource, req, response, errcode, errstr
		} else {
			errcode = utils.ERR_SATUSEHAT_UNDEFINED
			if target == "ildki" {
				errcode = utils.ERR_HAPI_UNDEFINED
			}
			return resource, rawResponse, req, response, errcode, fmt.Sprintf("response OperationOutcome dari %s-%s: %s \n Detail: %s", target, env, helper.ToLogJSON(oo), helper.ToLogJSON(resource))
		}
	}

	return resource, rawResponse, req, response, 0, ""

}

func (s *ProxyService) forwardRequestToIldki(method string, url string, env string, auth string, resourceType string, body []byte) {
	go func() {
		logger.Info("===== Forward Request di Background =====", "Method", method, "Url", url, "Env", env)

		defer func() {
			if r := recover(); r != nil {
				logger.Error(fmt.Sprintf("Panic terjadi di background forwardRequestToIldki: %v", r))
			}
		}()

		target := "ildki"
		var transactionId string
		var patientId string

		if resourceType == "" {
			resourceType = "Bundle"
		}

		if resourceType != "Bundle" {
			transactionId, patientId = s.ResourceTransactionDetail(resourceType, body)
		} else {
			transactionId, patientId = s.BundleTransactionDetail(resourceType, body)
		}

		var errcode int
		var errstr string

		req, err := http.NewRequest(method, url, bytes.NewReader(body))
		if err != nil {
			s.logForwardError(url, env, "gagal membuat request", err)
			return
		}

		// Set Request Headers
		req.Header.Add("Authorization", auth)
		req.Header.Add("content-type", "application/json")
		req.Header.Add(s.Config.FhirSource.Header, "backup-satusehat")

		if s.Config.Ildki.ZeroTrust {
			errSign := utils2.AddSignatureToRequest(s.Config.Ildki.Faskes, req, body)
			if errSign != nil {
				logger.Error("Gagal menambahkan signature pada request", "Err", errSign.Error())
				s.saveErrorTransaction(transactionId, "forward", resourceType, env, url, patientId, body, errSign.Error())
				return
			}
		}

		response, err := s.httpClient(&env).Do(req)
		if err != nil {
			errcode, errstr = utils.HttpError(target, req, nil, err)
			logger.Error("Forward Request Connection Error", "To", url, "env", env, "Err", errstr, "code", errcode)

			s.saveErrorTransaction(transactionId, "forward", resourceType, env, url, patientId, body, errstr)
			return
		}

		defer response.Body.Close()

		if response.StatusCode != fiber.StatusOK && response.StatusCode != fiber.StatusCreated {
			_, msg := utils.HttpError(target, req, response, nil)
			s.saveErrorTransaction(transactionId, "forward", resourceType, env, url, patientId, body, msg)
		}

		bodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			s.logForwardError(url, env, "gagal membaca response body", err)
			return
		}
		base := map[string]any{}
		err1 := json.Unmarshal(bodyBytes, &base)

		if err1 != nil {
			logger.Error("Error Marshal Response ", err1)
			return
		}

		logger.InfoJson(fmt.Sprintf("Response Raw Forward Request (%s) ==> : ", url), base)
	}()
}

func (s *ProxyService) ResourceTransactionDetail(resourceType string, body []byte) (string, string) {
	var Payload types.BackupPayload
	var bodyResource *types.BaseResource
	var transactionId *string
	var patientId *string

	err := json.Unmarshal(body, &Payload)

	if err != nil {
		logger.Error("Gagal Unmarshal BackupPayload", "Err", err.Error())
		return "", ""
	}

	logger.InfoJson("Payload", Payload)

	dataBytes, err := json.Marshal(Payload.Response.Data)
	if err != nil {
		logger.Error("Gagal Remarshal Response Data", "Err", err.Error())
		return "", ""
	}

	bodyResource, err = processor.UnmarshalResource(dataBytes, nil)
	if err != nil {
		logger.Error("Gagal Unmarshal ke BaseResource", "Err", err.Error())
		return "", ""
	}

	logger.InfoJson("Body Request", bodyResource)
	patientId = s.getFHIRPatientReference(bodyResource)
	transactionId = bodyResource.Id

	return *transactionId, *patientId
}

func (s *ProxyService) BundleTransactionDetail(resourceType string, body []byte) (string, string) {
	var Payload types.BackupPayload
	var Request types.AxiosRequestConfig
	var input *fhir.Bundle
	var transactionId *string
	var patientId *string

	err := json.Unmarshal(body, &Payload)

	if err != nil {
		logger.Error("Gagal Unmarshal BackupPayload", "Err", err.Error())
		return "", ""
	}

	dataBytes, err := json.Marshal(Payload.Request)
	if err != nil {
		logger.Error("Gagal Remarshal Request", "Err", err.Error())
		return "", ""
	}

	err = json.Unmarshal(dataBytes, &Request)
	if err != nil {
		logger.Error("Gagal Unmarshal ke Request", "Err", err.Error())
		return "", ""
	}

	dataBytes, err = json.Marshal(Request.Data)
	if err != nil {
		logger.Error("Gagal Remarshal Request Data", "Err", err.Error())
		return "", ""
	}

	err = json.Unmarshal(dataBytes, &input)
	if err != nil {
		logger.Error("Gagal Unmarshal ke Bundle Entry", "Err", err.Error())
		return "", ""
	}

	logger.InfoJson("Bundle Input", input.Entry)
	transactionId = helper.StringPtr(uuid.New().String())
	patientId = s.getFHIRPatientReferenceFromBundleEntry(input.Entry)

	return *transactionId, *patientId
}

func (s *ProxyService) logForwardError(url, env, msg string, err error) {
	errcode := utils.ERR_SATUSEHAT_FORMAT
	errcode = utils.ERR_HAPI_FORMAT
	logger.Error("Forward Request", "To", url, "env", env, "Msg", msg, "Err", err.Error(), "code", errcode)
}

func (h *ProxyService) httpClient(env *string) *http.Client {
	if *env == "prod" {
		return h.ProdHttpClient
	}
	return h.DevHttpClient
}

func forwardPayload(headers map[string][]string, httpResponse *http.Response, url string, requestBody interface{}, raw any) types.BackupPayload {
	statusText := http.StatusText(httpResponse.StatusCode)
	if statusText == "" {
		statusText = "OK"
	}

	expressRequestHeaders := make(map[string]string)
	for key, values := range headers {
		if len(values) > 0 {
			expressRequestHeaders[strings.ToLower(key)] = strings.Join(values, ", ")
		}
	}

	// 2. Flatten Response Headers dari SATUSEHAT
	expressResponseHeaders := make(map[string]string)
	for key, values := range httpResponse.Header {
		if len(values) > 0 {
			expressResponseHeaders[strings.ToLower(key)] = strings.Join(values, ", ")
		}
	}

	cleanRequestPayload := types.AxiosRequestConfig{
		URL:     url,
		Method:  "POST",
		Headers: expressRequestHeaders,
		Data:    requestBody,
	}

	payload := types.BackupPayload{
		Request: cleanRequestPayload,
		Response: types.AxiosResponse{
			Data:       raw,
			Status:     httpResponse.StatusCode,
			StatusText: statusText,
			Headers:    httpResponse.Header,
			Config: types.AxiosRequestConfig{
				URL:     url,
				Method:  "POST",
				Headers: expressRequestHeaders,
				Data:    requestBody,
			},
		},
	}

	return payload
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

func queryParams(params map[string]string) string {
	v := url.Values{}
	for key, value := range params {
		v.Add(key, value)
	}

	return v.Encode()
}

func getHostName(u string) string {
	ur, err := url.Parse(u)
	if err != nil {
		logger.ErrorJson("Gagal Parse Url", err)
		return ""
	}

	return ur.Hostname()
}
func (s *ProxyService) GetUrl(env string, resourceType string, ctx *fiber.Ctx) (string, string, string) {
	// Get Priority Header Value
	priority := s.Config.FhirSource.Priority
	header := ctx.Get(s.Config.FhirSource.Header, "")

	satusehatUrl := ""
	ildkiUrl := ""
	if header != "" {
		priority = header
	}

	if env == "dev" {
		satusehatUrl = s.Config.SatSetDev.BaseURL + "/" + resourceType
		ildkiUrl = s.Config.Ildki.DevelopmentURL + "/" + resourceType
	} else {
		satusehatUrl = s.Config.SatSetProd.BaseURL + "/" + resourceType
		ildkiUrl = s.Config.Ildki.ProductionURL + "/" + resourceType
	}

	switch priority {
	case "local-first":
		return ildkiUrl, satusehatUrl, priority
	case "satusehat-first":
		return satusehatUrl, ildkiUrl, priority
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
	logger.Info("Simpan Transaksi Error - Background")
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error(fmt.Sprintf("Panic terjadi di background saveErrorTransaction: %v", r))
			}
		}()

		transaction := types.TransactionError{
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
			logger.ErrorJson("Gagal Simpan Transaction", err)
		}
	}()
}
