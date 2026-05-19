package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/nersus15/mini-proxy/mod-proxy/entity"
	"github.com/nersus15/mini-proxy/mod-proxy/repository"
	"github.com/semanggilab/lib-go-fhir/helper/utils"
	"github.com/webcore-go/webcore/infra/logger"
)

type TaskService struct {
	Repository *repository.ProxyRepository
	Proxy      *ProxyService
}

func NewTaskService(r *repository.ProxyRepository, p *ProxyService) *TaskService {
	return &TaskService{Repository: r, Proxy: p}
}

// Task yang akan dijalankan oleh Cron
func (s *TaskService) ProcessRetryTasks() {
	tasks, err := s.Repository.GetPendingTransactions(5)
	if err != nil {
		fmt.Println("Error fetching tasks:", err)
		return
	}

	logger.Info("Available Task", len(tasks))
	for i := range tasks {
		fmt.Printf("Processing Task ID: %s (Retry: %d)\n", tasks[i].ID, tasks[i].RetryCount)
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

	// Sekarang tasks berisi data yang sudah terupdate
	logger.Info("Loop finished, starting bulk update...")
	if len(tasks) > 0 {
		err := s.Repository.UpdateBulkTransactions(tasks)
		if err != nil {
			logger.DebugJson("Bulk Update Error:", err)
		}
	}
}

func (s *TaskService) executeAction(task entity.Transactions) string {
	// Logika pengiriman data sebenarnya di sini
	logger.InfoJson("MenjalankanTask", map[string]any{
		"id":          task.ID,
		"resouceType": task.ResourceType,
		"retryCount":  task.RetryCount,
		"patientId":   task.PatientId,
		"status":      task.Status,
		"createdAt":   task.CreatedAt,
		"updatedAt":   task.UpdatedAt,
		"env":         task.Env,
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
		req.Header.Add(s.Proxy.Config.FhirSource.Header, "local-first")
	}

	response, err := s.Proxy.httpClient(&env).Do(req)
	if err != nil {
		_, errstr = utils.HttpError("kafka", req, nil, err)
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
		_, msg := utils.HttpError("kafka", req, response, nil)

		return msg
	}

	return ""
}
