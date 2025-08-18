// File: internal/logger/logger.go
package logger

import (
	"log/slog"
	"os"
)

func NewLogger() *slog.Logger {
	// TODO: Allow configuring log level (e.g., Debug) via environment variables or flags
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	handler := slog.NewTextHandler(os.Stdout, opts)

	logger := slog.New(handler)

	slog.SetDefault(logger)
	return logger
}
