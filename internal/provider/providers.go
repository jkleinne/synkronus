// File: internal/provider/providers.go
package provider

// This file explicitly imports all provider implementation packages.
// The blank identifier (_) ensures that the init() function of each package runs,
// allowing them to register themselves with the central provider registry.
//
// To add a new provider (e.g., Azure), implement the provider logic in pkg/storage/azure
// ensuring it self-registers in its init() function, and then add the import here.\

import (
	_ "synkronus/pkg/storage/aws"
	_ "synkronus/pkg/storage/gcp"
)
