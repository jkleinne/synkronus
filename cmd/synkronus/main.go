// File: cmd/synkronus/main.go
package main

import (
	"os"
	"synkronus/internal/logger"
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
