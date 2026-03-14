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
	"synkronus/pkg/sql"
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

// Returns a list of SQL providers that are registered and configured
func (f *Factory) GetConfiguredSqlProviders() []string {
	var configuredProviders []string
	allRegistrations := registry.GetAllSqlRegistrations()

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

// Checks if a specific SQL provider is registered and configured
func (f *Factory) IsSqlConfigured(providerName string) bool {
	registration, exists := registry.GetSqlRegistration(providerName)
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

// Initializes and returns the SQL client for the specified provider
func (f *Factory) GetSqlProvider(ctx context.Context, providerName string) (sql.SQL, error) {
	normalizedName := strings.ToLower(providerName)
	providerLogger := f.logger.With("provider", normalizedName)

	registration, exists := registry.GetSqlRegistration(normalizedName)

	if !exists {
		return nil, fmt.Errorf("unsupported SQL provider: %s. Supported SQL providers are: %v", providerName, registry.GetSupportedSqlProviders())
	}

	if !registration.ConfigCheck(f.cfg) {
		return nil, fmt.Errorf("SQL provider '%s' is not configured. Use 'synkronus config set %s.<key> <value>' (e.g., 'gcp.project')", normalizedName, normalizedName)
	}

	// Dynamically initialize the SQL provider using the registered initializer function
	client, err := registration.SqlInitializer(ctx, f.cfg, providerLogger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize SQL provider %s: %w", normalizedName, err)
	}

	return client, nil
}
