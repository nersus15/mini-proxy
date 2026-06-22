package utils

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/nersus15/mini-proxy/mod-proxy/config"
	"github.com/nersus15/mini-proxy/mod-proxy/helper/processor"
	"github.com/nersus15/mini-proxy/mod-proxy/helper/types"
	"github.com/nersus15/mini-proxy/mod-proxy/repository"
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
	"github.com/webcore-go/webcore/app/helper"
	"github.com/webcore-go/webcore/infra/logger"
)

func ResolveTokenEndpoint(env, target string, cfg *config.ModuleConfig) string {
	if env == "dev" {
		if target == "ildki" {
			return cfg.Ildki.DevAuthUrl
		}
		return cfg.SatSetDev.AuthURL
	}

	if target == "ildki" {
		return cfg.Ildki.ProductionURL
	}
	return cfg.SatSetProd.AuthURL
}

func SendRequest(client *http.Client, cfg *config.ModuleConfig, method, urlValue, env, target, auth string, body []byte) (*types.BaseResource, any, *http.Request, *http.Response, int, string) {
	req, err := http.NewRequest(method, urlValue, bytes.NewReader(body))
	if err != nil {
		errcode := ERR_SATUSEHAT_FORMAT
		if target == "ildki" {
			errcode = ERR_HAPI_FORMAT
		}
		return nil, nil, req, nil, errcode, fmt.Sprintf("gagal membuat request: %s", err)
	}

	req.Header.Add("Authorization", auth)
	req.Header.Add("content-type", "application/json")

	if target == "ildki" {
		req.Header.Add(cfg.FhirSource.Header, "local-first")
		if cfg.Ildki.ZeroTrust {
			if errSign := AddSignatureToRequest(cfg.Ildki.Faskes, req, body); errSign != nil {
				logger.Error("Gagal menambahkan signature pada request", "Err", errSign.Error())
				return nil, nil, req, nil, 500, errSign.Error()
			}
		}
	}

	response, err := client.Do(req)
	if err != nil {
		errcode, errstr := HttpError(target, req, nil, err)
		return nil, nil, req, nil, errcode, errstr
	}
	defer CloseBody(response.Body)

	bodyBytes, _ := io.ReadAll(response.Body)
	var rawResponse any
	err1 := json.Unmarshal(bodyBytes, &rawResponse)

	resource, err := processor.UnmarshalResource(bodyBytes, nil)
	logger.InfoJson("Response "+urlValue, resource)
	logger.InfoJson("Raw Response "+urlValue, rawResponse)

	if err != nil {
		errcode := ERR_SATUSEHAT_FORMAT
		if target == "ildki" {
			errcode = ERR_HAPI_FORMAT
		}
		return resource, rawResponse, req, response, errcode, fmt.Sprintf("gagal mem-parsing JSON: %s", err)
	}

	if err1 != nil {
		rawResponse = resource
		logger.InfoJson("Gagal Unmarshall Body", err1)
	}

	if resource.ResourceType == nil {
		errcode := ERR_SATUSEHAT_FORMAT
		if target == "ildki" {
			errcode = ERR_HAPI_FORMAT
		}
		return resource, rawResponse, req, response, errcode, "resourceType is Nul"
	}

	if *resource.ResourceType == "OperationOutcome" {
		oo := resource.ResourceReal.(fhir.OperationOutcome)
		if len(oo.Issue) > 0 && oo.Issue[0].Diagnostics != nil {
			errcode, errstr := OperationOutcomeError(target, *oo.Issue[0].Diagnostics)
			return resource, resource, req, response, errcode, errstr
		}

		errcode := ERR_SATUSEHAT_UNDEFINED
		if target == "ildki" {
			errcode = ERR_HAPI_UNDEFINED
		}
		return resource, rawResponse, req, response, errcode, fmt.Sprintf("response OperationOutcome dari %s-%s: %s \n Detail: %s", target, env, helper.ToLogJSON(oo), helper.ToLogJSON(resource))
	}

	return resource, rawResponse, req, response, 0, ""
}

func BuildForwardRequest(ctx *fiber.Ctx, mainUrl string, raw any, httpResponse *http.Response) ([]byte, error) {
	var requestBodyJson interface{}
	if err := json.Unmarshal(ctx.Body(), &requestBodyJson); err != nil {
		requestBodyJson = string(ctx.Body())
	}

	payload := ForwardPayload(ctx.GetReqHeaders(), httpResponse, mainUrl, requestBodyJson, raw)
	return json.Marshal(payload)
}

func ExtractTransactionDetails(resourceType string, body []byte) (string, string) {
	if resourceType != "Bundle" {
		return ResourceTransactionDetail(resourceType, body)
	}
	return BundleTransactionDetail(resourceType, body)
}

func ResourceTransactionDetail(resourceType string, body []byte) (string, string) {
	var payload types.BackupPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		logger.Error("Gagal Unmarshal BackupPayload", "Err", err.Error())
		return "", ""
	}

	dataBytes, err := json.Marshal(payload.Response.Data)
	if err != nil {
		logger.Error("Gagal Remarshal Response Data", "Err", err.Error())
		return "", ""
	}

	bodyResource, err := processor.UnmarshalResource(dataBytes, nil)
	if err != nil {
		logger.Error("Gagal Unmarshal ke BaseResource", "Err", err.Error())
		return "", ""
	}

	logger.InfoJson("Body Request", bodyResource)
	patientId := GetFHIRPatientReference(bodyResource)
	if patientId == nil || bodyResource.Id == nil {
		return "", ""
	}
	return *bodyResource.Id, *patientId
}

func BundleTransactionDetail(resourceType string, body []byte) (string, string) {
	var payload types.BackupPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		logger.Error("Gagal Unmarshal BackupPayload", "Err", err.Error())
		return "", ""
	}

	requestBytes, err := json.Marshal(payload.Request)
	if err != nil {
		logger.Error("Gagal Remarshal Request", "Err", err.Error())
		return "", ""
	}

	var requestConfig types.AxiosRequestConfig
	if err := json.Unmarshal(requestBytes, &requestConfig); err != nil {
		logger.Error("Gagal Unmarshal ke Request", "Err", err.Error())
		return "", ""
	}

	inputBytes, err := json.Marshal(requestConfig.Data)
	if err != nil {
		logger.Error("Gagal Remarshal Request Data", "Err", err.Error())
		return "", ""
	}

	var input *fhir.Bundle
	if err := json.Unmarshal(inputBytes, &input); err != nil {
		logger.Error("Gagal Unmarshal ke Bundle Entry", "Err", err.Error())
		return "", ""
	}

	logger.InfoJson("Bundle Input", input.Entry)
	transactionId := uuid.New().String()
	patientId := GetFHIRPatientReferenceFromBundleEntry(input.Entry)
	if patientId == nil {
		return transactionId, ""
	}
	return transactionId, *patientId
}

func ForwardRequestToIldkiBackground(cfg *config.ModuleConfig, client *http.Client, repo *repository.ProxyRepository, method string, urlValue string, env string, auth string, resourceType string, body []byte) {
	logger.Info("===== Forward Request di Background =====", "Method", method, "Url", urlValue, "Env", env)

	if resourceType == "" {
		resourceType = "Bundle"
	}

	transactionId, patientId := ExtractTransactionDetails(resourceType, body)
	if transactionId == "" {
		logger.Error("Failed to determine transaction details", "resourceType", resourceType)
	}

	req, err := http.NewRequest(method, urlValue, bytes.NewReader(body))
	if err != nil {
		LogForwardError(urlValue, env, "gagal membuat request", err)
		SaveErrorTransactionBackground(repo, types.TransactionError{
			ID:           transactionId,
			Type:         "forward",
			ResourceType: resourceType,
			Env:          env,
			Url:          urlValue,
			PatientId:    patientId,
			Payload:      body,
			ErrorMessage: err.Error(),
			Status:       "PENDING",
			RetryCount:   0,
		})
		return
	}

	req.Header.Add("Authorization", auth)
	req.Header.Add("content-type", "application/json")
	req.Header.Add(cfg.FhirSource.Header, "backup-satusehat")

	if cfg.Ildki.ZeroTrust {
		if errSign := AddSignatureToRequest(cfg.Ildki.Faskes, req, body); errSign != nil {
			logger.Error("Gagal menambahkan signature pada request", "Err", errSign.Error())
			SaveErrorTransactionBackground(repo, types.TransactionError{
				ID:           transactionId,
				Type:         "forward",
				ResourceType: resourceType,
				Env:          env,
				Url:          urlValue,
				PatientId:    patientId,
				Payload:      body,
				ErrorMessage: errSign.Error(),
				Status:       "PENDING",
				RetryCount:   0,
			})
			return
		}
	}

	ExecuteBackgroundRequest(client, req, func(response *http.Response) {
		if response.StatusCode != fiber.StatusOK && response.StatusCode != fiber.StatusCreated {
			_, msg := HttpError("ildki", req, response, nil)
			SaveErrorTransactionBackground(repo, types.TransactionError{
				ID:           transactionId,
				Type:         "forward",
				ResourceType: resourceType,
				Env:          env,
				Url:          urlValue,
				PatientId:    patientId,
				Payload:      body,
				ErrorMessage: msg,
				Status:       "PENDING",
				RetryCount:   0,
			})
			logger.Error("Forward Request Error", "msg", msg)
			return
		}

		bodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			LogForwardError(urlValue, env, "gagal membaca response body", err)
			return
		}

		base := map[string]any{}
		if err := json.Unmarshal(bodyBytes, &base); err != nil {
			logger.Error("Error Marshal Response", "err", err)
			return
		}

		logger.InfoJson(fmt.Sprintf("Response Raw Forward Request (%s) ==> : ", urlValue), base)
	}, func(err error) {
		errcode, errstr := HttpError("ildki", req, nil, err)
		SaveErrorTransactionBackground(repo, types.TransactionError{
			ID:           transactionId,
			Type:         "forward",
			ResourceType: resourceType,
			Env:          env,
			Url:          urlValue,
			PatientId:    patientId,
			Payload:      body,
			ErrorMessage: errstr,
			Status:       "PENDING",
			RetryCount:   0,
		})
		logger.Error("Forward Request Connection Error", "To", urlValue, "env", env, "Err", errstr, "code", errcode)
	})
}

func LogForwardError(url, env, msg string, err error) {
	logger.Error("Forward Request", "To", url, "env", env, "Msg", msg, "Err", err.Error(), "code", ERR_HAPI_FORMAT)

}

func SaveErrorTransactionBackground(repo *repository.ProxyRepository, transaction types.TransactionError) {
	if err := repo.SaveTransactionError(transaction.ToEntity()); err != nil {
		logger.ErrorJson("Gagal Simpan Transaction Error", err)
	}
}
