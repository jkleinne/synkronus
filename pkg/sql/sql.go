// File: pkg/sql/sql.go
package sql

import (
	"context"
	"synkronus/pkg/common"
)

// SQL defines the interface for interacting with managed SQL database instances
// across cloud providers. Each provider implementation must satisfy this interface.
type SQL interface {
	// ListInstances returns all SQL database instances for the configured project/account
	ListInstances(ctx context.Context) ([]Instance, error)

	// DescribeInstance returns detailed information about a specific SQL instance
	DescribeInstance(ctx context.Context, instanceName string) (Instance, error)

	// ProviderName returns the cloud provider identifier
	ProviderName() common.Provider

	// Close cleans up any underlying resources (e.g., HTTP clients)
	Close() error
}
