// File: internal/provider/factory/factory.go
package factory

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"synkronus/internal/config"
	"synkronus/internal/domain/sql"
	"synkronus/internal/domain/storage"
	"synkronus/internal/provider/registry"
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
	return getConfigured(registry.GetAllRegistrations(), f.cfg)
}

// Returns a list of SQL providers that are registered and configured
func (f *Factory) GetConfiguredSqlProviders() []string {
	return getConfigured(registry.GetAllSqlRegistrations(), f.cfg)
}

func getConfigured[T any](regs map[string]registry.Registration[T], cfg *config.Config) []string {
	var configured []string
	for name, reg := range regs {
		if reg.ConfigCheck(cfg) {
			configured = append(configured, name)
		}
	}
	slices.Sort(configured)
	return configured
}

// getProvider is a generic helper that looks up, validates, and initializes any registered provider.
// serviceLabel is used in error messages (e.g. "storage", "SQL").
func getProvider[T any](
	ctx context.Context,
	name string,
	getRegistration func(string) (registry.Registration[T], bool),
	getSupportedNames func() []string,
	cfg *config.Config,
	logger *slog.Logger,
	serviceLabel string,
) (T, error) {
	var zero T
	normalizedName := strings.ToLower(name)
	providerLogger := logger.With("provider", normalizedName)

	registration, exists := getRegistration(normalizedName)
	if !exists {
		return zero, fmt.Errorf("unsupported %s provider: %s. Supported %s providers are: %v", serviceLabel, name, serviceLabel, getSupportedNames())
	}

	if !registration.ConfigCheck(cfg) {
		return zero, fmt.Errorf("%s provider '%s' is not configured. Use 'synkronus config set %s.<key> <value>'", serviceLabel, normalizedName, normalizedName)
	}

	client, err := registration.Initializer(ctx, cfg, providerLogger)
	if err != nil {
		return zero, fmt.Errorf("failed to initialize %s provider %s: %w", serviceLabel, normalizedName, err)
	}

	return client, nil
}

// isProviderConfigured is a generic helper that checks whether a named provider is registered and configured.
func isProviderConfigured[T any](name string, getRegistration func(string) (registry.Registration[T], bool), cfg *config.Config) bool {
	registration, exists := getRegistration(name)
	return exists && registration.ConfigCheck(cfg)
}

// Checks if a specific provider is registered and configured
func (f *Factory) IsConfigured(providerName string) bool {
	return isProviderConfigured(providerName, registry.GetRegistration, f.cfg)
}

// Checks if a specific SQL provider is registered and configured
func (f *Factory) IsSqlConfigured(providerName string) bool {
	return isProviderConfigured(providerName, registry.GetSqlRegistration, f.cfg)
}

// Initializes and returns the storage client for the specified provider
func (f *Factory) GetStorageProvider(ctx context.Context, name string) (storage.Storage, error) {
	return getProvider(ctx, name, registry.GetRegistration, registry.GetSupportedProviders, f.cfg, f.logger, "storage")
}

// Initializes and returns the SQL client for the specified provider
func (f *Factory) GetSqlProvider(ctx context.Context, name string) (sql.SQL, error) {
	return getProvider(ctx, name, registry.GetSqlRegistration, registry.GetSupportedSqlProviders, f.cfg, f.logger, "SQL")
}
