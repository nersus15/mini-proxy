package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nersus15/mini-proxy/mod-proxy/config"
	"github.com/nersus15/mini-proxy/mod-proxy/helper/types"
	"github.com/nersus15/mini-proxy/mod-proxy/service"
	"github.com/webcore-go/webcore/app/core"
	"github.com/webcore-go/webcore/app/out"
	"github.com/webcore-go/webcore/infra/logger"
)

type HttpHandler struct {
	proxyService *service.ProxyService
	config       *config.ModuleConfig
}

type TokenPayload struct {
	ClientId     string `form:"client_id"`
	ClientSecret string `form:"client_secret"`
}

// type Reque

// NewHandler creates a new Handler instance
func NewHandler(wctx *core.AppContext, cfg *config.ModuleConfig, service *service.ProxyService) *HttpHandler {
	return &HttpHandler{
		proxyService: service,
		config:       cfg,
	}
}

func (h *HttpHandler) GenerateToken(c *fiber.Ctx) error {
	env := c.Params("env")
	var body TokenPayload
	err := c.BodyParser(&body)

	if err != nil || body.ClientId == "" || body.ClientSecret == "" {
		return c.Status(fiber.StatusBadRequest).JSON(out.Error(
			fiber.StatusBadRequest,
			fiber.StatusBadRequest,
			"INVALID PAYLOAD",
			"client_id and client_secret cannot be empty",
		))
	}
	logger.InfoJson("Body", body)

	// Panggil service (otomatis fallback hapi -> satusehat di dalam service)
	clientCredential, errcode, errstr := h.proxyService.GenerateToken(env, "satusehat", body.ClientId, body.ClientSecret)

	if errstr != "" {

		return c.Status(fiber.StatusBadRequest).JSON(out.Error(
			fiber.StatusBadRequest,
			errcode,
			"GENERATION_FAILED",
			errstr,
		))
	}

	// Pastikan access_token benar-benar ada sebelum memicu background worker
	if clientCredential.AccessToken != "" {
		h.proxyService.SaveCredential(env, &clientCredential)
	} else {
		logger.Warn("Access token is empty, skipping background save")
	}

	return c.Status(fiber.StatusOK).JSON(clientCredential)
}

func (h *HttpHandler) GetResource(c *fiber.Ctx) error {
	env := c.Params("env")
	resourceType := c.Params("resourceType", "")
	resourceId := c.Params("resourceId", "")
	queryParams := c.Queries()

	_, raw, errcode, errstr := h.proxyService.GetResource(env, resourceType, resourceId, queryParams, c)
	if errstr != "" {
		return c.Status(fiber.StatusBadRequest).JSON(out.Error(
			fiber.StatusBadRequest,
			errcode,
			"GENERATED",
			errstr,
		))
	}
	return c.Status(fiber.StatusOK).JSON(raw)
}

func (h *HttpHandler) PostResource(c *fiber.Ctx) error {
	env := c.Params("env")
	resourceType := c.Params("resourceType", "")

	var res *types.BaseResource
	var raw any
	var errcode int
	var errstr string

	res, raw, errcode, errstr = h.proxyService.PostResource(env, resourceType, c)

	// if resourceType != "" {
	// 	res, raw, errcode, errstr = h.proxyService.PostResource(env, resourceType, c)
	// } else {
	// 	res, raw, errcode, errstr = h.proxyService.PostBundle(env, resourceType, c)
	// }
	if errstr != "" {
		if res != nil && res.ResourceType != nil && *res.ResourceType == "OperationOutcome" {
			return c.Status(fiber.StatusBadRequest).JSON(raw)
		}

		return c.Status(fiber.StatusBadRequest).JSON(out.Error(
			fiber.StatusBadRequest,
			errcode,
			"GENERATED",
			errstr,
		))
	}
	return c.Status(fiber.StatusOK).JSON(raw)
}

func (h *HttpHandler) PutResource(c *fiber.Ctx) error {
	env := c.Params("env")
	resourceType := c.Params("resourceType", "")
	resId := c.Params("resourceId", "")

	_, raw, errcode, errstr := h.proxyService.PutResource(env, resourceType, resId, c)

	if errstr != "" {
		return c.Status(fiber.StatusBadRequest).JSON(out.Error(
			fiber.StatusBadRequest,
			errcode,
			"GENERATED",
			errstr,
		))
	}

	return c.Status(fiber.StatusOK).JSON(raw)
}

func (h *HttpHandler) PatchResource(c *fiber.Ctx) error {
	env := c.Params("env")
	resourceType := c.Params("resourceType", "")
	resId := c.Params("resourceId", "")

	_, raw, errcode, errstr := h.proxyService.PatchResource(env, resourceType, resId, c)

	if errstr != "" {
		return c.Status(fiber.StatusBadRequest).JSON(out.Error(
			fiber.StatusBadRequest,
			errcode,
			"GENERATED",
			errstr,
		))
	}

	return c.Status(fiber.StatusOK).JSON(raw)
}
