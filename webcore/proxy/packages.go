package proxy

import (
	proxy "github.com/nersus15/mini-proxy/mod-proxy"
	"github.com/webcore-go/webcore/app/core"
)

var APP_PACKAGES = []core.Module{
	proxy.NewModule(),
	// Add your packages here
}
