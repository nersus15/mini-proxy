package tests

import (
	"testing"

	"github.com/nersus15/mini-proxy/mod-proxy/service"
)

func TestReplaceHostname(t *testing.T) {
	cases := []struct {
		name    string
		rawURL  string
		newHost string
		want    string
		wantErr bool
	}{
		{
			name:    "replace host without port",
			rawURL:  "https://api.ildki.appgo.my.id/api/dev/backup",
			newHost: "api-debug.ildki.appgo.my.id",
			want:    "https://api-debug.ildki.appgo.my.id/api/dev/backup",
		},
		{
			name:    "replace host with existing port",
			rawURL:  "https://api.ildki.appgo.my.id:8443/api/dev/backup",
			newHost: "api-debug.ildki.appgo.my.id",
			want:    "https://api-debug.ildki.appgo.my.id:8443/api/dev/backup",
		},
		{
			name:    "invalid url",
			rawURL:  "://bad-url",
			newHost: "backup.example.com",
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := service.ReplaceHostname(tc.rawURL, tc.newHost)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil (result=%q)", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("ReplaceHostname(%q, %q) = %q, want %q", tc.rawURL, tc.newHost, got, tc.want)
			}
		})
	}
}
