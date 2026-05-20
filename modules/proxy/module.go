package proxy

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	cron "github.com/nersus15/lib-go-cron"
	"github.com/nersus15/mini-proxy/mod-proxy/config"
	"github.com/nersus15/mini-proxy/mod-proxy/handler"
	"github.com/nersus15/mini-proxy/mod-proxy/repository"
	service2 "github.com/nersus15/mini-proxy/mod-proxy/service"
	repository2 "github.com/semanggilab/lib-go-fhir/repository"
	"github.com/semanggilab/lib-go-fhir/service"
	kafka "github.com/webcore-go/lib-kafka"
	"github.com/webcore-go/webcore/app/core"
	appConfig "github.com/webcore-go/webcore/infra/config"
	"github.com/webcore-go/webcore/infra/logger"
	"github.com/webcore-go/webcore/port"
)

const (
	ModuleName    = "proxy"
	ModuleVersion = "1.0.0"
)

// Module implements the module.Module interface
type Module struct {
	config      *config.ModuleConfig
	service     *service2.ProxyService
	authService *service2.AuthDbService
	fhirService *service.FhirTransactionService
	repository  *repository.ProxyRepository
	handler     *handler.HttpHandler
	routes      []*core.ModuleRoute
	memory      port.ICacheMemory
	kafka       *kafka.KafkaProducer
	cron        *cron.CronLibrary
	taskService *service2.TaskService
}

// NewModule creates a new Module instance
func NewModule() *Module {
	return &Module{}
}

// Name returns the unique name of the module
func (m *Module) Name() string {
	return ModuleName
}

// Version returns the version of the module
func (m *Module) Version() string {
	return ModuleVersion
}

// Dependencies returns the dependencies of the module to other modules
func (m *Module) Dependencies() []string {
	return []string{}
}

// Init initializes the module with the given app and dependencies
func (m *Module) Init(ctx *core.AppContext) error {
	// Load configuration into ModuleConfig (bind to key)
	m.config = &config.ModuleConfig{}
	if err := core.LoadDefaultConfigModule("fhir", m.config); err != nil {
		return err
	}

	// Register services and repositories
	libMem, ok := core.Instance().Context.GetDefaultSingletonInstance("cache:memory")
	if !ok {
		return fmt.Errorf("Gagal memuat instance database")
	}

	m.memory = libMem.(port.ICacheMemory)

	lib, ok := core.Instance().Context.GetDefaultSingletonInstance("database")

	if !ok {
		return fmt.Errorf("Gagal memuat instance database")
	}

	db := lib.(port.IDatabase)
	fhirConfig := m.config.ToFhirConfig()
	fhirRepo := repository2.NewFhirRepository(ctx, fhirConfig, db)

	m.repository = repository.NewProxyRepository(ctx, m.config, db)

	if m.config.Kafka.Enabled {
		loader, err := core.Instance().Context.GetDefaultLibraryLoader("kafka:producer")
		if err != nil {
			return err
		}
		libKafka, err := core.Instance().Context.LoadSingletonInstance(loader, ctx.Config.Kafka)
		if err != nil {
			return err
		}

		m.kafka = libKafka.(*kafka.KafkaProducer)
	}

	m.service = service2.NewProxyService(ctx, m.config, m.repository, m.kafka)
	m.fhirService = service.NewFhirTransactionService(ctx, fhirConfig, fhirRepo, m.memory)
	m.authService = service2.NewAuthDbService(ctx, m.config, m.fhirService)
	m.handler = handler.NewHandler(ctx, m.config, m.service)
	m.taskService = service2.NewTaskService(m.repository, m.service)

	// Register routes
	m.registerStandardRoutes(ctx.Root)
	m.registerFhirRoutes(ctx.Web)

	if m.config.Cron.Enabled {
		// Inisiai CronJob
		cronLoader, err := core.Instance().Context.GetDefaultLibraryLoader("cron")

		if err != nil {
			return fmt.Errorf("Gagal memuat instance cronLoader")
		}
		cronLib, err := core.Instance().Context.LoadSingletonInstance(cronLoader)

		if err != nil {
			return fmt.Errorf("Gagal Load CronLib")
		}
		m.cron = cronLib.(*cron.CronLibrary)

		if err := m.cron.Connect(); err != nil {
			logger.ErrorJson("Gagal Menjalankan Cron", err)
		}

		if _, err := m.cron.AddFunc(m.config.Cron.Schedule, func() {
			m.taskService.ProcessRetryTasks()
		}); err != nil {
			logger.ErrorJson("Error Menambahkan Cronjob", err)
		}
	}

	// These can be accessed through the central registry
	logger.Info("Module Stream initialized successfully")

	return nil
}

func (m *Module) Destroy() error {
	return nil
}

func (m *Module) Config() appConfig.Configurable {
	return m.config
}

func (m *Module) Routes() []*core.ModuleRoute {
	return m.routes
}

// Services returns the services provided by this module
func (m *Module) Services() map[string]any {
	// Return services that can be used by other modules
	return map[string]any{
		"proxy": m.service,
	}
}

// Repositories returns the repositories provided by this module
func (m *Module) Repositories() map[string]any {
	// Return repositories that can be used by other modules
	return map[string]any{
		"proxy": m.repository,
	}
}

// registerRoutes registers the module's routes
func (m *Module) registerStandardRoutes(root fiber.Router) {
	// Module routes
	moduleRoot := root.Group("/" + m.Name())

	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method:  "POST",
		Path:    "/oauth/:env/accesstoken",
		Handler: m.handler.GenerateToken,
		Root:    moduleRoot,
	})

	// Module-specific routes
	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method:  "GET",
		Path:    "/health",
		Handler: m.Health,
		Root:    moduleRoot,
	})

	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method:  "GET",
		Path:    "/info",
		Handler: m.Info,
		Root:    moduleRoot,
	})
}

func (m *Module) registerFhirRoutes(web *fiber.App) {
	fhirRoot := web.Group("/api", m.AuthWithSatuSehatToken)

	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method:  "GET",
		Path:    "/fhir/:env/:resourceType/:resourceId?",
		Handler: m.handler.GetResource,
		Root:    fhirRoot,
	})
	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method:  "POST",
		Path:    "/fhir/:env/:resourceType?",
		Handler: m.handler.PostResource,
		Root:    fhirRoot,
	})

	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method:  "PUT",
		Path:    "/fhir/:env/:resourceType/:resourceId",
		Handler: m.handler.PutResource,
		Root:    fhirRoot,
	})

	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method:  "PATCH",
		Path:    "/fhir/:env/:resourceType/:resourceId",
		Handler: m.handler.PatchResource,
		Root:    fhirRoot,
	})
}

func (m *Module) AuthWithSatuSehatToken(c *fiber.Ctx) error {
	var tokenString string

	// Coba dapatkan dari Authorization
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return fmt.Errorf("Authorization header required")
	}

	// konten dimulai dengan prefiks "Bearer "
	if after, ok := strings.CutPrefix(authHeader, "Bearer "); ok {
		tokenString = after

		if tokenString == "" {
			return fmt.Errorf("Token in Authorization header is missing")
		}
	} else {
		return fmt.Errorf("Required prefix in Authorization header is missing")
	}

	// Extract env from path manually since middleware runs before route matching
	// Path pattern: /local/fhir/:env/:resourceType/:resourceId?
	pathParts := strings.Split(c.Path(), "/")
	var env string
	if len(pathParts) >= 4 && pathParts[1] == "api" && pathParts[2] == "fhir" {
		env = pathParts[3]
	}
	if env != "prod" && env != "dev" {
		return fmt.Errorf("Invalid environment. Must be 'prod' or 'dev'")
	}

	// err := m.authService.ValidateToken(c, tokenString, env)
	// if err != nil {
	// 	return fmt.Errorf("Error validating token: %v", err)
	// }

	return c.Next()
}

// ModuleHealth returns the health status of the module
func (m *Module) Health(c *fiber.Ctx) error {
	health := map[string]any{
		"status":    "healthy",
		"module":    ModuleName,
		"version":   ModuleVersion,
		"timestamp": time.Now().Format(time.RFC3339),
	}
	return c.JSON(health)
}

// ModuleInfo returns information about the module
func (m *Module) Info(c *fiber.Ctx) error {
	endpoints := []string{}
	for _, endpoint := range m.routes {
		endpoint := endpoint.Method + " " + endpoint.Path
		endpoints = append(endpoints, endpoint)
	}

	path := "/" + ModuleName

	info := map[string]any{
		"name":        ModuleName,
		"version":     ModuleVersion,
		"description": "FHIR proxy",
		"path":        path,
		"endpoints":   endpoints,
		"config":      m.config,
	}
	return c.JSON(info)
}
