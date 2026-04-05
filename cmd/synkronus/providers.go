package main

import (
	"fmt"
	"strings"
)

// ProviderResolver validates and resolves provider names for list operations.
// It is used by both storage and SQL commands to avoid duplicating the resolution logic.
type ProviderResolver struct {
	IsSupported   func(string) bool
	IsConfigured  func(string) bool
	GetConfigured func() []string
	GetSupported  func() []string
	Label         string
}

// Resolve validates the requested providers and returns a deduplicated, normalized list.
// If no providers are requested, returns all configured providers.
func (r *ProviderResolver) Resolve(requested []string) ([]string, error) {
	if len(requested) == 0 {
		return r.GetConfigured(), nil
	}

	var validated []string
	var unsupported []string
	seen := make(map[string]bool)

	for _, provider := range requested {
		provider = strings.ToLower(strings.TrimSpace(provider))
		if seen[provider] {
			continue
		}
		seen[provider] = true

		if r.IsSupported(provider) {
			if r.IsConfigured(provider) {
				validated = append(validated, provider)
			} else {
				return nil, fmt.Errorf("%s provider '%s' was requested but is not configured. Use 'synkronus config set %s.<key> <value>'", r.Label, provider, provider)
			}
		} else {
			unsupported = append(unsupported, provider)
		}
	}

	if len(unsupported) > 0 {
		return nil, fmt.Errorf("unsupported %s providers requested: %v. Supported %s providers are: %v", r.Label, unsupported, r.Label, r.GetSupported())
	}

	return validated, nil
}
