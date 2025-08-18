// File: cmd/synkronus/app.go
package main

import (
	"synkronus/internal/config"
	"synkronus/internal/provider"
	"synkronus/internal/service"
	"synkronus/pkg/formatter"
)

// appContainer holds all the shared dependencies for the application
// This includes configuration, service clients, and formatters
type appContainer struct {
	Config           *config.Config
	ProviderFactory  *provider.Factory
	StorageService   *service.StorageService
	StorageFormatter *formatter.StorageFormatter
}

// Creates and initializes a new application container
// It loads the configuration and sets up all the necessary services
func newApp() (*appContainer, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}

	providerFactory := provider.NewFactory(cfg)
	storageService := service.NewStorageService(providerFactory)
	storageFormatter := formatter.NewStorageFormatter()

	return &appContainer{
		Config:           cfg,
		ProviderFactory:  providerFactory,
		StorageService:   storageService,
		StorageFormatter: storageFormatter,
	}, nil
}
