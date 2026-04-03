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

	// StorageClass flags specify the storage class for bucket creation
	StorageClass      = "storage-class"
	StorageClassShort = "s"

	// Labels flags specify key-value labels for resource creation
	Labels = "labels"

	// VersioningFlag flags enable object versioning on bucket creation
	VersioningFlag = "versioning"

	// UniformAccess flags enable Uniform Bucket-Level Access on bucket creation
	UniformAccess = "uniform-access"

	// PublicAccessPreventionFlag flags set public access prevention policy on bucket creation
	PublicAccessPreventionFlag = "public-access-prevention"

	// OutputPath flags specify the file or directory path for download output
	OutputPath = "output-path"

	// Debug flags are used to enable verbose logging
	Debug      = "debug"
	DebugShort = "d"

	// Output flags are used to select the output format (table, json, yaml)
	Output      = "output"
	OutputShort = "o"
)
