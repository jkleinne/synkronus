// File: internal/service/sql_service.go
package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"synkronus/internal/provider/factory"
	"synkronus/pkg/sql"
)

type SqlService struct {
	providerFactory *factory.Factory
	logger          *slog.Logger
}

func NewSqlService(providerFactory *factory.Factory, logger *slog.Logger) *SqlService {
	return &SqlService{
		providerFactory: providerFactory,
		logger:          logger.With("service", "SqlService"),
	}
}

// ListAllInstances queries all specified providers for SQL instances concurrently
func (s *SqlService) ListAllInstances(ctx context.Context, providerNames []string) ([]sql.Instance, error) {
	if len(providerNames) == 0 {
		return nil, nil
	}

	s.logger.Debug("Starting ListAllInstances operation", "providers", providerNames)

	var allInstances []sql.Instance
	var errs []error
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, pName := range providerNames {
		wg.Add(1)
		go func(pName string) {
			defer wg.Done()

			client, err := s.providerFactory.GetSqlProvider(ctx, pName)
			if err != nil {
				s.logger.Error("Failed to initialize SQL provider client", "provider", pName, "error", err)
				mu.Lock()
				errs = append(errs, fmt.Errorf("provider %s: %w", pName, err))
				mu.Unlock()
				return
			}
			defer client.Close()

			instances, err := client.ListInstances(ctx)
			if err != nil {
				s.logger.Error("Failed to list SQL instances from provider", "provider", pName, "error", err)
				mu.Lock()
				errs = append(errs, fmt.Errorf("provider %s: %w", pName, err))
				mu.Unlock()
				return
			}

			mu.Lock()
			allInstances = append(allInstances, instances...)
			mu.Unlock()

			s.logger.Debug("Successfully fetched SQL instances", "provider", pName, "count", len(instances))
		}(pName)
	}

	wg.Wait()

	return allInstances, errors.Join(errs...)
}

// DescribeInstance returns detailed information about a specific SQL instance from a single provider
func (s *SqlService) DescribeInstance(ctx context.Context, instanceName, providerName string) (sql.Instance, error) {
	s.logger.Debug("Starting DescribeInstance operation", "instance", instanceName, "provider", providerName)

	client, err := s.getSqlClient(ctx, providerName)
	if err != nil {
		return sql.Instance{}, err
	}
	defer client.Close()

	instance, err := client.DescribeInstance(ctx, instanceName)
	if err != nil {
		s.logger.Error("Failed to describe SQL instance", "instance", instanceName, "provider", providerName, "error", err)
		return sql.Instance{}, err
	}
	return instance, nil
}

// Helper to initialize the SQL client and handle common error logging
func (s *SqlService) getSqlClient(ctx context.Context, providerName string) (sql.SQL, error) {
	client, err := s.providerFactory.GetSqlProvider(ctx, providerName)
	if err != nil {
		s.logger.Error("Failed to initialize SQL provider", "provider", providerName, "error", err)
		return nil, fmt.Errorf("error initializing SQL provider: %w", err)
	}
	return client, nil
}
