package shared

import "strings"

// ProviderCapabilities maps provider names (lowercase) to the set of
// create-bucket options they support. This is the single source of truth
// for feature gating at the CLI and TUI layers - no hardcoded provider
// name comparisons should appear outside this map.
var ProviderCapabilities = map[string]map[string]bool{
	"gcp": {
		"uniform-access": true,
	},
	"aws": {},
}

// SupportsOption reports whether a provider supports a given create-bucket
// option key. Both provider and option are matched case-insensitively.
// Returns false for unknown providers or unknown options.
func SupportsOption(provider, option string) bool {
	caps, ok := ProviderCapabilities[strings.ToLower(provider)]
	if !ok {
		return false
	}
	return caps[option]
}
