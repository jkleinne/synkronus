// File: internal/provider/registry/registry.go
package registry

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"sync"
	"synkronus/internal/config"
	"synkronus/pkg/sql"
	"synkronus/pkg/storage"
)

// Defines the function signature for checking if a provider is configured
type ProviderConfigCheck func(cfg *config.Config) bool

// Defines the function signature for creating a new storage provider client
type ProviderInitializer func(ctx context.Context, cfg *config.Config, logger *slog.Logger) (storage.Storage, error)

// Holds the necessary functions to check configuration and initialize a storage provider
type ProviderRegistration struct {
	ConfigCheck ProviderConfigCheck
	Initializer ProviderInitializer
}

// Defines the function signature for creating a new SQL provider client
type SqlProviderInitializer func(ctx context.Context, cfg *config.Config, logger *slog.Logger) (sql.SQL, error)

// Holds the necessary functions to check configuration and initialize a SQL provider
type SqlProviderRegistration struct {
	ConfigCheck    ProviderConfigCheck
	SqlInitializer SqlProviderInitializer
}

var (
	// Stores the storage registrations, keyed by the provider name (lowercase)
	providerRegistry = make(map[string]ProviderRegistration)
	// Stores the SQL registrations, keyed by the provider name (lowercase)
	sqlProviderRegistry = make(map[string]SqlProviderRegistration)
	registryMu          sync.RWMutex
)

// Allows a provider implementation package to register itself during initialization (init())
func RegisterProvider(name string, registration ProviderRegistration) {
	registryMu.Lock()
	defer registryMu.Unlock()

	normalizedName := strings.ToLower(name)
	if _, exists := providerRegistry[normalizedName]; exists {
		panic(fmt.Sprintf("provider %s already registered", normalizedName))
	}

	if registration.ConfigCheck == nil {
		panic(fmt.Sprintf("provider %s registration missing ConfigCheck", normalizedName))
	}
	if registration.Initializer == nil {
		panic(fmt.Sprintf("provider %s registration missing Initializer", normalizedName))
	}

	providerRegistry[normalizedName] = registration
}

// Allows a SQL provider implementation package to register itself during initialization (init())
func RegisterSqlProvider(name string, registration SqlProviderRegistration) {
	registryMu.Lock()
	defer registryMu.Unlock()

	normalizedName := strings.ToLower(name)
	if _, exists := sqlProviderRegistry[normalizedName]; exists {
		panic(fmt.Sprintf("sql provider %s already registered", normalizedName))
	}

	if registration.ConfigCheck == nil {
		panic(fmt.Sprintf("sql provider %s registration missing ConfigCheck", normalizedName))
	}
	if registration.SqlInitializer == nil {
		panic(fmt.Sprintf("sql provider %s registration missing SqlInitializer", normalizedName))
	}

	sqlProviderRegistry[normalizedName] = registration
}

// Returns a sorted list of all registered provider names
func GetSupportedProviders() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()

	providers := make([]string, 0, len(providerRegistry))
	for name := range providerRegistry {
		providers = append(providers, name)
	}
	sort.Strings(providers)
	return providers
}

// Returns a sorted list of all registered SQL provider names
func GetSupportedSqlProviders() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()

	providers := make([]string, 0, len(sqlProviderRegistry))
	for name := range sqlProviderRegistry {
		providers = append(providers, name)
	}
	sort.Strings(providers)
	return providers
}

// Checks if a provider name has been registered
func IsSupported(providerName string) bool {
	registryMu.RLock()
	defer registryMu.RUnlock()

	_, exists := providerRegistry[strings.ToLower(providerName)]
	return exists
}

// Checks if a SQL provider name has been registered
func IsSqlSupported(providerName string) bool {
	registryMu.RLock()
	defer registryMu.RUnlock()

	_, exists := sqlProviderRegistry[strings.ToLower(providerName)]
	return exists
}

// Retrieves the registration details for a provider
func GetRegistration(providerName string) (ProviderRegistration, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()

	registration, exists := providerRegistry[strings.ToLower(providerName)]
	return registration, exists
}

// Retrieves the SQL registration details for a provider
func GetSqlRegistration(providerName string) (SqlProviderRegistration, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()

	registration, exists := sqlProviderRegistry[strings.ToLower(providerName)]
	return registration, exists
}

// Returns a copy of the entire registry map (primarily for use by the factory)
func GetAllRegistrations() map[string]ProviderRegistration {
	registryMu.RLock()
	defer registryMu.RUnlock()

	// Return a copy to ensure thread safety without requiring the caller to manage the mutex
	registrations := make(map[string]ProviderRegistration, len(providerRegistry))
	for k, v := range providerRegistry {
		registrations[k] = v
	}
	return registrations
}

// Returns a copy of the entire SQL registry map (primarily for use by the factory)
func GetAllSqlRegistrations() map[string]SqlProviderRegistration {
	registryMu.RLock()
	defer registryMu.RUnlock()

	registrations := make(map[string]SqlProviderRegistration, len(sqlProviderRegistry))
	for k, v := range sqlProviderRegistry {
		registrations[k] = v
	}
	return registrations
}
