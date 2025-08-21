// File: pkg/storage/gcp/client.go
package gcp

import (
	"context"
	"fmt"
	"log/slog"
	"synkronus/internal/config"
	"synkronus/internal/provider/registry"
	"synkronus/pkg/common"
	"synkronus/pkg/storage"

	gcpstorage "cloud.google.com/go/storage"
)

func init() {
	registry.RegisterProvider("gcp", registry.ProviderRegistration{
		ConfigCheck: isConfigured,
		Initializer: initialize,
	})
}

// Checks if the GCP configuration block is present and the project ID is set
func isConfigured(cfg *config.Config) bool {
	return cfg.GCP != nil && cfg.GCP.Project != ""
}

// Initializes the GCP storage client from the configuration
func initialize(ctx context.Context, cfg *config.Config, logger *slog.Logger) (storage.Storage, error) {
	if !isConfigured(cfg) {
		return nil, fmt.Errorf("GCP configuration missing or incomplete")
	}
	return NewGCPStorage(ctx, cfg.GCP.Project, logger)
}

type GCPStorage struct {
	client    *gcpstorage.Client
	projectID string
	logger    *slog.Logger
}

var _ storage.Storage = (*GCPStorage)(nil)

func NewGCPStorage(ctx context.Context, projectID string, logger *slog.Logger) (*GCPStorage, error) {
	client, err := gcpstorage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCP storage client: %w", err)
	}

	return &GCPStorage{
		client:    client,
		projectID: projectID,
		logger:    logger,
	}, nil
}

func (g *GCPStorage) ProviderName() common.Provider {
	return common.GCP
}

func (g *GCPStorage) Close() error {
	if g.client != nil {
		return g.client.Close()
	}
	return nil
}
