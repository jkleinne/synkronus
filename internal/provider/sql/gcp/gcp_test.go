package gcp

import (
	"synkronus/internal/domain"
	domainsql "synkronus/internal/domain/sql"
	"testing"
	"time"

	sqladmin "google.golang.org/api/sqladmin/v1"
)

func TestMapInstance_NilSettings(t *testing.T) {
	dbInstance := &sqladmin.DatabaseInstance{
		Name:            "test-instance",
		Region:          "us-central1",
		DatabaseVersion: "POSTGRES_15",
		State:           "RUNNABLE",
		ConnectionName:  "project:region:test-instance",
		Settings:        nil,
	}

	instance := mapInstance(dbInstance, "test-project")

	if instance.Name != "test-instance" {
		t.Errorf("expected Name 'test-instance', got %q", instance.Name)
	}
	if instance.Provider != domain.GCP {
		t.Errorf("expected Provider GCP, got %v", instance.Provider)
	}
	if instance.Tier != "" {
		t.Errorf("expected empty Tier with nil Settings, got %q", instance.Tier)
	}
	if instance.StorageSizeGB != 0 {
		t.Errorf("expected StorageSizeGB 0 with nil Settings, got %d", instance.StorageSizeGB)
	}
	if instance.Labels != nil {
		t.Errorf("expected nil Labels with nil Settings, got %v", instance.Labels)
	}
}

func TestMapInstance_WithSettings(t *testing.T) {
	dbInstance := &sqladmin.DatabaseInstance{
		Name:            "prod-db",
		Region:          "europe-west1",
		DatabaseVersion: "MYSQL_8_0",
		State:           "RUNNABLE",
		ConnectionName:  "project:region:prod-db",
		Settings: &sqladmin.Settings{
			Tier:           "db-n1-standard-4",
			DataDiskSizeGb: 100,
			UserLabels:     map[string]string{"env": "prod"},
		},
		IpAddresses: []*sqladmin.IpMapping{
			{IpAddress: "10.0.0.1"},
		},
		CreateTime: "2025-01-15T10:30:00Z",
	}

	instance := mapInstance(dbInstance, "my-project")

	expected := domainsql.Instance{
		Name:            "prod-db",
		Provider:        domain.GCP,
		Region:          "europe-west1",
		DatabaseVersion: "MYSQL_8_0",
		State:           "RUNNABLE",
		Project:         "my-project",
		ConnectionName:  "project:region:prod-db",
		Tier:            "db-n1-standard-4",
		StorageSizeGB:   100,
		Labels:          map[string]string{"env": "prod"},
		PrimaryAddress:  "10.0.0.1",
		CreatedAt:       time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	if instance.Name != expected.Name {
		t.Errorf("Name: got %q, want %q", instance.Name, expected.Name)
	}
	if instance.Tier != expected.Tier {
		t.Errorf("Tier: got %q, want %q", instance.Tier, expected.Tier)
	}
	if instance.StorageSizeGB != expected.StorageSizeGB {
		t.Errorf("StorageSizeGB: got %d, want %d", instance.StorageSizeGB, expected.StorageSizeGB)
	}
	if instance.PrimaryAddress != expected.PrimaryAddress {
		t.Errorf("PrimaryAddress: got %q, want %q", instance.PrimaryAddress, expected.PrimaryAddress)
	}
	if !instance.CreatedAt.Equal(expected.CreatedAt) {
		t.Errorf("CreatedAt: got %v, want %v", instance.CreatedAt, expected.CreatedAt)
	}
	if instance.Labels["env"] != "prod" {
		t.Errorf("Labels[env]: got %q, want %q", instance.Labels["env"], "prod")
	}
}
