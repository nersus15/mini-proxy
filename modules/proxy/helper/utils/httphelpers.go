package utils

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

const BackgroundHTTPTimeout = 15 * time.Second

// ExecuteBackgroundRequest dipakai untuk jalur BACKUP/FORWARD resource ke
// ILDKI (lihat ForwardRequestToIldkiBackground di proxy_helpers.go), pakai
// ildkiBackupRateLimiter. Jalur real-time (GetResource/PostResource/dst)
// TIDAK lewat sini (mereka pakai SendRequest() langsung).
func ExecuteBackgroundRequest(client *http.Client, req *http.Request, onSuccess func(*http.Response), onError func(error)) {
	executeBackgroundRequestWithLimiter(client, req, onSuccess, onError, ildkiBackupRateLimiter)
}

// ExecuteCredentialBackgroundRequest dipakai KHUSUS untuk jalur sync
// credential ke ILDKI (lihat sendCredentialToProxyILBackground di proxy.go),
// pakai ildkiCredentialRateLimiter yang terpisah dari limiter backup —
// supaya volume backup yang besar tidak bisa menyandera credential yang
// kritis untuk resolve dependency resource di sisi ILDKI.
func ExecuteCredentialBackgroundRequest(client *http.Client, req *http.Request, onSuccess func(*http.Response), onError func(error)) {
	executeBackgroundRequestWithLimiter(client, req, onSuccess, onError, ildkiCredentialRateLimiter)
}

func executeBackgroundRequestWithLimiter(client *http.Client, req *http.Request, onSuccess func(*http.Response), onError func(error), limiter *tokenBucket) {
	limitCtx, limitCancel := context.WithTimeout(context.Background(), ildkiRateLimitWaitTimeout)
	defer limitCancel()
	if err := limiter.wait(limitCtx); err != nil {
		// Sudah menunggu sampai ildkiRateLimitWaitTimeout tapi jatah rate limit
		// belum juga longgar -> jangan tahan goroutine/slot worker pool lebih
		// lama lagi. Lepas ke jalur onError yang sudah ada, yang di kedua
		// pemanggil (credential & forward) akan menyimpannya sebagai transaksi
		// PENDING untuk dicoba lagi oleh cronjob retry.
		onError(fmt.Errorf("rate limiter ke ILDKI penuh setelah menunggu %s, request dialihkan ke retry: %w", ildkiRateLimitWaitTimeout, err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), BackgroundHTTPTimeout)
	defer cancel()

	req = req.WithContext(ctx)
	response, err := client.Do(req)
	if err != nil {
		onError(err)
		return
	}
	defer CloseBody(response.Body)
	onSuccess(response)
}

func CloseBody(body io.ReadCloser) {
	_ = body.Close()
}
