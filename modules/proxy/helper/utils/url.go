package utils

import (
	"net/url"

	"github.com/gofiber/fiber/v2"
	"github.com/nersus15/mini-proxy/mod-proxy/config"
	"github.com/webcore-go/webcore/infra/logger"
)

func GetHostName(u string) string {
	ur, err := url.Parse(u)
	if err != nil {
		logger.ErrorJson("Gagal Parse Url", err)
		return ""
	}
	return ur.Hostname()
}

// GetUrl returns (mainUrl, forwardUrl, priority)
func GetUrl(cfg *config.ModuleConfig, env string, resourceType string, ctx *fiber.Ctx) (string, string, string) {
	priority := cfg.FhirSource.Priority
	header := ctx.Get(cfg.FhirSource.Header, "")

	if header != "" {
		priority = header
	}

	satusehatUrl := ""
	localUrl := ""
	if env == "dev" {
		satusehatUrl = cfg.SatSetDev.BaseURL + "/" + resourceType
		localUrl = cfg.Ildki.DevelopmentURL + "/" + resourceType
	} else {
		satusehatUrl = cfg.SatSetProd.BaseURL + "/" + resourceType
		localUrl = cfg.Ildki.ProductionURL + "/" + resourceType
	}

	switch priority {
	case "local-first":
		return localUrl, satusehatUrl, priority
	case "satusehat-first":
		return satusehatUrl, localUrl, priority
	default:
		return "", "", priority
	}
}

func GetTarget(priority string) (string, string) {
	target := "satusehat"
	forward := "ildki"
	if priority == "local-first" {
		target = "ildki"
		forward = "satusehat"
	}

	return target, forward
}
