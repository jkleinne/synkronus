package output

import (
	"strings"
	"testing"
	"time"

	"synkronus/internal/domain"
	"synkronus/internal/domain/storage"
)

func TestBucketListView_RenderTable(t *testing.T) {
	buckets := BucketListView{
		{
			Name:         "my-bucket",
			Provider:     domain.GCP,
			Location:     "US-CENTRAL1",
			StorageClass: "STANDARD",
			UsageBytes:   1048576, // 1 MB
			CreatedAt:    time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		},
	}

	result := buckets.RenderTable()

	expectedSubstrings := []string{
		"BUCKET NAME", "PROVIDER", "LOCATION", "USAGE", "STORAGE CLASS", "CREATED",
		"my-bucket", "GCP", "US-CENTRAL1", "1.0 MB", "STANDARD", "2025-01-15",
	}

	for _, s := range expectedSubstrings {
		if !strings.Contains(result, s) {
			t.Errorf("expected output to contain %q, got:\n%s", s, result)
		}
	}
}

func TestBucketListView_Empty(t *testing.T) {
	buckets := BucketListView{}

	result := buckets.RenderTable()

	headers := []string{"BUCKET NAME", "PROVIDER", "LOCATION", "USAGE", "STORAGE CLASS", "CREATED"}
	for _, h := range headers {
		if !strings.Contains(result, h) {
			t.Errorf("empty list should still contain header %q, got:\n%s", h, result)
		}
	}
}

func TestBucketDetailView_RenderTable(t *testing.T) {
	bucket := storage.Bucket{
		Name:         "detail-bucket",
		Provider:     domain.GCP,
		Location:     "US-EAST1",
		LocationType: "region",
		StorageClass: "NEARLINE",
		UsageBytes:   5368709120, // 5 GB
		CreatedAt:    time.Date(2025, 3, 1, 10, 0, 0, 0, time.UTC),
		UpdatedAt:    time.Date(2025, 3, 15, 12, 0, 0, 0, time.UTC),
		Autoclass:    &storage.Autoclass{Enabled: true},
		Versioning:   &storage.Versioning{Enabled: true},
		Encryption:   &storage.Encryption{KmsKeyName: "projects/my-proj/locations/us/keyRings/kr/cryptoKeys/key1"},
		Labels:       map[string]string{"env": "prod", "team": "data"},
		IAMPolicy: &storage.IAMPolicy{
			Bindings: []storage.IAMBinding{
				{Role: "roles/storage.admin", Principals: []string{"user:admin@example.com"}},
			},
		},
		UniformBucketLevelAccess: &storage.UniformBucketLevelAccess{Enabled: true},
		PublicAccessPrevention:   "enforced",
		SoftDeletePolicy:         &storage.SoftDeletePolicy{RetentionDuration: 7 * 24 * time.Hour},
		RetentionPolicy:          &storage.RetentionPolicy{RetentionPeriod: 30 * 24 * time.Hour, IsLocked: true},
		LifecycleRules: []storage.LifecycleRule{
			{Action: "Delete", Condition: storage.LifecycleCondition{Age: 365}},
		},
	}

	view := BucketDetailView{bucket}
	result := view.RenderTable()

	expectedSections := []string{
		"Bucket: detail-bucket",
		"-- Overview --",
		"-- Access Control & Logging --",
		"-- Data Protection --",
		"-- Lifecycle Rules --",
		"-- Labels --",
	}

	for _, section := range expectedSections {
		if !strings.Contains(result, section) {
			t.Errorf("expected output to contain section %q, got:\n%s", section, result)
		}
	}

	// Verify specific field values
	expectedValues := []string{
		"GCP", "US-EAST1", "region", "NEARLINE", "Enabled", // Autoclass
		"projects/my-proj/locations/us/keyRings/kr/cryptoKeys/key1", // CMEK
		"env", "prod", "team", "data", // Labels
		"roles/storage.admin", "user:admin@example.com", // IAM
		"Locked", // Retention
		"Delete", "Age > 365 days", // Lifecycle
	}

	for _, v := range expectedValues {
		if !strings.Contains(result, v) {
			t.Errorf("expected output to contain %q, got:\n%s", v, result)
		}
	}
}

func TestBucketDetailView_NilOptionalFields(t *testing.T) {
	bucket := storage.Bucket{
		Name:         "minimal-bucket",
		Provider:     domain.GCP,
		Location:     "US",
		StorageClass: "STANDARD",
		UsageBytes:   0,
		CreatedAt:    time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:    time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	view := BucketDetailView{bucket}

	// Should not panic with nil optional fields
	result := view.RenderTable()

	if !strings.Contains(result, "minimal-bucket") {
		t.Errorf("expected bucket name in output, got:\n%s", result)
	}

	// Lifecycle and Labels sections should be absent when empty
	if strings.Contains(result, "-- Lifecycle Rules --") {
		t.Error("lifecycle section should not appear when no rules exist")
	}
	if strings.Contains(result, "-- Labels --") {
		t.Error("labels section should not appear when no labels exist")
	}
}

func TestObjectListView_RenderTable(t *testing.T) {
	objectList := storage.ObjectList{
		BucketName: "my-bucket",
		Prefix:     "data/",
		Objects: []storage.Object{
			{
				Key:          "data/file1.csv",
				Size:         2048,
				StorageClass: "STANDARD",
				LastModified: time.Date(2025, 2, 10, 8, 30, 0, 0, time.UTC),
			},
		},
		CommonPrefixes: []string{"data/subdir/"},
	}

	view := ObjectListView{objectList}
	result := view.RenderTable()

	expectedSubstrings := []string{
		"my-bucket",
		"data/",
		"KEY", "SIZE", "STORAGE CLASS", "LAST MODIFIED",
		"data/subdir/", "(DIR)",
		"data/file1.csv", "2.0 KB", "STANDARD",
	}

	for _, s := range expectedSubstrings {
		if !strings.Contains(result, s) {
			t.Errorf("expected output to contain %q, got:\n%s", s, result)
		}
	}
}

func TestObjectListView_Empty(t *testing.T) {
	objectList := storage.ObjectList{
		BucketName: "empty-bucket",
	}

	view := ObjectListView{objectList}
	result := view.RenderTable()

	if !strings.Contains(result, "No objects or directories found") {
		t.Errorf("expected empty message, got:\n%s", result)
	}
}

func TestObjectDetailView_RenderTable(t *testing.T) {
	object := storage.Object{
		Key:          "data/report.pdf",
		Bucket:       "my-bucket",
		Provider:     domain.GCP,
		Size:         1048576,
		StorageClass: "STANDARD",
		LastModified: time.Date(2025, 3, 1, 12, 0, 0, 0, time.UTC),
		CreatedAt:    time.Date(2025, 3, 1, 12, 0, 0, 0, time.UTC),
		ETag:         "abc123",
		ContentType:  "application/pdf",
		Encryption:   &storage.Encryption{KmsKeyName: "projects/p/locations/l/keyRings/kr/cryptoKeys/k", Algorithm: "AES256"},
		CRC32C:       "AABBCC==",
		Generation:   12345,
		Metageneration: 2,
		Metadata:     map[string]string{"author": "test-user"},
	}

	view := ObjectDetailView{object}
	result := view.RenderTable()

	expectedSections := []string{
		"Object: data/report.pdf",
		"-- Overview --",
		"-- HTTP Headers --",
		"-- User-Defined Metadata --",
	}

	for _, section := range expectedSections {
		if !strings.Contains(result, section) {
			t.Errorf("expected output to contain section %q, got:\n%s", section, result)
		}
	}

	expectedValues := []string{
		"application/pdf",
		"AES256",
		"projects/p/locations/l/keyRings/kr/cryptoKeys/k",
		"CRC32C",
		"AABBCC==",
		"Generation",
		"12345",
		"author", "test-user",
	}

	for _, v := range expectedValues {
		if !strings.Contains(result, v) {
			t.Errorf("expected output to contain %q, got:\n%s", v, result)
		}
	}
}
