package aws

import (
	"context"
	"log/slog"
	"synkronus/internal/config"
	"testing"
)

func newTestStorage(t *testing.T) *AWSStorage {
	t.Helper()
	s, err := NewAWSStorage(context.Background(), &config.AWSConfig{
		Region:   "us-east-1",
		Endpoint: "http://localhost:4566",
	}, slog.Default())
	if err != nil {
		t.Fatalf("failed to create test storage: %v", err)
	}
	return s
}

func TestIsConfigured(t *testing.T) {
	tests := []struct {
		name string
		cfg  *config.Config
		want bool
	}{
		{"nil config", &config.Config{}, false},
		{"nil aws", &config.Config{AWS: nil}, false},
		{"empty region", &config.Config{AWS: &config.AWSConfig{Region: ""}}, false},
		{"valid", &config.Config{AWS: &config.AWSConfig{Region: "us-east-1"}}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isConfigured(tt.cfg); got != tt.want {
				t.Errorf("isConfigured() = %v, want %v", got, tt.want)
			}
		})
	}
}
