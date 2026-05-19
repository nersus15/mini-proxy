package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nersus15/mini-proxy/mod-proxy/config"
	"github.com/nersus15/mini-proxy/mod-proxy/service"
	"github.com/semanggilab/lib-go-fhir/helper/types"
	"github.com/webcore-go/webcore/app/core"
	"github.com/webcore-go/webcore/app/out"
)

type HttpHandler struct {
	proxyService *service.ProxyService
	config       *config.ModuleConfig
}

// type Reque

// NewHandler creates a new Handler instance
func NewHandler(wctx *core.AppContext, cfg *config.ModuleConfig, service *service.ProxyService) *HttpHandler {
	return &HttpHandler{
		proxyService: service,
		config:       cfg,
	}
}

func (h *HttpHandler) PostResource(c *fiber.Ctx) error {
	env := c.Params("env")
	resourceType := c.Params("resourceType", "")

	var res *types.BaseResource
	var errcode int
	var errstr string

	if resourceType != "" {
		res, _, errcode, errstr = h.proxyService.PostResource(env, resourceType, c)
	} else {
		res, _, errcode, errstr = h.proxyService.PostBundle(env, resourceType, c)
	}

	if errstr != "" {
		return c.Status(fiber.StatusBadRequest).JSON(out.Error(
			fiber.StatusBadRequest,
			errcode,
			"GENERATED",
			errstr,
		))
	}

	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *HttpHandler) GetResource(c *fiber.Ctx) error {
	return nil
}

func (h *HttpHandler) PutResource(c *fiber.Ctx) error {
	return nil
}

func (h *HttpHandler) PatchResource(c *fiber.Ctx) error {
	return nil
}

func (h *HttpHandler) GetAccessToken(c *fiber.Ctx) error {
	return nil
}
