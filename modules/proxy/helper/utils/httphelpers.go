package utils

import (
	"context"
	"io"
	"net/http"
	"time"
)

const BackgroundHTTPTimeout = 15 * time.Second

func ExecuteBackgroundRequest(client *http.Client, req *http.Request, onSuccess func(*http.Response), onError func(error)) {
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
