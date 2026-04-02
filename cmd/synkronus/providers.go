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

	for _, p := range requested {
		p = strings.ToLower(strings.TrimSpace(p))
		if seen[p] {
			continue
		}
		seen[p] = true

		if r.IsSupported(p) {
			if r.IsConfigured(p) {
				validated = append(validated, p)
			} else {
				return nil, fmt.Errorf("%s provider '%s' was requested but is not configured. Use 'synkronus config set %s.<key> <value>'", r.Label, p, p)
			}
		} else {
			unsupported = append(unsupported, p)
		}
	}

	if len(unsupported) > 0 {
		return nil, fmt.Errorf("unsupported %s providers requested: %v. Supported %s providers are: %v", r.Label, unsupported, r.Label, r.GetSupported())
	}

	return validated, nil
}
