package main

import (
	"strings"
	"testing"
)

func newTestResolver() *ProviderResolver {
	supported := map[string]bool{"gcp": true, "aws": true}
	configured := map[string]bool{"gcp": true}

	return &ProviderResolver{
		IsSupported:   func(name string) bool { return supported[name] },
		IsConfigured:  func(name string) bool { return configured[name] },
		GetConfigured: func() []string { return []string{"gcp"} },
		GetSupported:  func() []string { return []string{"aws", "gcp"} },
		Label:         "storage",
	}
}

func TestResolve_NoProviders_ReturnsConfigured(t *testing.T) {
	r := newTestResolver()
	result, err := r.Resolve(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0] != "gcp" {
		t.Errorf("expected [gcp], got %v", result)
	}
}

func TestResolve_ValidProvider(t *testing.T) {
	r := newTestResolver()
	result, err := r.Resolve([]string{"gcp"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0] != "gcp" {
		t.Errorf("expected [gcp], got %v", result)
	}
}

func TestResolve_UnconfiguredProvider(t *testing.T) {
	r := newTestResolver()
	_, err := r.Resolve([]string{"aws"})
	if err == nil {
		t.Fatal("expected error for unconfigured provider")
	}
	if !strings.Contains(err.Error(), "not configured") {
		t.Errorf("expected 'not configured' in error, got: %v", err)
	}
}

func TestResolve_UnsupportedProvider(t *testing.T) {
	r := newTestResolver()
	_, err := r.Resolve([]string{"azure"})
	if err == nil {
		t.Fatal("expected error for unsupported provider")
	}
	if !strings.Contains(err.Error(), "unsupported") {
		t.Errorf("expected 'unsupported' in error, got: %v", err)
	}
}

func TestResolve_Duplicates(t *testing.T) {
	r := newTestResolver()
	result, err := r.Resolve([]string{"gcp", "GCP", "gcp"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected deduplication to 1 provider, got %d", len(result))
	}
}

func TestResolve_CaseInsensitive(t *testing.T) {
	r := newTestResolver()
	result, err := r.Resolve([]string{"GCP"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0] != "gcp" {
		t.Errorf("expected [gcp], got %v", result)
	}
}
