// File: cmd/synkronus/app.go
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"synkronus/internal/config"
	"synkronus/internal/logger"
	"synkronus/internal/provider/factory"
	"synkronus/internal/service"
	"synkronus/internal/ui/prompt"
	"synkronus/pkg/formatter"
)

// Defines a specific type for the context key to avoid collisions
type contextKey string

const appContextKey contextKey = "appContainer"

// appContainer holds all the shared dependencies for the application
// This includes configuration, service clients, formatters, and the logger
type appContainer struct {
	Config           *config.Config
	ConfigManager    *config.ConfigManager
	ProviderFactory  *factory.Factory
	StorageService   *service.StorageService
	StorageFormatter *formatter.StorageFormatter
	Prompter         prompt.Prompter
	Logger           *slog.Logger
}

// Creates and initializes a new application container based on the debug mode setting
func newApp(debugMode bool) (*appContainer, error) {
	// 1. Initialize the logger first, as it's required by other components
	logLevel := slog.LevelInfo
	if debugMode {
		logLevel = slog.LevelDebug
	}
	log := logger.NewLogger(logLevel)

	// 2. Load configuration
	cfgManager, err := config.NewConfigManager()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize config manager: %w", err)
	}

	cfg, err := cfgManager.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// 3. Initialize factories and services
	providerFactory := factory.NewFactory(cfg, log)
	storageService := service.NewStorageService(providerFactory, log)
	storageFormatter := formatter.NewStorageFormatter()

	// 4. Initialize UI components
	prompter := prompt.NewStandardPrompter(os.Stdin, os.Stdout)

	return &appContainer{
		Config:           cfg,
		ConfigManager:    cfgManager,
		ProviderFactory:  providerFactory,
		StorageService:   storageService,
		StorageFormatter: storageFormatter,
		Prompter:         prompter,
		Logger:           log,
	}, nil
}

// Injects the application container into the given context
func (a *appContainer) ToContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, appContextKey, a)
}

// Retrieves the application container from the context
func appFromContext(ctx context.Context) (*appContainer, error) {
	app, ok := ctx.Value(appContextKey).(*appContainer)
	if !ok {
		return nil, fmt.Errorf("application container not found in context")
	}
	return app, nil
}
