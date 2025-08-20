// File: cmd/synkronus/app.go
package main

import (
	"log/slog"
	"synkronus/internal/config"
	"synkronus/internal/provider/factory"
	"synkronus/internal/service"
	"synkronus/pkg/formatter"
)

// appContainer holds all the shared dependencies for the application
// This includes configuration, service clients, formatters, and the logger
type appContainer struct {
	Config           *config.Config
	ConfigManager    *config.ConfigManager
	ProviderFactory  *factory.Factory
	StorageService   *service.StorageService
	StorageFormatter *formatter.StorageFormatter
	Logger           *slog.Logger
}

// Creates and initializes a new application container
func newApp(logger *slog.Logger) (*appContainer, error) {
	cfgManager, err := config.NewConfigManager()
	if err != nil {
		return nil, err
	}

	cfg, err := cfgManager.LoadConfig()
	if err != nil {
		return nil, err
	}

	providerFactory := factory.NewFactory(cfg, logger)
	storageService := service.NewStorageService(providerFactory, logger)
	storageFormatter := formatter.NewStorageFormatter()

	return &appContainer{
		Config:           cfg,
		ConfigManager:    cfgManager,
		ProviderFactory:  providerFactory,
		StorageService:   storageService,
		StorageFormatter: storageFormatter,
		Logger:           logger,
	}, nil
}
