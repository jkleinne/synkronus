// File: internal/provider/registry.go
package provider

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

type ProviderConfigCheck func(cfg *config.Config) bool

type ProviderInitializer func(ctx context.Context, cfg *config.Config, logger *slog.Logger) (storage.Storage, error)

type ProviderRegistration struct {
	ConfigCheck ProviderConfigCheck
	Initializer ProviderInitializer
}

var (
	providerRegistry = make(map[string]ProviderRegistration)
	registryMu       sync.RWMutex
)

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

func IsSupported(providerName string) bool {
	registryMu.RLock()
	defer registryMu.RUnlock()

	_, exists := providerRegistry[strings.ToLower(providerName)]
	return exists
}

func getRegistration(providerName string) (ProviderRegistration, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()

	registration, exists := providerRegistry[strings.ToLower(providerName)]
	return registration, exists
}
