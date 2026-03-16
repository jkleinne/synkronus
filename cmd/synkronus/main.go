// File: cmd/synkronus/main.go
package main

import (
	// Explicitly import the provider registration package to ensure all provider init() functions run
	// and they register themselves (both storage and SQL providers)
	_ "synkronus/internal/provider"
)

func main() {
	// The application container (including the logger) is initialized within the root command's
	// PersistentPreRunE hook, ensuring flags (like --debug) are parsed first
	Execute()
}
