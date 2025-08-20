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
	registryMu.RLock()
	defer registryMu.RUnlock()

	var configuredProviders []string
	for name, registration := range providerRegistry {
		if registration.ConfigCheck(f.cfg) {
			configuredProviders = append(configuredProviders, name)
		}
	}
	sort.Strings(configuredProviders)
	return configuredProviders
}

func (f *Factory) IsConfigured(providerName string) bool {
	registration, exists := getRegistration(providerName)
	if !exists {
		return false
	}
	return registration.ConfigCheck(f.cfg)
}

// Initializes and returns the storage client for the specified provider
func (f *Factory) GetStorageProvider(ctx context.Context, providerName string) (storage.Storage, error) {
	normalizedName := strings.ToLower(providerName)
	providerLogger := f.logger.With("provider", normalizedName)

	registration, exists := getRegistration(normalizedName)

	if !exists {
		return nil, fmt.Errorf("unsupported provider: %s. Supported providers are: %v", providerName, GetSupportedProviders())
	}

	if !registration.ConfigCheck(f.cfg) {
		return nil, fmt.Errorf("provider '%s' is not configured. Use 'synkronus config set %s.<key> <value>' (e.g., 'gcp.project' or 'aws.region')", normalizedName, normalizedName)
	}

	// Dynamically initialize the provider using the registered initializer function
	client, err := registration.Initializer(ctx, f.cfg, providerLogger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize provider %s: %w", normalizedName, err)
	}

	return client, nil
}
