package tests

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nersus15/mini-proxy/mod-proxy/helper/utils"
)

func newDummyRequest(t *testing.T) *http.Request {
	t.Helper()
	req, err := http.NewRequest("GET", "https://example.com/fhir/Patient/123", nil)
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}
	return req
}

func newDummyResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Status:     http.StatusText(statusCode),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestHttpError_ConnectionError(t *testing.T) {
	req := newDummyRequest(t)
	connErr := errors.New("connection refused")

	// target = satusehat, resp = nil (koneksi gagal sebelum ada response)
	code, msg := utils.HttpError("satusehat", req, nil, connErr)
	if code != utils.ERR_SATUSEHAT_NETWORK {
		t.Errorf("code = %d, want ERR_SATUSEHAT_NETWORK (%d)", code, utils.ERR_SATUSEHAT_NETWORK)
	}
	if msg == "" {
		t.Error("expected non-empty error message")
	}

	// target = ildki, resp = nil
	code, _ = utils.HttpError("ildki", req, nil, connErr)
	if code != utils.ERR_HAPI_NETWORK {
		t.Errorf("code = %d, want ERR_HAPI_NETWORK (%d)", code, utils.ERR_HAPI_NETWORK)
	}
}

func TestHttpError_ServerError(t *testing.T) {
	req := newDummyRequest(t)
	resp := newDummyResponse(http.StatusInternalServerError, `{"error":"internal"}`)

	code, msg := utils.HttpError("satusehat", req, resp, nil)
	if code != utils.ERR_SATUSEHAT_SERVER {
		t.Errorf("code = %d, want ERR_SATUSEHAT_SERVER (%d)", code, utils.ERR_SATUSEHAT_SERVER)
	}
	if msg == "" {
		t.Error("expected non-empty error message")
	}
}

func TestHttpError_Unauthorized(t *testing.T) {
	req := newDummyRequest(t)
	resp := newDummyResponse(http.StatusUnauthorized, `{"error":"unauthorized"}`)

	code, _ := utils.HttpError("satusehat", req, resp, nil)
	if code != utils.ERR_SATUSEHAT_TOKEN {
		t.Errorf("code = %d, want ERR_SATUSEHAT_TOKEN (%d)", code, utils.ERR_SATUSEHAT_TOKEN)
	}

	respIldki := newDummyResponse(http.StatusUnauthorized, `{"error":"unauthorized"}`)
	code, _ = utils.HttpError("ildki", req, respIldki, nil)
	if code != utils.ERR_HAPI_ACCESS_RESOURCE {
		t.Errorf("code = %d, want ERR_HAPI_ACCESS_RESOURCE (%d)", code, utils.ERR_HAPI_ACCESS_RESOURCE)
	}
}

func TestHttpError_Forbidden(t *testing.T) {
	req := newDummyRequest(t)
	resp := newDummyResponse(http.StatusForbidden, `{"error":"forbidden"}`)

	code, _ := utils.HttpError("satusehat", req, resp, nil)
	if code != utils.ERR_SATUSEHAT_ACCESS_RESOURCE {
		t.Errorf("code = %d, want ERR_SATUSEHAT_ACCESS_RESOURCE (%d)", code, utils.ERR_SATUSEHAT_ACCESS_RESOURCE)
	}
}

func TestHttpError_Timeout(t *testing.T) {
	cases := []int{http.StatusRequestTimeout, http.StatusGone, http.StatusTooManyRequests}

	for _, statusCode := range cases {
		req := newDummyRequest(t)
		resp := newDummyResponse(statusCode, `{"error":"timeout"}`)

		code, _ := utils.HttpError("satusehat", req, resp, nil)
		if code != utils.ERR_SATUSEHAT_NETWORK {
			t.Errorf("status %d: code = %d, want ERR_SATUSEHAT_NETWORK (%d)", statusCode, code, utils.ERR_SATUSEHAT_NETWORK)
		}
	}
}

func TestHttpError_Success(t *testing.T) {
	req := newDummyRequest(t)
	resp := newDummyResponse(http.StatusOK, `{"resourceType":"Patient"}`)

	code, msg := utils.HttpError("satusehat", req, resp, nil)
	if code != 0 || msg != "" {
		t.Errorf("expected no error for 200 OK, got code=%d msg=%q", code, msg)
	}
}

func TestOperationOutcomeError(t *testing.T) {
	code, msg := utils.OperationOutcomeError("satusehat", "duplicate resource")
	if code != utils.ERR_SATUSEHAT_VALIDATION {
		t.Errorf("code = %d, want ERR_SATUSEHAT_VALIDATION (%d)", code, utils.ERR_SATUSEHAT_VALIDATION)
	}
	if !strings.Contains(msg, "duplicate resource") {
		t.Errorf("msg = %q, want it to contain diagnostic text", msg)
	}

	code, _ = utils.OperationOutcomeError("ildki", "invalid reference")
	if code != utils.ERR_HAPI_VALIDATION {
		t.Errorf("code = %d, want ERR_HAPI_VALIDATION (%d)", code, utils.ERR_HAPI_VALIDATION)
	}
}

// Pastikan helper request test benar-benar valid untuk dipakai ulang test lain.
func TestNewDummyRequest_Sanity(t *testing.T) {
	req := newDummyRequest(t)
	rr := httptest.NewRecorder()
	if rr == nil {
		t.Fatal("unexpected nil recorder")
	}
	if req.Method != "GET" {
		t.Errorf("method = %q, want GET", req.Method)
	}
}
