package service

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/nersus15/mini-proxy/mod-proxy/config"
	"github.com/nersus15/mini-proxy/mod-proxy/helper/types"
	"github.com/nersus15/mini-proxy/mod-proxy/helper/utils"
	"github.com/nersus15/mini-proxy/mod-proxy/repository"
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
}

func NewProxyService(wctx *core.AppContext, cfg *config.ModuleConfig, repository *repository.ProxyRepository) *ProxyService {
	return &ProxyService{
		Context:        wctx,
		Config:         cfg,
		Repository:     repository,
		DevHttpClient:  utils.CreateHttpClient(cfg.Ildki.HttpProxy),
		ProdHttpClient: utils.CreateHttpClient(cfg.Ildki.HttpProxy),
	}
}

func (s *ProxyService) GetResource(env string, resourceType string, resid string, params map[string]string, ctx *fiber.Ctx) (*types.BaseResource, any, int, string) {
	auth := ctx.Get("Authorization")
	mainUrl, _, priority := utils.GetUrl(s.Config, env, resourceType, ctx)
	target, _ := utils.GetTarget(priority)

	if utils.InvalidURL(mainUrl) {
		return nil, nil, 400, "priority must be 'local-first' or 'satusehat-first'"
	}

	if resid != "" {
		mainUrl = fmt.Sprintf("%s/%s", mainUrl, resid)
	}

	if len(params) > 0 {
		mainUrl = fmt.Sprintf("%s?%s", mainUrl, utils.QueryParams(params))
	}

	logger.Info("Send GET Request Resource ==> ", "ENV", env, "Resource Type", resourceType, "Resource Id", resid, "Params: "+helper.ToLogJSON(params), "To", mainUrl)
	resource, raw, _, _, errcode, errstr := utils.SendRequest(s.httpClient(&env), s.Config, "GET", mainUrl, env, target, auth, nil)
	return resource, raw, errcode, errstr
}

func (s *ProxyService) PostResource(env string, resourceType string, ctx *fiber.Ctx) (*types.BaseResource, any, int, string) {
	auth := ctx.Get("Authorization")
	noForward := []string{"Patient", "Location", "Organization"}
	mainUrl, forwardUrl, priority := utils.GetUrl(s.Config, env, resourceType, ctx)
	target, _ := utils.GetTarget(priority)

	if utils.InvalidURL(mainUrl) || utils.InvalidURL(forwardUrl) {
		return nil, nil, 400, "priority must be 'local-first' or 'satusehat-first'"
	}

	logger.Info("Send POST Request Resource ==> ", "ENV", env, "Resource Type", resourceType, "To", mainUrl, "Forward", forwardUrl)
	resource, raw, _, httpResponse, errcode, errstr := utils.SendRequest(s.httpClient(&env), s.Config, "POST", mainUrl, env, target, auth, ctx.Body())

	if target == "ildki" {
		return resource, raw, 0, ""
	}

	if errcode == 0 && errstr == "" && !slices.Contains(noForward, resourceType) {
		if newBody, err := utils.BuildForwardRequest(ctx, mainUrl, raw, httpResponse); err != nil {
			logger.Error("buildForwardRequest", "err", err)
		} else {
			hostname := utils.GetHostName(s.Config.Ildki.ProductionURL)
			forwardUrl = fmt.Sprintf("https://%s/api/%s/backup", hostname, env)
			s.forwardRequestToIldki("POST", forwardUrl, env, auth, resourceType, newBody)
		}
	}

	return resource, raw, errcode, errstr
}

func (s *ProxyService) PutResource(env string, resourceType string, id string, ctx *fiber.Ctx) (*types.BaseResource, any, int, string) {
	auth := ctx.Get("Authorization")
	mainUrl, forwardUrl, priority := utils.GetUrl(s.Config, env, resourceType, ctx)
	target, _ := utils.GetTarget(priority)

	if utils.InvalidURL(mainUrl) || utils.InvalidURL(forwardUrl) {
		return nil, nil, 400, "priority must be 'local-first' or 'satusehat-first'"
	}

	mainUrl = fmt.Sprintf("%s/%s", mainUrl, id)
	forwardUrl = fmt.Sprintf("%s/%s", forwardUrl, id)
	logger.Info("Send PUT Request ==> ", "ENV", env, "Resource Type", resourceType, "To", mainUrl, "Forward", forwardUrl)

	resource, raw, _, httpResponse, errcode, errstr := utils.SendRequest(s.httpClient(&env), s.Config, "PUT", mainUrl, env, target, auth, ctx.Body())
	if target == "ildki" {
		return resource, raw, 0, ""
	}

	if errcode == 0 && errstr == "" {
		if newBody, err := utils.BuildForwardRequest(ctx, mainUrl, raw, httpResponse); err != nil {
			logger.Error("buildForwardRequest", "err", err)
		} else {
			hostname := utils.GetHostName(s.Config.Ildki.ProductionURL)
			forwardUrl = fmt.Sprintf("https://%s/api/%s/backup", hostname, env)
			s.forwardRequestToIldki("POST", forwardUrl, env, auth, resourceType, newBody)
		}
	}

	return resource, raw, errcode, errstr
}

func (s *ProxyService) PatchResource(env string, resourceType string, id string, ctx *fiber.Ctx) (*types.BaseResource, any, int, string) {
	auth := ctx.Get("Authorization")
	mainUrl, forwardUrl, priority := utils.GetUrl(s.Config, env, resourceType, ctx)
	target, _ := utils.GetTarget(priority)

	if utils.InvalidURL(mainUrl) || utils.InvalidURL(forwardUrl) {
		return nil, nil, 400, "priority must be 'local-first' or 'satusehat-first'"
	}

	mainUrl = fmt.Sprintf("%s/%s", mainUrl, id)
	forwardUrl = fmt.Sprintf("%s/%s", forwardUrl, id)
	logger.Info("Send PATCH Request ==> ", "ENV", env, "Resource Type", resourceType, "To", mainUrl, "Forward", forwardUrl)

	resource, raw, _, httpResponse, errcode, errstr := utils.SendRequest(s.httpClient(&env), s.Config, "PATCH", mainUrl, env, target, auth, ctx.Body())
	if target == "ildki" {
		return resource, raw, errcode, errstr
	}

	if errcode == 0 && errstr == "" {
		if newBody, err := utils.BuildForwardRequest(ctx, mainUrl, raw, httpResponse); err != nil {
			logger.Error("buildForwardRequest", "err", err)
		} else {
			hostname := utils.GetHostName(s.Config.Ildki.ProductionURL)
			forwardUrl = fmt.Sprintf("https://%s/api/%s/backup", hostname, env)
			s.forwardRequestToIldki("POST", forwardUrl, env, auth, resourceType, newBody)
		}
	}

	return resource, raw, errcode, errstr
}

func (s *ProxyService) GenerateToken(env string, target string, clientId string, clientSecret string) (types.SatuSehatTokenResponse, int, string) {
	endpoint := utils.ResolveTokenEndpoint(env, target, s.Config)
	logger.Info("Request Token", "ENV", env, "Target", target, "Endpoint", endpoint)

	data := url.Values{}
	data.Set("client_id", clientId)
	data.Set("client_secret", clientSecret)
	requestPayload := strings.NewReader(data.Encode())

	// Get Query Params
	params := map[string]string{
		"grant_type": "client_credentials",
	}
	qp := utils.QueryParams(params)
	endpoint = fmt.Sprintf("%s?%s", endpoint, qp)

	req, err := http.NewRequest("POST", endpoint, requestPayload)
	if err != nil {
		if target == "ildki" {
			logger.Warn("Failed creating request to ILDKI, falling back to SATUSEHAT")
			return s.GenerateToken(env, "satusehat", clientId, clientSecret)
		}
		return types.SatuSehatTokenResponse{}, utils.ERR_SATUSEHAT_FORMAT, fmt.Sprintf("gagal membuat request: %s", err)
	}

	req.Header.Add("content-type", "application/x-www-form-urlencoded")
	if target == "ildki" && s.Config.Ildki.ZeroTrust {
		if errSign := utils.AddSignatureToRequest(s.Config.Ildki.Faskes, req, []byte(data.Encode())); errSign != nil {
			logger.Error("Gagal menambahkan signature pada request", "Err", errSign.Error())
			return s.GenerateToken(env, "satusehat", clientId, clientSecret)
		}
	}

	response, err := s.httpClient(&env).Do(req)
	if err != nil {
		errcode, errstr := utils.HttpError("satusehat", req, nil, err)
		logger.Error("Connection Error", "To", endpoint, "env", env, "Err", errstr, "code", errcode)
		if target == "ildki" {
			logger.Warn("Failed connecting to ILDKI, falling back to SATUSEHAT")
			return s.GenerateToken(env, "satusehat", clientId, clientSecret)
		}
		return types.SatuSehatTokenResponse{}, errcode, errstr
	}
	defer utils.CloseBody(response.Body)

	bodyBytes, _ := io.ReadAll(response.Body)
	var satsetRes types.SatuSehatTokenResponse
	if response.StatusCode != fiber.StatusOK && response.StatusCode != fiber.StatusCreated {
		if target == "ildki" {
			logger.Warn(fmt.Sprintf("HAPI returned status %d, falling back to SATUSEHAT", response.StatusCode))
			return s.GenerateToken(env, "satusehat", clientId, clientSecret)
		}
		return types.SatuSehatTokenResponse{}, response.StatusCode, fmt.Sprintf("upstream returned error status: %d body: %s", response.StatusCode, string(bodyBytes))
	}

	if err := json.Unmarshal(bodyBytes, &satsetRes); err != nil {
		if target == "ildki" {
			logger.Warn("Failed to unmarshal ILDKI response, falling back to SATUSEHAT")
			return s.GenerateToken(env, "satusehat", clientId, clientSecret)
		}
		return types.SatuSehatTokenResponse{}, utils.ERR_SATUSEHAT_FORMAT, fmt.Sprintf("gagal unmarshall Response request: %s", err)
	}

	logger.InfoJson("ClientCredential Dari "+target, satsetRes)
	if target == "satusehat" {
		s.SendCredentialToProxyIL(env, &satsetRes)
	}
	return satsetRes, 0, ""
}

func (s *ProxyService) SendCredentialToProxyIL(env string, credential *types.SatuSehatTokenResponse) {
	utils.RunBackground("SendCredentialToProxyIL", func() {
		s.sendCredentialToProxyILBackground(s.Config, s.Repository, s.httpClient(&env), env, credential)
	})
}

func (s *ProxyService) SaveCredential(env string, credential *types.SatuSehatTokenResponse) {
	logger.Info("Save User Credential In Background Process")
	utils.RunBackground("SaveCredential", func() {
		s.saveCredentialBackground(s.Repository, env, credential)
	})
}

func (s *ProxyService) forwardRequestToIldki(method string, urlValue string, env string, auth string, resourceType string, body []byte) {
	utils.RunBackground("forwardRequestToIldki", func() {
		utils.ForwardRequestToIldkiBackground(s.Config, s.httpClient(&env), s.Repository, method, urlValue, env, auth, resourceType, body)
	})
}

func (s *ProxyService) sendCredentialToProxyILBackground(cfg *config.ModuleConfig, repo *repository.ProxyRepository, client *http.Client, env string, credential *types.SatuSehatTokenResponse) {
	logger.Info("Send Credential To ILDKI In Background Process")

	body, err := json.Marshal(credential)
	if err != nil {
		logger.DebugJson("Failed To Marshal Credential", err)
		return
	}

	if _, err := repo.DeleteOldCredTransactions(credential.AccessToken); err != nil {
		logger.ErrorJson("Gagal Hapus Transaction", err)
	}

	ur, err := url.Parse(cfg.Ildki.DevelopmentURL)
	if err != nil {
		logger.Error("Invalid ILDKI URL", "err", err)
		return
	}

	endpoint := fmt.Sprintf("https://%s/api/%s/credential/sync", ur.Hostname(), env)
	logger.Info("Sending Credential To ILDKI", "endpoint", endpoint)

	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(body))
	if err != nil {
		logger.DebugJson("Failed To Create Request", err)
		utils.SaveErrorTransactionBackground(repo, types.TransactionError{
			ID:           credential.AccessToken,
			Type:         "forward",
			ResourceType: "credential",
			Env:          env,
			Url:          endpoint,
			PatientId:    "",
			Payload:      body,
			ErrorMessage: err.Error(),
			Status:       "PENDING",
			RetryCount:   0,
		})
		return
	}

	if cfg.Ildki.ZeroTrust {
		if errSign := utils.AddSignatureToRequest(cfg.Ildki.Faskes, req, body); errSign != nil {
			logger.Error("Gagal menambahkan signature pada request", "Err", errSign.Error())
			utils.SaveErrorTransactionBackground(repo, types.TransactionError{
				ID:           credential.AccessToken,
				Type:         "forward",
				ResourceType: "credential",
				Env:          env,
				Url:          endpoint,
				PatientId:    "",
				Payload:      body,
				ErrorMessage: errSign.Error(),
				Status:       "PENDING",
				RetryCount:   0,
			})
			return
		}
	}
	req.Header.Add("Content-Type", "application/json")

	utils.ExecuteBackgroundRequest(client, req, func(response *http.Response) {
		if response.StatusCode == fiber.StatusOK || response.StatusCode == fiber.StatusCreated {
			logger.Info("Successfully Sent Credential To ILDKI")
			return
		}

		utils.SaveErrorTransactionBackground(repo, types.TransactionError{
			ID:           credential.AccessToken,
			Type:         "forward",
			ResourceType: "credential",
			Env:          env,
			Url:          endpoint,
			PatientId:    "",
			Payload:      body,
			ErrorMessage: "Response Error",
			Status:       "PENDING",
			RetryCount:   0,
		})
		logger.ErrorJson("Failed To Sent Credential To ILDKI", response)
	}, func(err error) {
		logger.DebugJson("Request Error", err)
		utils.SaveErrorTransactionBackground(repo, types.TransactionError{
			ID:           credential.AccessToken,
			Type:         "forward",
			ResourceType: "credential",
			Env:          env,
			Url:          endpoint,
			PatientId:    "",
			Payload:      body,
			ErrorMessage: err.Error(),
			Status:       "PENDING",
			RetryCount:   0,
		})
	})
}

func (s *ProxyService) saveCredentialBackground(repo *repository.ProxyRepository, env string, credential *types.SatuSehatTokenResponse) {
	entityData := credential.ToEntity()
	entityData.Env = env

	issuedAtUnixMilli := time.Now().UnixMilli()
	if credential.IssuedAt != "" {
		if parsed, err := strconv.ParseInt(credential.IssuedAt, 10, 64); err == nil {
			issuedAtUnixMilli = parsed
		} else {
			logger.Error("Failed to parse issued_at, fallback to time.Now()", "error", err)
		}
	}

	expiresInSeconds, err := strconv.ParseInt(credential.ExpiresIn, 10, 64)
	if err != nil {
		logger.Error("Failed to parse expires_in", "error", err)
		return
	}

	issuedAtTime := time.UnixMilli(issuedAtUnixMilli)
	entityData.ExpiredAt = issuedAtTime.Add(time.Duration(expiresInSeconds) * time.Second)

	if err := repo.SaveClientCredentials(entityData); err != nil {
		logger.ErrorJson("Error saving client credential", err)
		return
	}

	logger.Info("Successfully saved client credential", "client_id", entityData.ClientID, "env", env)
}

func (h *ProxyService) httpClient(env *string) *http.Client {
	if *env == "prod" {
		return h.ProdHttpClient
	}
	return h.DevHttpClient
}
