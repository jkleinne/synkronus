// File: cmd/synkronus/main.go
package main

import (
	"os"
	"synkronus/internal/logger"

	// Explicitly import provider implementations to ensure their init() functions run and they register themselves
	_ "synkronus/pkg/storage/aws"
	_ "synkronus/pkg/storage/gcp"
)

func main() {
	log := logger.NewLogger()

	app, err := newApp(log)
	if err != nil {
		log.Error("Failed to initialize application", "error", err)
		os.Exit(1)
	}
	Execute(app)
}
