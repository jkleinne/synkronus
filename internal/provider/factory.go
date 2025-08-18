// File: internal/provider/factory.go
package provider

import (
	"context"
	"fmt"
	"log/slog"
	"synkronus/internal/config"
	"synkronus/pkg/storage"
	"synkronus/pkg/storage/aws"
	"synkronus/pkg/storage/gcp"
)

type Factory struct {
	cfg    *config.Config
	logger *slog.Logger
}

func NewFactory(cfg *config.Config, logger *slog.Logger) *Factory {
	return &Factory{
		cfg:    cfg,
		logger: logger,
	}
}

func (f *Factory) GetStorageProvider(ctx context.Context, providerName string) (storage.Storage, error) {
	providerLogger := f.logger.With("provider", providerName)

	switch providerName {
	case "gcp":
		if f.cfg.GCP == nil || f.cfg.GCP.Project == "" {
			return nil, fmt.Errorf("GCP project not configured. Use 'synkronus config set gcp.project <project-id>'")
		}
		return gcp.NewGCPStorage(ctx, f.cfg.GCP.Project, providerLogger)
	case "aws":
		if f.cfg.AWS == nil || f.cfg.AWS.Region == "" {
			return nil, fmt.Errorf("AWS region not configured. Use 'synkronus config set aws.region <region>'")
		}
		return aws.NewAWSStorage(f.cfg.AWS.Region, providerLogger), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", providerName)
	}
}
