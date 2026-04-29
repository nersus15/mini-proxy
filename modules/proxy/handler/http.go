package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nersus15/mini-proxy/mod-proxy/config"
	"github.com/nersus15/mini-proxy/mod-proxy/service"
	"github.com/webcore-go/webcore/app/core"
)

type HttpHandler struct {
	proxyService *service.ProxyService
	config       *config.ModuleConfig
}

// NewHandler creates a new Handler instance
func NewHandler(wctx *core.AppContext, cfg *config.ModuleConfig, service *service.ProxyService) *HttpHandler {
	return &HttpHandler{
		proxyService: service,
		config:       cfg,
	}
}

func (h *HttpHandler) PostResource(c *fiber.Ctx) error {
	message := map[string]any{
		"message": "Test",
	}

	return c.Status(fiber.StatusOK).JSON(message)
}
