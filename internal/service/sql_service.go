// File: internal/service/sql_service.go
package service

import (
	"context"
	"fmt"
	"log/slog"

	"synkronus/internal/provider/factory"
	"synkronus/internal/domain/sql"
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

	return concurrentFanOut(
		ctx,
		providerNames,
		s.providerFactory.GetSqlProvider,
		func(ctx context.Context, client sql.SQL) ([]sql.Instance, error) {
			return client.ListInstances(ctx)
		},
		s.logger,
	)
}

// DescribeInstance returns detailed information about a specific SQL instance from a single provider
func (s *SqlService) DescribeInstance(ctx context.Context, instanceName, providerName string) (sql.Instance, error) {
	s.logger.Debug("Starting DescribeInstance operation", "instance", instanceName, "provider", providerName)
	return withClientResult(ctx, s.getSqlClient, providerName, func(client sql.SQL) (sql.Instance, error) {
		instance, err := client.DescribeInstance(ctx, instanceName)
		if err != nil {
			return sql.Instance{}, fmt.Errorf("describing SQL instance %q on %s: %w", instanceName, providerName, err)
		}
		return instance, nil
	})
}

func (s *SqlService) getSqlClient(ctx context.Context, providerName string) (sql.SQL, error) {
	client, err := s.providerFactory.GetSqlProvider(ctx, providerName)
	if err != nil {
		return nil, fmt.Errorf("initializing SQL provider %s: %w", providerName, err)
	}
	return client, nil
}
