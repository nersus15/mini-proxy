package handler

import (
	"github.com/nersus15/mini-proxy/config"
	"github.com/nersus15/mini-proxy/service"
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
