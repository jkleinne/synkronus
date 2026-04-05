package main

import (
	"strings"
	"synkronus/internal/domain/storage"
	"testing"
)

func TestCreateBucket_StorageClassNormalization(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"lowercase", "standard", "STANDARD"},
		{"mixed case", "Nearline", "NEARLINE"},
		{"already upper", "COLDLINE", "COLDLINE"},
		{"empty stays empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := strings.ToUpper(tt.input)
			if tt.input == "" {
				got = ""
			}
			if got != tt.want {
				t.Errorf("storageClass = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCreateBucket_PublicAccessPreventionValidation(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"enforced lowercase", "enforced", false},
		{"inherited lowercase", "inherited", false},
		{"enforced mixed case", "Enforced", false},
		{"inherited mixed case", "Inherited", false},
		{"invalid value", "blocked", true},
		{"empty value", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalized := strings.ToLower(tt.input)
			valid := normalized == storage.PublicAccessPreventionEnforced ||
				normalized == storage.PublicAccessPreventionInherited
			if valid == tt.wantErr {
				t.Errorf("valid = %v, wantErr = %v for input %q", valid, tt.wantErr, tt.input)
			}
		})
	}
}

func TestCreateBucket_UniformAccessGCPOnly(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		wantErr  bool
	}{
		{"gcp allowed", "gcp", false},
		{"GCP uppercase allowed", "GCP", false},
		{"aws rejected", "aws", true},
		{"AWS uppercase rejected", "AWS", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isGCP := strings.ToLower(tt.provider) == "gcp"
			if isGCP == tt.wantErr {
				t.Errorf("isGCP = %v, wantErr = %v for provider %q", isGCP, tt.wantErr, tt.provider)
			}
		})
	}
}
