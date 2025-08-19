// File: internal/provider/registry.go
package provider

import (
	"sort"
	"strings"
	"synkronus/internal/config"
)

type ProviderConfigCheck func(cfg *config.Config) bool

var providerRegistry = map[string]ProviderConfigCheck{
	"gcp": func(cfg *config.Config) bool {
		return cfg.GCP != nil && cfg.GCP.Project != ""
	},
	"aws": func(cfg *config.Config) bool {
		return cfg.AWS != nil && cfg.AWS.Region != ""
	},
}

func GetSupportedProviders() []string {
	providers := make([]string, 0, len(providerRegistry))
	for name := range providerRegistry {
		providers = append(providers, name)
	}
	sort.Strings(providers)
	return providers
}

func IsSupported(providerName string) bool {
	_, exists := providerRegistry[strings.ToLower(providerName)]
	return exists
}

func getConfigurationCheck(providerName string) (ProviderConfigCheck, bool) {
	check, exists := providerRegistry[strings.ToLower(providerName)]
	return check, exists
}
