package utils

import (
	"fmt"
	"io"
	"net/http"
)

const (
	ERR_SATUSEHAT_TOKEN           = iota
	ERR_SATUSEHAT_VALIDATION      // validasi payload gagal
	ERR_SATUSEHAT_FORMAT          // validasi format response gagal
	ERR_SATUSEHAT_REFERENCE       // gagal simpan karena reference tidak tersedia
	ERR_SATUSEHAT_ACCESS_RESOURCE // akses ke resource ditolak
	ERR_SATUSEHAT_SERVER          // server error 50x
	ERR_SATUSEHAT_NETWORK         // timeout, unreachable
	ERR_SATUSEHAT_UNDEFINED       // keranjang sampah
	ERR_HAPI_VALIDATION           // validasi payload gagal
	ERR_HAPI_FORMAT               // validasi format response gagal
	ERR_HAPI_REFERENCE            // gagal simpan karena reference tidak tersedia
	ERR_HAPI_ACCESS_RESOURCE      // akses ke resource ditolak
	ERR_HAPI_SERVER               // server error 50x
	ERR_HAPI_NETWORK              // timeout, unreachable
	ERR_HAPI_UNDEFINED            // keranjang sampah
)

func HttpError(target string, req *http.Request, resp *http.Response, err error) (int, string) {
	var errstr string
	if err != nil {
		if resp == nil {
			errstr = fmt.Sprintf("gagal mengirim request generate: %s", err)
			if target == "ildki" {
				return ERR_HAPI_NETWORK, errstr
			}
			return ERR_SATUSEHAT_NETWORK, errstr
		} else {
			if target == "ildki" {
				return ERR_HAPI_UNDEFINED, err.Error()
			}
			return ERR_SATUSEHAT_UNDEFINED, err.Error()
		}
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 500 {
		errstr = fmt.Sprintf("Error Server Status: %s, Method: %s, URL: %s, Body: %s", resp.Status, req.Method, req.URL, string(bodyBytes))
		if target == "ildki" {
			return ERR_HAPI_SERVER, errstr
		}
		return ERR_SATUSEHAT_SERVER, errstr
	} else if resp.StatusCode == 401 {
		errstr = fmt.Sprintf("Error Token: %s, Method: %s, URL: %s, Body: %s", resp.Status, req.Method, req.URL, string(bodyBytes))
		if target == "ildki" {
			return ERR_HAPI_ACCESS_RESOURCE, errstr
		}
		return ERR_SATUSEHAT_TOKEN, errstr
	} else if resp.StatusCode >= 403 && resp.StatusCode <= 406 {
		errstr = fmt.Sprintf("Error Akses: %s, Method: %s, URL: %s, Body: %s", resp.Status, req.Method, req.URL, string(bodyBytes))
		if target == "ildki" {
			return ERR_HAPI_ACCESS_RESOURCE, errstr
		}
		return ERR_SATUSEHAT_ACCESS_RESOURCE, errstr
	} else if resp.StatusCode == 408 || resp.StatusCode == 410 || resp.StatusCode == 425 || resp.StatusCode == 429 {
		errstr = fmt.Sprintf("Timeout: %s, Method: %s, URL: %s, Body: %s", resp.Status, req.Method, req.URL, string(bodyBytes))
		if target == "ildki" {
			return ERR_HAPI_NETWORK, errstr
		}
		return ERR_SATUSEHAT_NETWORK, errstr
	} else if resp.StatusCode >= 400 {
		errstr = fmt.Sprintf("Error Akses: %s, Method: %s, URL: %s, Body: %s", resp.Status, req.Method, req.URL, string(bodyBytes))
		if target == "ildki" {
			return ERR_HAPI_ACCESS_RESOURCE, errstr
		}
		return ERR_SATUSEHAT_ACCESS_RESOURCE, errstr
	}

	return 0, ""
}

func OperationOutcomeError(target string, diagnostic string) (int, string) {
	errstr := fmt.Sprintf("Response OperationOutcome dari %s: %s", target, diagnostic)
	// TODO: disini untuk errof VALIDATION, FORMAT, TOKEN, ACCESS_RESOURCE dan REFERENCE
	if target == "ildki" {
		return ERR_HAPI_VALIDATION, errstr
	}
	return ERR_SATUSEHAT_VALIDATION, errstr
}
