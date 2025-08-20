// File: internal/provider/factory/factory.go
package factory

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"synkronus/internal/config"
	"synkronus/internal/provider/registry"
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

// Returns a list of providers that are registered and configured
func (f *Factory) GetConfiguredProviders() []string {
	var configuredProviders []string
	allRegistrations := registry.GetAllRegistrations()

	for name, registration := range allRegistrations {
		if registration.ConfigCheck(f.cfg) {
			configuredProviders = append(configuredProviders, name)
		}
	}
	sort.Strings(configuredProviders)
	return configuredProviders
}

// Checks if a specific provider is registered and configured
func (f *Factory) IsConfigured(providerName string) bool {
	registration, exists := registry.GetRegistration(providerName)
	if !exists {
		return false
	}
	return registration.ConfigCheck(f.cfg)
}

// Initializes and returns the storage client for the specified provider
func (f *Factory) GetStorageProvider(ctx context.Context, providerName string) (storage.Storage, error) {
	normalizedName := strings.ToLower(providerName)
	providerLogger := f.logger.With("provider", normalizedName)

	registration, exists := registry.GetRegistration(normalizedName)

	if !exists {
		return nil, fmt.Errorf("unsupported provider: %s. Supported providers are: %v", providerName, registry.GetSupportedProviders())
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
