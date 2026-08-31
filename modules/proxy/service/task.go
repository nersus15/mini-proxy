package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/gofiber/fiber/v2"
	"github.com/nersus15/mini-proxy/mod-proxy/config"
	"github.com/nersus15/mini-proxy/mod-proxy/entity"
	"github.com/nersus15/mini-proxy/mod-proxy/helper/types"
	"github.com/nersus15/mini-proxy/mod-proxy/helper/utils"
	"github.com/nersus15/mini-proxy/mod-proxy/repository"
	"github.com/webcore-go/webcore/infra/logger"
)

type TaskService struct {
	Repository           *repository.ProxyRepository
	Proxy                *ProxyService
	Config               *config.ModuleConfig
	resourceProcessing   int32
	credentialProcessing int32
}

func NewTaskService(r *repository.ProxyRepository, p *ProxyService, config *config.ModuleConfig) *TaskService {
	return &TaskService{Repository: r, Proxy: p, Config: config}
}
func (s *TaskService) ProcessResourceRetryTasks() {
	if !atomic.CompareAndSwapInt32(&s.resourceProcessing, 0, 1) {
		logger.Warn("ProcessResourceRetryTasks dilewati: eksekusi sebelumnya masih berjalan, mencegah duplikasi pengiriman resource ke ILDKI")
		return
	}
	defer atomic.StoreInt32(&s.resourceProcessing, 0)

	tasks, err := s.Repository.GetPendingTransactions(nil, s.Config.Cron.ChunkSize)
	if err != nil {
		logger.Error("Gagal mengambil resource retry tasks", "err", err)
		return
	}

	logger.Info("Resource Retry Tasks Tersedia: " + strconv.Itoa(len(tasks)))
	if len(tasks) == 0 {
		return
	}

	newTokens := s.Repository.GetToken(nil)
	logger.InfoJson("Token Untuk Refresh Authorization", newTokens)

	for i := range tasks {
		if err := refreshTaskAuthorization(&tasks[i], newTokens); err != nil {
			logger.Error("Gagal refresh authorization task, dilewati", "id", tasks[i].ID, "err", err)
			continue
		}
		s.applyBackupHostnameFallback(&tasks[i])
	}
	s.runRetryBatch(tasks, false)
}
func (s *TaskService) ProcessCredentialRetryTasks() {
	if !atomic.CompareAndSwapInt32(&s.credentialProcessing, 0, 1) {
		logger.Warn("ProcessCredentialRetryTasks dilewati: eksekusi sebelumnya masih berjalan, mencegah duplikasi sync credential ke ILDKI")
		return
	}
	defer atomic.StoreInt32(&s.credentialProcessing, 0)

	tasks, err := s.Repository.GetPendingCredentialTransactions(s.Config.Cron.ChunkSize)
	if err != nil {
		logger.Error("Gagal mengambil credential retry tasks", "err", err)
		return
	}
	logger.Info("Credential Retry Tasks Tersedia: " + strconv.Itoa(len(tasks)))

	for i := range tasks {
		s.applyBackupHostnameFallback(&tasks[i])
	}
	s.runRetryBatch(tasks, true)
}

func (s *TaskService) SyncCredential() {
	logger.Info("Jalankan Job Untuk Sinkronisasi Credential")
	s.ProcessCredentialRetryTasks()
}

func (s *TaskService) runRetryBatch(tasks []entity.Transactions, credOnlyCleanup bool) {
	for i := range tasks {
		errMsg := s.executeAction(tasks[i])
		if errMsg == "" {
			tasks[i].Status = "COMPLETED"
			tasks[i].ErrorMessage = ""
		} else {
			logger.Warn("Task gagal, akan di-retry siklus berikutnya", "id", tasks[i].ID, "retryCount", tasks[i].RetryCount, "err", errMsg)
			tasks[i].Status = "RETRY"
			tasks[i].RetryCount = tasks[i].RetryCount + 1
			tasks[i].ErrorMessage = errMsg
		}
	}

	if len(tasks) == 0 {
		return
	}

	logger.Info("Selesai proses batch, mulai bulk update...")
	if err := s.Repository.UpdateBulkTransactions(tasks); err != nil {
		logger.Error("Bulk Update Error", "err", err)
	}

	logger.Info("Bersihkan transaksi COMPLETED...")
	if _, err := s.Repository.DeleteOldTransactions("COMPLETED", credOnlyCleanup); err != nil {
		logger.Error("Gagal cleanup transaksi COMPLETED", "err", err)
	}
}
func (s *TaskService) applyBackupHostnameFallback(task *entity.Transactions) {
	if task.RetryCount < int(s.Config.Cron.BackupLimit) || s.Config.Ildki.BackupBaseUrl == "" {
		return
	}

	newUrl, err := ReplaceHostname(task.Url, s.Config.Ildki.BackupBaseUrl)
	if err != nil {
		logger.Error("Gagal replace hostname untuk fallback domain", "id", task.ID, "err", err)
		return
	}

	task.Url = strings.ReplaceAll(newUrl, "api/", "")
	logger.Info(fmt.Sprintf("Hostname Error Fallback TO alternative Domain(%s) .....", task.Url))
}
func refreshTaskAuthorization(task *entity.Transactions, newTokens []types.Token) error {
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(task.Payload, &payload); err != nil {
		return fmt.Errorf("gagal unmarshal payload: %w", err)
	}

	newToken := ""

	for _, token := range newTokens {
		if task.Client == token.ClientId && task.Env == token.Env {
			newToken = token.AccessToken
			break
		}
	}

	if newToken == "" {
		return nil
	}
	newAuthorization, err := json.Marshal(newToken)
	if err != nil {
		return fmt.Errorf("gagal marshal token baru: %w", err)
	}
	payload["authorization"] = newAuthorization
	marshalled, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("gagal marshal payload: %w", err)
	}
	task.Payload = marshalled
	return nil
}
func (s *TaskService) executeAction(task entity.Transactions) string {
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(task.Payload, &payload); err != nil {
		logger.Error(err.Error())
		return err.Error()
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
	return s.sendRequest(task.Type, task.ResourceType, body, task.Env, task.ID, "", task.Url)
}

func (s *TaskService) sendRequest(task_type string, resourceType string, body []byte, env string, transactionId string, auth string, url string) string {
	var errstr string
	if waitErr := utils.WaitForIldkiSlot(resourceType); waitErr != nil {
		return fmt.Sprintf("rate limiter ke ILDKI penuh, task ditunda ke siklus retry berikutnya: %s", waitErr.Error())
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(body)))
	if err != nil {
		return err.Error()
	}

	req.Header.Add("Authorization", auth)
	req.Header.Add("content-type", "application/json")

	if task_type == "kafka" && transactionId != "" {
		req.Header.Add("x-request-id", transactionId)
	} else if task_type == "forward" {
		req.Header.Add(s.Proxy.Config.FhirSource.Header, "backup-satusehat")
	}

	response, err := s.Proxy.httpClientBackground(&env).Do(req)
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
	if err := json.Unmarshal(bodyBytes, &base); err != nil {
		return err.Error()
	}

	if response.StatusCode != fiber.StatusOK && response.StatusCode != fiber.StatusCreated {
		_, msg := utils.HttpError("satusehat", req, response, nil)
		return msg
	}

	return ""
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
