package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/nersus15/mini-proxy/mod-proxy/config"
	"github.com/nersus15/mini-proxy/mod-proxy/entity"
	"github.com/nersus15/mini-proxy/mod-proxy/helper/utils"
	"github.com/nersus15/mini-proxy/mod-proxy/repository"
	"github.com/webcore-go/webcore/infra/logger"
)

type TaskService struct {
	Repository *repository.ProxyRepository
	Proxy      *ProxyService
	Config     *config.ModuleConfig
}

func NewTaskService(r *repository.ProxyRepository, p *ProxyService, config *config.ModuleConfig) *TaskService {
	return &TaskService{Repository: r, Proxy: p, Config: config}
}

// Task yang akan dijalankan oleh Cron
func (s *TaskService) ProcessRetryTasks(token *string) {
	// limit = 0 -> unlimited
	tasks, err := s.Repository.GetPendingTransactions(token, s.Config.Cron.ChunkSize)
	if err != nil {
		fmt.Println("Error fetching tasks:", err)
		return
	}

	logger.Info("Available Task", len(tasks))
	// Get Token Yang Terbaru
	newTokens := s.Repository.GetToken()
	logger.InfoJson("Token Yang Akan Digunakan", newTokens)

	for i := range tasks {
		fmt.Printf("Processing Task ID: %s (Retry: %d, ErrLimit: %d)\n", tasks[i].ID, tasks[i].RetryCount, s.Config.Cron.BackupLimit)

		if token == nil {
			var payload map[string]json.RawMessage
			if err := json.Unmarshal(tasks[i].Payload, &payload); err != nil {
				panic(err)
			}

			env := tasks[i].Env
			if _, ok := newTokens[env]; ok {
				newAuthorization, _ := json.Marshal(newTokens[env])
				payload["authorization"] = newAuthorization
			}
			tasks[i].Payload, err = json.Marshal(payload)
			if err != nil {
				logger.ErrorJson("Gagal Marshall Payload", err)
			}
		}

		if tasks[i].RetryCount >= int(s.Config.Cron.BackupLimit) && s.Config.Ildki.BackupBaseUrl != "" {
			newUrl, err := ReplaceHostname(tasks[i].Url, s.Config.Ildki.BackupBaseUrl)
			if err == nil {
				newUrl = strings.ReplaceAll(newUrl, "api/", "")
				tasks[i].Url = newUrl
			}
			logger.Info(fmt.Sprintf("Hostname Error Fallback TO alternative Domain(%s) .....", tasks[i].Url))
		}

		errMsg := s.executeAction(tasks[i])
		if errMsg == "" {
			tasks[i].Status = "COMPLETED"
			tasks[i].ErrorMessage = ""
		} else {
			fmt.Printf("Task %s failed: %s. Continuing next task...\n", tasks[i].ID, errMsg)

			tasks[i].Status = "RETRY"
			tasks[i].RetryCount = tasks[i].RetryCount + 1
			tasks[i].ErrorMessage = errMsg
		}
	}

	logger.Info("Loop finished, starting bulk update...")
	if len(tasks) > 0 {
		err := s.Repository.UpdateBulkTransactions(tasks)
		if err != nil {
			logger.DebugJson("Bulk Update Error:", err)
		}

		// CleanUp
		logger.Info("Start Cleanup Completed Tasks.....")
		s.Repository.DeleteOldTransactions("COMPLETED", token != nil)
	}
}

func (s *TaskService) executeAction(task entity.Transactions) string {
	// Logika pengiriman data sebenarnya di sini
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(task.Payload, &payload); err != nil {
		panic(err)
	}
	logger.InfoJson("MenjalankanTask", map[string]any{
		"id":          task.ID,
		"URL":         task.Url,
		"resouceType": task.ResourceType,
		"retryCount":  task.RetryCount,
		"patientId":   task.PatientId,
		"status":      task.Status,
		"createdAt":   task.CreatedAt,
		"updatedAt":   task.UpdatedAt,
		"env":         task.Env,
		"Auth":        payload["authorization"],
	})

	body, _ := json.Marshal(task.Payload)
	err := s.sendRequest(task.Type, body, task.Env, task.ID, "", task.Url)

	return err
}

func (s *TaskService) sendRequest(task_type string, body []byte, env string, transactionId string, auth string, url string) string {
	var errstr string
	req, err := http.NewRequest("POST", url, strings.NewReader(string(body)))
	if err != nil {
		return err.Error()
	}

	// Set Request Headers
	req.Header.Add("Authorization", auth)
	req.Header.Add("content-type", "application/json")

	if task_type == "kafka" && transactionId != "" {
		req.Header.Add("x-request-id", transactionId)
	} else if task_type == "forward" {
		req.Header.Add(s.Proxy.Config.FhirSource.Header, "backup-satusehat")
	}

	response, err := s.Proxy.httpClient(&env).Do(req)
	if err != nil {
		_, errstr = utils.HttpError("satusehat", req, nil, err)
		return errstr
	}

	defer response.Body.Close()

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return err.Error()
	}
	base := map[string]any{}
	err1 := json.Unmarshal(bodyBytes, &base)

	if err1 != nil {
		return err1.Error()
	}

	if response.StatusCode != fiber.StatusOK && response.StatusCode != fiber.StatusCreated {
		_, msg := utils.HttpError("satusehat", req, response, nil)

		return msg
	}

	return ""
}

// Credential
func (s *TaskService) SyncCredential() {
	// Get Active Token
	logger.Info("Jalankan Job Untuk Sinkronisasi Credential")
	tokens := s.Repository.GetToken()

	logger.InfoJson("Tokens", tokens)
	if _, ok := tokens["dev"]; ok {
		token := tokens["dev"]

		s.ProcessRetryTasks(&token)
	}

	if _, ok := tokens["prod"]; ok {
		token := tokens["prod"]

		s.ProcessRetryTasks(&token)
	}

}

func ReplaceHostname(rawURL, newHost string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	if port := u.Port(); port != "" {
		u.Host = net.JoinHostPort(newHost, port)
	} else {
		u.Host = newHost
	}

	return u.String(), nil
}
