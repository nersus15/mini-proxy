package tests

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/nersus15/mini-proxy/mod-proxy/config"
	"github.com/nersus15/mini-proxy/mod-proxy/helper/utils"
)

func TestInvalidURL(t *testing.T) {
	cases := []struct {
		name  string
		value string
		want  bool
	}{
		{"empty string", "", true},
		{"root slash", "/", true},
		{"valid url", "https://example.com/Patient", false},
		{"relative path", "/Patient", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := utils.InvalidURL(tc.value)
			if got != tc.want {
				t.Errorf("InvalidURL(%q) = %v, want %v", tc.value, got, tc.want)
			}
		})
	}
}

func TestQueryParams(t *testing.T) {
	params := map[string]string{
		"name": "budi",
	}

	got := utils.QueryParams(params)
	want := "name=budi"

	if got != want {
		t.Errorf("QueryParams() = %q, want %q", got, want)
	}
}

func TestQueryParams_Empty(t *testing.T) {
	got := utils.QueryParams(map[string]string{})
	if got != "" {
		t.Errorf("QueryParams(empty) = %q, want empty string", got)
	}
}

func TestGetHostName(t *testing.T) {
	// Catatan: kasus URL invalid sengaja tidak diuji di sini karena
	// GetHostName memanggil logger.ErrorJson dari webcore-go/webcore pada
	// jalur error, yang membutuhkan inisialisasi runtime aplikasi
	// (core.NewApp) dan akan panic (nil pointer) jika dijalankan sebagai
	// unit test standalone tanpa bootstrap aplikasi.
	cases := []struct {
		name string
		url  string
		want string
	}{
		{"valid https url", "https://api-satusehat.kemkes.go.id/fhir-r4/v1", "api-satusehat.kemkes.go.id"},
		{"url with port", "https://api.ildki.appgo.my.id:8443/dev/fhir/r4/v1", "api.ildki.appgo.my.id"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := utils.GetHostName(tc.url)
			if got != tc.want {
				t.Errorf("GetHostName(%q) = %q, want %q", tc.url, got, tc.want)
			}
		})
	}
}

func TestGetTarget(t *testing.T) {
	cases := []struct {
		name        string
		priority    string
		wantTarget  string
		wantForward string
	}{
		{"local first", "local-first", "ildki", "satusehat"},
		{"satusehat first", "satusehat-first", "satusehat", "ildki"},
		{"unknown priority defaults to satusehat", "unknown-priority", "satusehat", "ildki"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			target, forward := utils.GetTarget(tc.priority)
			if target != tc.wantTarget || forward != tc.wantForward {
				t.Errorf("GetTarget(%q) = (%q, %q), want (%q, %q)", tc.priority, target, forward, tc.wantTarget, tc.wantForward)
			}
		})
	}
}

// buildTestConfig menyediakan konfigurasi minimal untuk pengujian GetUrl.
func buildTestConfig() *config.ModuleConfig {
	return &config.ModuleConfig{
		FhirSource: config.FhirSource{
			Priority: "satusehat-first",
			Header:   "X-FHIR-Source-Priority",
		},
		SatSetDev: config.SatusehatConfig{
			BaseURL: "https://satusehat-dev.example.com",
		},
		SatSetProd: config.SatusehatConfig{
			BaseURL: "https://satusehat-prod.example.com",
		},
		Ildki: config.IldkiConfig{
			DevelopmentURL: "https://ildki-dev.example.com",
			ProductionURL:  "https://ildki-prod.example.com",
		},
	}
}

func TestGetUrl_SatusehatFirstDefault(t *testing.T) {
	cfg := buildTestConfig()

	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		mainUrl, forwardUrl, priority := utils.GetUrl(cfg, "dev", "Patient", c)
		return c.JSON(fiber.Map{"main": mainUrl, "forward": forwardUrl, "priority": priority})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if result["main"] != "https://satusehat-dev.example.com/Patient" {
		t.Errorf("main url = %q, want %q", result["main"], "https://satusehat-dev.example.com/Patient")
	}
	if result["forward"] != "https://ildki-dev.example.com/Patient" {
		t.Errorf("forward url = %q, want %q", result["forward"], "https://ildki-dev.example.com/Patient")
	}
	if result["priority"] != "satusehat-first" {
		t.Errorf("priority = %q, want %q", result["priority"], "satusehat-first")
	}
}

func TestGetUrl_HeaderOverridesToLocalFirst(t *testing.T) {
	cfg := buildTestConfig()

	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		mainUrl, forwardUrl, priority := utils.GetUrl(cfg, "prod", "Encounter", c)
		return c.JSON(fiber.Map{"main": mainUrl, "forward": forwardUrl, "priority": priority})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-FHIR-Source-Priority", "local-first")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if result["main"] != "https://ildki-prod.example.com/Encounter" {
		t.Errorf("main url = %q, want %q", result["main"], "https://ildki-prod.example.com/Encounter")
	}
	if result["forward"] != "https://satusehat-prod.example.com/Encounter" {
		t.Errorf("forward url = %q, want %q", result["forward"], "https://satusehat-prod.example.com/Encounter")
	}
	if result["priority"] != "local-first" {
		t.Errorf("priority = %q, want %q", result["priority"], "local-first")
	}
}
