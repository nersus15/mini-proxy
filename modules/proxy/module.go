package proxy

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nersus15/mini-proxy/proxy/config"
	"github.com/nersus15/mini-proxy/proxy/handler"
	"github.com/nersus15/mini-proxy/proxy/repository"
	"github.com/nersus15/mini-proxy/proxy/service"
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
	config     *config.ModuleConfig
	service    *service.ProxyService
	repository *repository.ProxyRepository
	handler    *handler.HttpHandler
	routes     []*core.ModuleRoute
	memory     port.ICacheMemory
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

	m.repository = repository.NewProxyRepository(ctx, m.config, db)

	m.service = service.NewProxyService(ctx, m.config, m.repository)
	m.handler = handler.NewHandler(ctx, m.config, m.service)

	// Register routes
	m.registerStandardRoutes(ctx.Root)

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
		"stream": m.service,
	}
}

// Repositories returns the repositories provided by this module
func (m *Module) Repositories() map[string]any {
	// Return repositories that can be used by other modules
	return map[string]any{
		"stream": m.repository,
	}
}

// registerRoutes registers the module's routes
func (m *Module) registerStandardRoutes(root fiber.Router) {
	// Module routes
	moduleRoot := root.Group("/" + m.Name())

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
		"description": "FHIR Stream",
		"path":        path,
		"endpoints":   endpoints,
		"config":      m.config,
	}
	return c.JSON(info)
}
