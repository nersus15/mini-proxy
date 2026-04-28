package service

import (
	"github.com/nersus15/mini-proxy/proxy/config"
	"github.com/nersus15/mini-proxy/proxy/repository"
	"github.com/webcore-go/webcore/app/core"
)

type ProxyService struct {
	Context    *core.AppContext
	Config     *config.ModuleConfig
	Repository *repository.ProxyRepository
	Token      *string
}

func NewProxyService(wctx *core.AppContext, cfg *config.ModuleConfig, repository *repository.ProxyRepository) *ProxyService {
	return &ProxyService{
		Context:    wctx,
		Config:     cfg,
		Repository: repository,
	}
}
