package ui

import (
	"strings"
	"testing"
	"time"

	domainsql "synkronus/internal/domain/sql"
)

func TestRenderInstanceListWithData(t *testing.T) {
	instances := []domainsql.Instance{
		{Name: "prod-db", Provider: "GCP", Region: "us-central1", DatabaseVersion: "POSTGRES_15", Tier: "db-custom-2-8192", State: "RUNNABLE", CreatedAt: time.Now()},
	}
	result := RenderInstanceList(instances, 0, 0, 80)
	if !strings.Contains(result, "prod-db") {
		t.Error("instance list should contain instance name")
	}
}

func TestRenderInstanceListEmpty(t *testing.T) {
	result := RenderInstanceList(nil, 0, 0, 80)
	if !strings.Contains(result, "No instances") {
		t.Error("empty list should show 'No instances' message")
	}
}

func TestRenderInstanceDetailSections(t *testing.T) {
	instance := domainsql.Instance{
		Name: "prod-db", Provider: "GCP", Region: "us-central1",
		DatabaseVersion: "POSTGRES_15", Tier: "db-custom-2-8192", State: "RUNNABLE",
	}
	result := RenderInstanceDetail(instance, 80)
	if !strings.Contains(result, "Overview") {
		t.Error("detail should contain Overview section")
	}
}

func TestRenderInstanceListShowsHeaders(t *testing.T) {
	instances := []domainsql.Instance{
		{Name: "test-db", Provider: "GCP", Region: "us-east1"},
	}
	result := RenderInstanceList(instances, 0, 0, 120)
	if !strings.Contains(result, "NAME") {
		t.Error("instance list should contain NAME header")
	}
	if !strings.Contains(result, "PROVIDER") {
		t.Error("instance list should contain PROVIDER header")
	}
}

func TestRenderInstanceDetailWithLabels(t *testing.T) {
	instance := domainsql.Instance{
		Name:     "labeled-db",
		Provider: "GCP",
		Labels:   map[string]string{"env": "prod", "team": "platform"},
	}
	result := RenderInstanceDetail(instance, 80)
	if !strings.Contains(result, "Labels") {
		t.Error("detail with labels should contain Labels section")
	}
}

func TestRenderInstanceDetailWithCreatedAt(t *testing.T) {
	createdAt := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
	instance := domainsql.Instance{
		Name:      "dated-db",
		Provider:  "GCP",
		CreatedAt: createdAt,
	}
	result := RenderInstanceDetail(instance, 80)
	if !strings.Contains(result, "2024-03-15") {
		t.Error("detail should contain formatted creation date")
	}
}

func TestRenderInstanceListCreatedDateFormat(t *testing.T) {
	createdAt := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	instances := []domainsql.Instance{
		{Name: "db-with-date", Provider: "GCP", CreatedAt: createdAt},
	}
	result := RenderInstanceList(instances, 0, 0, 120)
	if !strings.Contains(result, "2024-06-01") {
		t.Error("instance list should format created date as YYYY-MM-DD")
	}
}

func TestRenderInstanceListCursorHighlight(t *testing.T) {
	instances := []domainsql.Instance{
		{Name: "db-one", Provider: "GCP"},
		{Name: "db-two", Provider: "GCP"},
	}
	result := RenderInstanceList(instances, 0, 0, 120)
	// Selected row should have the cursor indicator prefix
	if !strings.Contains(result, "▸") {
		t.Error("selected row should have cursor indicator")
	}
}
