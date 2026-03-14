// File: pkg/sql/model.go
package sql

import (
	"synkronus/pkg/common"
	"time"
)

// Instance represents a managed SQL database instance from a cloud provider
type Instance struct {
	Name            string
	Provider        common.Provider
	Region          string
	DatabaseVersion string
	Tier            string
	State           string
	PrimaryAddress  string
	StorageSizeGB   int64
	CreatedAt       time.Time

	// GCP-specific fields
	Project        string
	ConnectionName string

	Labels map[string]string
}
