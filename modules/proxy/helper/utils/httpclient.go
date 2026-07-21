package utils

import (
	"crypto/tls"
	"net"
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

		// Connection pooling — default Go MaxIdleConnsPerHost cuma 2, terlalu kecil
		// untuk trafik tinggi ke host yang sama (ILDKI/SatuSehat). Tanpa ini, tiap
		// request concurrent di atas 2 akan buka koneksi TCP+TLS baru (mahal, dan
		// mempercepat kena rate limit di sisi ILDKI karena tiap handshake dihitung).
		MaxIdleConns:        200,
		MaxIdleConnsPerHost:  50,
		MaxConnsPerHost:      50, // sekaligus jadi self-throttle agar mini-proxy tidak membanjiri rate limit ILDKI saat traffic spike
		IdleConnTimeout:      90 * time.Second,

		// Pisahkan timeout per-fase koneksi, bukan satu Timeout total 60s untuk semuanya.
		// Supaya fase yang macet (connect/TLS/response header) gagal cepat dan tidak
		// menahan slot goroutine/worker terlalu lama.
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
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
