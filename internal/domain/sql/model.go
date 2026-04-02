package sql

import (
	"synkronus/internal/domain"
	"time"
)

// Instance represents a managed SQL database instance from a cloud provider
type Instance struct {
	Name            string          `json:"name" yaml:"name"`
	Provider        domain.Provider `json:"provider" yaml:"provider"`
	Region          string          `json:"region" yaml:"region"`
	DatabaseVersion string          `json:"database_version" yaml:"database_version"`
	Tier            string          `json:"tier,omitempty" yaml:"tier,omitempty"`
	State           string          `json:"state" yaml:"state"`
	PrimaryAddress  string          `json:"primary_address,omitempty" yaml:"primary_address,omitempty"`
	StorageSizeGB   int64           `json:"storage_size_gb,omitempty" yaml:"storage_size_gb,omitempty"`
	CreatedAt       time.Time       `json:"created_at,omitempty" yaml:"created_at,omitempty"`

	// GCP-specific fields
	Project        string `json:"project,omitempty" yaml:"project,omitempty"`
	ConnectionName string `json:"connection_name,omitempty" yaml:"connection_name,omitempty"`

	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
}
