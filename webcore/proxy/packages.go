package proxy

import (
	"github.com/nersus15/mini-proxy/proxy"
	"github.com/webcore-go/webcore/app/core"
)

var APP_PACKAGES = []core.Module{
	proxy.NewModule(),
	// Add your packages here
}
