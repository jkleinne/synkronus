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
	"synkronus/pkg/storage"
)

// Defines the function signature for checking if a provider is configured
type ProviderConfigCheck func(cfg *config.Config) bool

// Defines the function signature for creating a new storage provider client
type ProviderInitializer func(ctx context.Context, cfg *config.Config, logger *slog.Logger) (storage.Storage, error)

// Holds the necessary functions to check configuration and initialize a provider
type ProviderRegistration struct {
	ConfigCheck ProviderConfigCheck
	Initializer ProviderInitializer
}

var (
	// Stores the registrations, keyed by the provider name (lowercase)
	providerRegistry = make(map[string]ProviderRegistration)
	registryMu       sync.RWMutex
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

// Checks if a provider name has been registered
func IsSupported(providerName string) bool {
	registryMu.RLock()
	defer registryMu.RUnlock()

	_, exists := providerRegistry[strings.ToLower(providerName)]
	return exists
}

// Retrieves the registration details for a provider
func GetRegistration(providerName string) (ProviderRegistration, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()

	registration, exists := providerRegistry[strings.ToLower(providerName)]
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
