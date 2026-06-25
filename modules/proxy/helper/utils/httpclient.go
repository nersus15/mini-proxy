package utils

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/webcore-go/webcore/infra/logger"
)

func CreateHttpClient(proxyUrl string) *http.Client {
	if proxyUrl == "" {
		if os.Getenv("HTTP_PROXY") != "" {
			proxyUrl = os.Getenv("HTTP_PROXY")
		} else if os.Getenv("http_proxy") != "" {
			proxyUrl = os.Getenv("http_proxy")
		}
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	if proxyUrl != "" {
		proxyURL, err := url.Parse(proxyUrl)
		if err != nil {
			logger.Fatal("Proxy URL Development", "error", err)
		}
		tr.Proxy = http.ProxyURL(proxyURL)
		logger.Debug("HTTP Client", "Proxy", proxyURL)
	}

	return &http.Client{
		Transport: tr,
		Timeout:   60 * time.Second,
	}
}
