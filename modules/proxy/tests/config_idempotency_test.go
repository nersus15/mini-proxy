package tests

import (
	"testing"

	"github.com/nersus15/mini-proxy/mod-proxy/config"
	appConfig "github.com/webcore-go/webcore/infra/config"
)

// Regression test untuk bug nyata: idempotency diam-diam nonaktif karena
// config.yaml tidak memuat blok "idempotency", sehingga Enabled jatuh ke zero
// value (false) dan fingerprint selalu kosong.
//
// Default hanya berlaku kalau key-nya terdaftar di SetEnvBindings — viper
// cuma memproses key yang sudah muncul di AllKeys, dan BindEnv yang
// mendaftarkannya saat blok tersebut tidak ada di file config.
func TestIdempotencyDefaultAktifTanpaBlokConfig(t *testing.T) {
	cfg := &config.ModuleConfig{}

	// Dijalankan dari direktori tests yang tidak punya config.yaml, jadi ini
	// meniru kondisi config tanpa blok idempotency sama sekali.
	if err := appConfig.LoadDefaultConfigModule("fhir", cfg); err != nil {
		t.Fatalf("gagal memuat config module: %v", err)
	}

	if !cfg.Idempotency.Enabled {
		t.Error("Idempotency.Enabled = false padahal default seharusnya true — key kemungkinan belum terdaftar di SetEnvBindings")
	}

	if cfg.Idempotency.TTLMinutes != 5 {
		t.Errorf("Idempotency.TTLMinutes = %d, mau 5 (default)", cfg.Idempotency.TTLMinutes)
	}
}

// Default juga harus terdaftar di SetDefaults, bukan hanya mengandalkan zero
// value kebetulan.
func TestIdempotencyTerdaftarDiDefaults(t *testing.T) {
	cfg := &config.ModuleConfig{}

	defaults := cfg.SetDefaults()
	bindings := cfg.SetEnvBindings()

	for _, key := range []string{"module.fhir.idempotency.enabled", "module.fhir.idempotency.ttl_minutes"} {
		if _, ok := defaults[key]; !ok {
			t.Errorf("%s tidak ada di SetDefaults", key)
		}
		if _, ok := bindings[key]; !ok {
			t.Errorf("%s tidak ada di SetEnvBindings — tanpa ini default tidak akan pernah diterapkan", key)
		}
	}
}
