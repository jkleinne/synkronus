// File: internal/provider/factory.go
package provider

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"
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

func (f *Factory) GetConfiguredProviders() []string {
	var configuredProviders []string
	for name, check := range providerRegistry {
		if check(f.cfg) {
			configuredProviders = append(configuredProviders, name)
		}
	}
	sort.Strings(configuredProviders)
	return configuredProviders
}

func (f *Factory) IsConfigured(providerName string) bool {
	check, exists := getConfigurationCheck(providerName)
	if !exists {
		return false
	}
	return check(f.cfg)
}

func (f *Factory) GetStorageProvider(ctx context.Context, providerName string) (storage.Storage, error) {
	normalizedName := strings.ToLower(providerName)
	providerLogger := f.logger.With("provider", normalizedName)

	if !IsSupported(normalizedName) {
		return nil, fmt.Errorf("unsupported provider: %s. Supported providers are: %v", providerName, GetSupportedProviders())
	}

	if !f.IsConfigured(normalizedName) {
		return nil, fmt.Errorf("provider '%s' is not configured. Use 'synkronus config set %s.<key> <value>' (e.g., 'gcp.project' or 'aws.region')", normalizedName, normalizedName)
	}

	switch normalizedName {
	case "gcp":
		return gcp.NewGCPStorage(ctx, f.cfg.GCP.Project, providerLogger)
	case "aws":
		return aws.NewAWSStorage(f.cfg.AWS.Region, providerLogger), nil
	default:
		return nil, fmt.Errorf("internal error: provider %s supported but not implemented in factory", normalizedName)
	}
}
