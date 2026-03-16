// File: pkg/sql/gcp/gcp.go
package gcp

import (
	"context"
	"fmt"
	"log/slog"
	"synkronus/internal/config"
	"synkronus/internal/provider/registry"
	"synkronus/pkg/common"
	"synkronus/pkg/sql"
	"time"

	sqladmin "google.golang.org/api/sqladmin/v1"
)

func init() {
	registry.RegisterSqlProvider("gcp", registry.SqlProviderRegistration{
		ConfigCheck:    isConfigured,
		SqlInitializer: initialize,
	})
}

// Checks if the GCP configuration block is present and the project ID is set
func isConfigured(cfg *config.Config) bool {
	return cfg.GCP != nil && cfg.GCP.Project != ""
}

// Initializes the GCP SQL client from the configuration
func initialize(ctx context.Context, cfg *config.Config, logger *slog.Logger) (sql.SQL, error) {
	if !isConfigured(cfg) {
		return nil, fmt.Errorf("GCP configuration missing or incomplete")
	}
	return NewGCPSQL(ctx, cfg.GCP.Project, logger)
}

// GCPSql implements the sql.SQL interface for Google Cloud SQL
type GCPSql struct {
	service   *sqladmin.Service
	projectID string
	logger    *slog.Logger
}

var _ sql.SQL = (*GCPSql)(nil)

func NewGCPSQL(ctx context.Context, projectID string, logger *slog.Logger) (*GCPSql, error) {
	svc, err := sqladmin.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCP SQL Admin client: %w", err)
	}

	return &GCPSql{
		service:   svc,
		projectID: projectID,
		logger:    logger,
	}, nil
}

func (g *GCPSql) ProviderName() common.Provider {
	return common.GCP
}

func (g *GCPSql) ListInstances(ctx context.Context) ([]sql.Instance, error) {
	g.logger.Debug("Starting GCP ListInstances operation")

	resp, err := g.service.Instances.List(g.projectID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("error listing Cloud SQL instances: %w", err)
	}

	var instances []sql.Instance
	for _, dbInstance := range resp.Items {
		instances = append(instances, mapInstance(dbInstance, g.projectID))
	}

	return instances, nil
}

func (g *GCPSql) DescribeInstance(ctx context.Context, instanceName string) (sql.Instance, error) {
	g.logger.Debug("Starting GCP DescribeInstance operation", "instance", instanceName)

	dbInstance, err := g.service.Instances.Get(g.projectID, instanceName).Context(ctx).Do()
	if err != nil {
		return sql.Instance{}, fmt.Errorf("error getting Cloud SQL instance '%s': %w", instanceName, err)
	}

	return mapInstance(dbInstance, g.projectID), nil
}

func (g *GCPSql) Close() error {
	// The sqladmin.Service uses an HTTP client that doesn't require explicit closing
	return nil
}

// mapInstance converts a GCP Cloud SQL DatabaseInstance to the domain model
func mapInstance(dbInstance *sqladmin.DatabaseInstance, projectID string) sql.Instance {
	instance := sql.Instance{
		Name:            dbInstance.Name,
		Provider:        common.GCP,
		Region:          dbInstance.Region,
		DatabaseVersion: dbInstance.DatabaseVersion,
		State:           dbInstance.State,
		Project:         projectID,
		ConnectionName:  dbInstance.ConnectionName,
	}

	// Map fields from settings (may be nil for deleted or failed instances)
	if dbInstance.Settings != nil {
		instance.Labels = dbInstance.Settings.UserLabels
		instance.Tier = dbInstance.Settings.Tier
		instance.StorageSizeGB = dbInstance.Settings.DataDiskSizeGb
	}

	// Map primary IP address
	if len(dbInstance.IpAddresses) > 0 {
		instance.PrimaryAddress = dbInstance.IpAddresses[0].IpAddress
	}

	// Parse creation timestamp
	if dbInstance.CreateTime != "" {
		t, err := time.Parse(time.RFC3339, dbInstance.CreateTime)
		if err == nil {
			instance.CreatedAt = t
		}
	}

	return instance
}
