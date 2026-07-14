package tests

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	miniproxy "github.com/nersus15/mini-proxy/mod-proxy"
)

// Catatan: AuthWithSatuSehatToken hanya memvalidasi FORMAT header Authorization
// dan parameter env pada path. Validasi token yang sesungguhnya sengaja
// diserahkan ke SATUSEHAT (lihat klarifikasi arsitektur: alur kerja selalu
// satusehat-first walau local-first tersedia). Test ini mengunci perilaku
// tersebut agar perubahan yang tidak disengaja pada middleware terdeteksi.
func newAuthTestApp() *fiber.App {
	m := &miniproxy.Module{}
	app := fiber.New()
	app.Get("/api/fhir/:env/:resourceType", m.AuthWithSatuSehatToken, func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
	return app
}

func TestAuthWithSatuSehatToken(t *testing.T) {
	cases := []struct {
		name       string
		path       string
		authHeader string
		wantOK     bool // true jika request diharapkan lolos ke handler berikutnya (200)
	}{
		{"missing authorization header", "/api/fhir/dev/Patient", "", false},
		{"missing bearer prefix", "/api/fhir/dev/Patient", "sometoken", false},
		{"empty bearer token", "/api/fhir/dev/Patient", "Bearer ", false},
		{"invalid env", "/api/fhir/staging/Patient", "Bearer sometoken", false},
		{"valid dev env with well-formed bearer token", "/api/fhir/dev/Patient", "Bearer sometoken", true},
		{"valid prod env with well-formed bearer token", "/api/fhir/prod/Patient", "Bearer sometoken", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			app := newAuthTestApp()
			req := httptest.NewRequest("GET", tc.path, nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("app.Test() error = %v", err)
			}
			defer resp.Body.Close()

			if tc.wantOK && resp.StatusCode != fiber.StatusOK {
				t.Errorf("status = %d, want %d (request seharusnya lolos middleware)", resp.StatusCode, fiber.StatusOK)
			}
			if !tc.wantOK && resp.StatusCode == fiber.StatusOK {
				t.Errorf("status = %d, want bukan 200 (request seharusnya ditolak middleware)", resp.StatusCode)
			}
		})
	}
}
