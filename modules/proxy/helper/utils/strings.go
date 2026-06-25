package utils

import (
    "net/http"
    "net/url"
    "strings"

    "github.com/nersus15/mini-proxy/mod-proxy/helper/types"
)

func InvalidURL(value string) bool {
    return value == "" || value == "/"
}

func QueryParams(params map[string]string) string {
    v := url.Values{}
    for key, value := range params {
        v.Add(key, value)
    }
    return v.Encode()
}

func ForwardPayload(headers map[string][]string, httpResponse *http.Response, url string, requestBody interface{}, raw any) types.BackupPayload {
    statusText := http.StatusText(httpResponse.StatusCode)
    if statusText == "" {
        statusText = "OK"
    }

    expressRequestHeaders := make(map[string]string)
    for key, values := range headers {
        if len(values) > 0 {
            expressRequestHeaders[strings.ToLower(key)] = strings.Join(values, ", ")
        }
    }

    expressResponseHeaders := make(map[string]string)
    for key, values := range httpResponse.Header {
        if len(values) > 0 {
            expressResponseHeaders[strings.ToLower(key)] = strings.Join(values, ", ")
        }
    }

    cleanRequestPayload := types.AxiosRequestConfig{
        URL:     url,
        Method:  "POST",
        Headers: expressRequestHeaders,
        Data:    requestBody,
    }

    payload := types.BackupPayload{
        Request: cleanRequestPayload,
        Response: types.AxiosResponse{
            Data:       raw,
            Status:     httpResponse.StatusCode,
            StatusText: statusText,
            Headers:    httpResponse.Header,
            Config: types.AxiosRequestConfig{
                URL:     url,
                Method:  "POST",
                Headers: expressRequestHeaders,
                Data:    requestBody,
            },
        },
    }

    if len(expressResponseHeaders) > 0 {
        payload.Response.Headers = httpResponse.Header
    }

    return payload
}
