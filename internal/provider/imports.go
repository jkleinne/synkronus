// File: internal/provider/imports.go
package provider

// Blank imports trigger provider self-registration via init() functions.

import (
	// Storage providers
	_ "synkronus/internal/provider/storage/aws"
	_ "synkronus/internal/provider/storage/gcp"

	// SQL providers
	_ "synkronus/internal/provider/sql/gcp"
)
