// File: pkg/storage/gcp/client.go
package gcp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"synkronus/internal/config"
	"synkronus/internal/domain"
	"synkronus/internal/domain/storage"
	"synkronus/internal/provider/registry"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	gcpstorage "cloud.google.com/go/storage"
)

func init() {
	registry.RegisterProvider("gcp", registry.Registration[storage.Storage]{
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
	client           *gcpstorage.Client
	projectID        string
	logger           *slog.Logger
	monitoringClient *monitoring.MetricClient
	monitoringOnce   sync.Once
	monitoringErr    error
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

func (g *GCPStorage) ProviderName() domain.Provider {
	return domain.GCP
}

func (g *GCPStorage) getMonitoringClient(ctx context.Context) (*monitoring.MetricClient, error) {
	g.monitoringOnce.Do(func() {
		g.monitoringClient, g.monitoringErr = monitoring.NewMetricClient(ctx)
	})
	return g.monitoringClient, g.monitoringErr
}

func (g *GCPStorage) Close() error {
	var storageErr, monitoringErr error
	if g.client != nil {
		storageErr = g.client.Close()
	}
	if g.monitoringClient != nil {
		monitoringErr = g.monitoringClient.Close()
	}
	return errors.Join(storageErr, monitoringErr)
}
