package repository

import (
	"github.com/nersus15/mini-proxy/proxy/config"
	"github.com/webcore-go/webcore/app/core"
	"github.com/webcore-go/webcore/port"
)

type ProxyRepository struct {
	Connnection port.IDatabase
	Context     *core.AppContext
	Config      *config.ModuleConfig
}

func NewProxyRepository(ctx *core.AppContext, cfg *config.ModuleConfig, conn port.IDatabase) *ProxyRepository {
	return &ProxyRepository{
		Connnection: conn,
		Context:     ctx,
		Config:      cfg,
	}
}
