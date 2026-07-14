package tests

import (
	"testing"

	"github.com/nersus15/mini-proxy/mod-proxy/helper/types"
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
)

func strPtr(s string) *string { return &s }

func TestIsTemporaryReference(t *testing.T) {
	cases := []struct {
		name string
		ref  *string
		want bool
	}{
		{"nil reference", nil, false},
		{"temporary uuid reference", strPtr("urn:uuid:1234-5678"), true},
		{"permanent reference", strPtr("Patient/1234-5678"), false},
		{"empty string", strPtr(""), false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := types.IsTemporaryReference(tc.ref)
			if got != tc.want {
				t.Errorf("IsTemporaryReference(%v) = %v, want %v", tc.ref, got, tc.want)
			}
		})
	}
}

func TestGetTemporaryReference(t *testing.T) {
	ref := strPtr("urn:uuid:abcd-1234")
	got := types.GetTemporaryReference(ref)
	if got == nil {
		t.Fatal("expected non-nil id for temporary reference")
	}
	if *got != "abcd-1234" {
		t.Errorf("GetTemporaryReference() = %q, want %q", *got, "abcd-1234")
	}

	permanentRef := strPtr("Patient/abcd-1234")
	if got := types.GetTemporaryReference(permanentRef); got != nil {
		t.Errorf("expected nil for permanent reference, got %q", *got)
	}
}

func TestGetReferencePath(t *testing.T) {
	got := types.GetReferencePath("Patient", "1234")
	want := "Patient/1234"
	if got != want {
		t.Errorf("GetReferencePath() = %q, want %q", got, want)
	}
}

func TestGetReferenceID(t *testing.T) {
	cases := []struct {
		name string
		ref  *fhir.Reference
		want *string
	}{
		{"nil reference struct", nil, nil},
		{"nil inner reference", &fhir.Reference{}, nil},
		{"temporary reference returns nil", &fhir.Reference{Reference: strPtr("urn:uuid:1234")}, nil},
		{"malformed reference (no slash)", &fhir.Reference{Reference: strPtr("Patient")}, nil},
		{"valid reference", &fhir.Reference{Reference: strPtr("Patient/9999")}, strPtr("9999")},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := types.GetReferenceID(tc.ref)
			if tc.want == nil {
				if got != nil {
					t.Errorf("GetReferenceID() = %q, want nil", *got)
				}
				return
			}
			if got == nil {
				t.Fatalf("GetReferenceID() = nil, want %q", *tc.want)
			}
			if *got != *tc.want {
				t.Errorf("GetReferenceID() = %q, want %q", *got, *tc.want)
			}
		})
	}
}
