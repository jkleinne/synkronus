package main

import (
	"fmt"
	"strings"
)

// resolveProviders validates and resolves a list of provider names for list operations.
// It is used by both storage and SQL commands to avoid duplicating the resolution logic.
func resolveProviders(
	requestedProviders []string,
	isSupported func(string) bool,
	isConfigured func(string) bool,
	getConfigured func() []string,
	getSupportedNames func() []string,
	providerLabel string,
) ([]string, error) {
	if len(requestedProviders) == 0 {
		return getConfigured(), nil
	}

	var validatedProviders []string
	var invalidProviders []string
	seen := make(map[string]bool)

	for _, p := range requestedProviders {
		p = strings.ToLower(strings.TrimSpace(p))

		if seen[p] {
			continue
		}
		seen[p] = true

		if isSupported(p) {
			if isConfigured(p) {
				validatedProviders = append(validatedProviders, p)
			} else {
				return nil, fmt.Errorf("%s provider '%s' was requested but is not configured. Use 'synkronus config set %s.<key> <value>'", providerLabel, p, p)
			}
		} else {
			invalidProviders = append(invalidProviders, p)
		}
	}

	if len(invalidProviders) > 0 {
		return nil, fmt.Errorf("unsupported %s providers requested: %v. Supported %s providers are: %v", providerLabel, invalidProviders, providerLabel, getSupportedNames())
	}

	return validatedProviders, nil
}
