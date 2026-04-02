package output

import (
	"strings"
	"testing"
	"time"

	"synkronus/internal/domain"
	domainsql "synkronus/internal/domain/sql"
)

func TestInstanceListView_RenderTable(t *testing.T) {
	instances := InstanceListView{
		{
			Name:            "prod-db",
			Provider:        domain.GCP,
			Region:          "us-central1",
			DatabaseVersion: "POSTGRES_15",
			Tier:            "db-custom-4-16384",
			State:           "RUNNABLE",
			PrimaryAddress:  "10.0.0.1",
			CreatedAt:       time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC),
		},
	}

	result := instances.RenderTable()

	expectedSubstrings := []string{
		"INSTANCE NAME", "PROVIDER", "REGION", "VERSION", "TIER", "STATE", "ADDRESS", "CREATED",
		"prod-db", "GCP", "us-central1", "POSTGRES_15", "db-custom-4-16384", "RUNNABLE", "10.0.0.1", "2025-01-20",
	}

	for _, s := range expectedSubstrings {
		if !strings.Contains(result, s) {
			t.Errorf("expected output to contain %q, got:\n%s", s, result)
		}
	}
}

func TestInstanceListView_Empty(t *testing.T) {
	instances := InstanceListView{}

	result := instances.RenderTable()

	headers := []string{"INSTANCE NAME", "PROVIDER", "REGION", "VERSION", "TIER", "STATE", "ADDRESS", "CREATED"}
	for _, h := range headers {
		if !strings.Contains(result, h) {
			t.Errorf("empty list should still contain header %q, got:\n%s", h, result)
		}
	}
}

func TestInstanceDetailView_RenderTable(t *testing.T) {
	instance := domainsql.Instance{
		Name:            "prod-db",
		Provider:        domain.GCP,
		Region:          "us-central1",
		DatabaseVersion: "POSTGRES_15",
		Tier:            "db-custom-4-16384",
		State:           "RUNNABLE",
		PrimaryAddress:  "10.0.0.1",
		StorageSizeGB:   100,
		Project:         "my-project",
		ConnectionName:  "my-project:us-central1:prod-db",
		CreatedAt:       time.Date(2025, 1, 20, 10, 30, 0, 0, time.UTC),
		Labels:          map[string]string{"env": "production", "team": "backend"},
	}

	view := InstanceDetailView{instance}
	result := view.RenderTable()

	expectedSections := []string{
		"SQL Instance: prod-db",
		"-- Overview --",
		"-- Labels --",
	}

	for _, section := range expectedSections {
		if !strings.Contains(result, section) {
			t.Errorf("expected output to contain section %q, got:\n%s", section, result)
		}
	}

	expectedValues := []string{
		"GCP", "us-central1", "POSTGRES_15", "db-custom-4-16384", "RUNNABLE",
		"10.0.0.1", "100 GB", "my-project", "my-project:us-central1:prod-db",
		"env", "production", "team", "backend",
	}

	for _, v := range expectedValues {
		if !strings.Contains(result, v) {
			t.Errorf("expected output to contain %q, got:\n%s", v, result)
		}
	}
}

func TestInstanceDetailView_NoLabels(t *testing.T) {
	instance := domainsql.Instance{
		Name:            "test-db",
		Provider:        domain.GCP,
		Region:          "us-east1",
		DatabaseVersion: "MYSQL_8_0",
		Tier:            "db-f1-micro",
		State:           "RUNNABLE",
		CreatedAt:       time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC),
	}

	view := InstanceDetailView{instance}
	result := view.RenderTable()

	if strings.Contains(result, "-- Labels --") {
		t.Error("labels section should not appear when no labels exist")
	}

	if !strings.Contains(result, "SQL Instance: test-db") {
		t.Errorf("expected instance name in output, got:\n%s", result)
	}
}
