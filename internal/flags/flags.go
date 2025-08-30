// File: internal/flags/flags.go
package flags

// Centralized definitions for CLI flags used across the application

const (
	// Provider flags are used when an operation targets a single, specific provider (e.g., describe, create, delete)
	Provider      = "provider"
	ProviderShort = "p"

	// Providers (plural) flags are used when an operation can target multiple providers (e.g., list)
	// Note: In this application, 'p' is reused for both singular and plural provider flags depending on the subcommand context
	Providers      = "providers"
	ProvidersShort = "p"

	// Location flags are used to specify the geographical location or region for resource creation.
	Location      = "location"
	LocationShort = "l"

	// Bucket flags are used to specify the target bucket for object-level operations
	Bucket      = "bucket"
	BucketShort = "b"

	// Prefix flags are used to filter object listings
	Prefix = "prefix"

	// Force flags are used to bypass interactive confirmation prompts for destructive operations
	Force      = "force"
	ForceShort = "f"

	// Debug flags are used to enable verbose logging
	Debug      = "debug"
	DebugShort = "d"
)
