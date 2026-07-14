package tests

import (
	"testing"

	"github.com/nersus15/mini-proxy/mod-proxy/config"
	"github.com/nersus15/mini-proxy/mod-proxy/helper/utils"
)

func buildTokenTestConfig() *config.ModuleConfig {
	return &config.ModuleConfig{
		SatSetDev: config.SatusehatConfig{
			AuthURL: "https://satusehat-dev.example.com/oauth2/v1/accesstoken",
		},
		SatSetProd: config.SatusehatConfig{
			AuthURL: "https://satusehat-prod.example.com/oauth2/v1/accesstoken",
		},
		Ildki: config.IldkiConfig{
			DevAuthUrl:  "https://ildki-dev.example.com/oauth2/v1/accesstoken",
			ProdAuthUrl: "https://ildki-prod.example.com/oauth2/v1/accesstoken",
		},
	}
}

func TestResolveTokenEndpoint(t *testing.T) {
	cfg := buildTokenTestConfig()

	cases := []struct {
		name   string
		env    string
		target string
		want   string
	}{
		{"dev + ildki", "dev", "ildki", cfg.Ildki.DevAuthUrl},
		{"dev + satusehat", "dev", "satusehat", cfg.SatSetDev.AuthURL},
		// Regresi untuk bug yang sudah diperbaiki: prod + ildki HARUS
		// mengembalikan ProdAuthUrl (endpoint auth), bukan ProductionURL
		// (endpoint resource FHIR).
		{"prod + ildki", "prod", "ildki", cfg.Ildki.ProdAuthUrl},
		{"prod + satusehat", "prod", "satusehat", cfg.SatSetProd.AuthURL},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := utils.ResolveTokenEndpoint(tc.env, tc.target, cfg)
			if got != tc.want {
				t.Errorf("ResolveTokenEndpoint(%q, %q) = %q, want %q", tc.env, tc.target, got, tc.want)
			}
		})
	}
}

func TestResolveTokenEndpoint_ProdIldkiIsNotResourceUrl(t *testing.T) {
	cfg := buildTokenTestConfig()
	cfg.Ildki.ProductionURL = "https://ildki-prod.example.com/fhir/r4/v1" // endpoint resource, BUKAN auth

	got := utils.ResolveTokenEndpoint("prod", "ildki", cfg)
	if got == cfg.Ildki.ProductionURL {
		t.Errorf("ResolveTokenEndpoint salah mengembalikan resource URL (%q) alih-alih auth URL", got)
	}
	if got != cfg.Ildki.ProdAuthUrl {
		t.Errorf("ResolveTokenEndpoint = %q, want ProdAuthUrl %q", got, cfg.Ildki.ProdAuthUrl)
	}
}
