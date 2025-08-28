// File: cmd/synkronus/main.go
package main

import (
	// Explicitly import provider implementations to ensure their init() functions run and they register themselves
	_ "synkronus/pkg/storage/aws"
	_ "synkronus/pkg/storage/gcp"
)

func main() {
	// The application container (including the logger) is initialized within the root command's
	// PersistentPreRunE hook, ensuring flags (like --debug) are parsed first
	Execute()
}
