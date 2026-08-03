package utils

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

const BackgroundHTTPTimeout = 15 * time.Second

func ExecuteBackgroundRequest(client *http.Client, req *http.Request, onSuccess func(*http.Response), onError func(error)) {
	executeBackgroundRequestWithLimiter(client, req, onSuccess, onError, ildkiBackupRateLimiter)
}

func ExecuteCredentialBackgroundRequest(client *http.Client, req *http.Request, onSuccess func(*http.Response), onError func(error)) {
	executeBackgroundRequestWithLimiter(client, req, onSuccess, onError, ildkiCredentialRateLimiter)
}

func executeBackgroundRequestWithLimiter(client *http.Client, req *http.Request, onSuccess func(*http.Response), onError func(error), limiter *tokenBucket) {
	limitCtx, limitCancel := context.WithTimeout(context.Background(), ildkiRateLimitWaitTimeout)
	defer limitCancel()
	if err := limiter.wait(limitCtx); err != nil {
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
