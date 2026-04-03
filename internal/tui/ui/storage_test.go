// File: internal/tui/ui/storage_test.go
package ui

import (
	"strings"
	"testing"
	"time"

	"synkronus/internal/domain/storage"
)

func TestRenderBucketListWithData(t *testing.T) {
	buckets := []storage.Bucket{
		{Name: "test-bucket", Provider: "GCP", Location: "us-central1", UsageBytes: 1024 * 1024, StorageClass: "STANDARD", CreatedAt: time.Now()},
	}
	result := RenderBucketList(buckets, 0, 0, 80)
	if !strings.Contains(result, "test-bucket") {
		t.Error("bucket list should contain bucket name")
	}
	if !strings.Contains(result, "GCP") {
		t.Error("bucket list should contain provider")
	}
}

func TestRenderBucketListEmpty(t *testing.T) {
	result := RenderBucketList(nil, 0, 0, 80)
	if !strings.Contains(result, "No buckets") {
		t.Error("empty bucket list should show 'No buckets' message")
	}
}

func TestRenderBucketDetailSections(t *testing.T) {
	bucket := storage.Bucket{
		Name:         "detail-bucket",
		Provider:     "GCP",
		Location:     "us-east1",
		StorageClass: "NEARLINE",
		UsageBytes:   5 * 1024 * 1024 * 1024,
	}
	result := RenderBucketDetail(bucket, 80)
	if !strings.Contains(result, "Overview") {
		t.Error("detail should contain Overview section")
	}
	if !strings.Contains(result, "detail-bucket") || !strings.Contains(result, "GCP") {
		t.Error("detail should contain bucket name and provider")
	}
}

func TestRenderBucketDetailAccessControlSection(t *testing.T) {
	bucket := storage.Bucket{
		Name:                     "access-bucket",
		Provider:                 "GCP",
		PublicAccessPrevention:   "enforced",
		UniformBucketLevelAccess: &storage.UniformBucketLevelAccess{Enabled: true},
	}
	result := RenderBucketDetail(bucket, 80)
	if !strings.Contains(result, "Access Control") {
		t.Error("detail should contain Access Control section")
	}
	if !strings.Contains(result, "enforced") {
		t.Error("detail should contain public access prevention value")
	}
}

func TestRenderBucketDetailDataProtectionSection(t *testing.T) {
	bucket := storage.Bucket{
		Name:       "protected-bucket",
		Provider:   "GCP",
		Versioning: &storage.Versioning{Enabled: true},
		SoftDeletePolicy: &storage.SoftDeletePolicy{
			RetentionDuration: 7 * 24 * time.Hour,
		},
	}
	result := RenderBucketDetail(bucket, 80)
	if !strings.Contains(result, "Data Protection") {
		t.Error("detail should contain Data Protection section")
	}
}

func TestRenderObjectListEmpty(t *testing.T) {
	result := RenderObjectList(storage.ObjectList{}, 0, 0, 80)
	if !strings.Contains(result, "No objects") {
		t.Error("empty object list should show 'No objects' message")
	}
}

func TestRenderObjectListWithData(t *testing.T) {
	objectList := storage.ObjectList{
		BucketName: "my-bucket",
		Objects: []storage.Object{
			{Key: "file.txt", Size: 1024, StorageClass: "STANDARD", LastModified: time.Now()},
		},
		CommonPrefixes: []string{"subdir/"},
	}
	result := RenderObjectList(objectList, 0, 0, 80)
	if !strings.Contains(result, "file.txt") {
		t.Error("object list should contain object key")
	}
	if !strings.Contains(result, "subdir/") {
		t.Error("object list should contain common prefix")
	}
	if !strings.Contains(result, directoryEntry) {
		t.Error("object list should mark directories with DIR marker")
	}
}

func TestRenderObjectDetailSections(t *testing.T) {
	obj := storage.Object{
		Key:          "path/to/file.txt",
		Bucket:       "my-bucket",
		Provider:     "GCP",
		Size:         2048,
		StorageClass: "STANDARD",
		LastModified: time.Now(),
		ETag:         "abc123",
		ContentType:  "text/plain",
		CacheControl: "no-cache",
		Metadata:     map[string]string{"author": "test"},
	}
	result := RenderObjectDetail(obj, 80)
	if !strings.Contains(result, "Overview") {
		t.Error("object detail should contain Overview section")
	}
	if !strings.Contains(result, "HTTP Headers") {
		t.Error("object detail should contain HTTP Headers section when headers present")
	}
	if !strings.Contains(result, "Metadata") {
		t.Error("object detail should contain Metadata section when metadata present")
	}
	if !strings.Contains(result, "path/to/file.txt") {
		t.Error("object detail should contain object key")
	}
}

func TestRenderObjectDetailNoHTTPHeaders(t *testing.T) {
	obj := storage.Object{
		Key:          "bare.bin",
		Bucket:       "my-bucket",
		Provider:     "GCP",
		Size:         512,
		StorageClass: "COLDLINE",
		LastModified: time.Now(),
	}
	result := RenderObjectDetail(obj, 80)
	// HTTP Headers section should still be rendered (with empty entries) — just verify no panic
	if !strings.Contains(result, "Overview") {
		t.Error("object detail should always contain Overview section")
	}
}

func TestRenderCreateBucketForm(t *testing.T) {
	fields := CreateBucketFormFields{
		Name: "my-bucket", Provider: "gcp", Location: "us-central1",
	}
	result := RenderCreateBucketForm(fields, 0, "[cursor]")
	if !strings.Contains(result, "Name") {
		t.Error("create form should show Name label")
	}
	if !strings.Contains(result, "Storage Class") {
		t.Error("create form should show Storage Class label")
	}
	if !strings.Contains(result, "Public Access Prevention") {
		t.Error("create form should show Public Access Prevention label")
	}
}

func TestRenderCreateBucketFormShowsHintsForEmptyOptionalFields(t *testing.T) {
	fields := CreateBucketFormFields{
		Name: "my-bucket", Provider: "gcp", Location: "us-central1",
	}
	result := RenderCreateBucketForm(fields, 0, "[cursor]")
	if !strings.Contains(result, "STANDARD") {
		t.Error("empty storage class should show hint with valid values")
	}
	if !strings.Contains(result, "yes/no") {
		t.Error("empty versioning should show yes/no hint")
	}
}

func TestRenderCreateBucketFormHighlightsActiveField(t *testing.T) {
	// Field 2 (Location) is active — should show text input cursor.
	fields := CreateBucketFormFields{
		Name:     "my-bucket",
		Provider: "gcp",
		Location: "us-central1",
	}
	result := RenderCreateBucketForm(fields, 2, "[cursor]")
	if !strings.Contains(result, "[cursor]") {
		t.Error("create form should render textInputView next to the active field")
	}
}

func TestRenderCreateBucketFormProviderSelector(t *testing.T) {
	fields := CreateBucketFormFields{
		Name:               "my-bucket",
		Provider:           "gcp",
		ProviderIsSelector: true,
		AvailableProviders: []string{"gcp", "aws"},
		Location:           "us-central1",
	}
	// Provider field (1) active — should show selector arrows, not text cursor.
	result := RenderCreateBucketForm(fields, 1, "[cursor]")
	if !strings.Contains(result, "◀") || !strings.Contains(result, "▶") {
		t.Error("active provider field should show selector arrows")
	}
	if strings.Contains(result, "[cursor]") {
		t.Error("active provider field should not show text cursor when selector is enabled")
	}
	if !strings.Contains(result, "gcp") {
		t.Error("active provider field should show selected provider name")
	}
}

func TestRenderDeleteConfirm(t *testing.T) {
	result := RenderDeleteConfirm("my-bucket", "", "[cursor]")
	if !strings.Contains(result, "my-bucket") {
		t.Error("delete confirm should show bucket name")
	}
}

func TestRenderDeleteConfirmShowsTextInput(t *testing.T) {
	result := RenderDeleteConfirm("my-bucket", "my-buc", "[cursor]")
	if !strings.Contains(result, "[cursor]") {
		t.Error("delete confirm should render the text input view")
	}
}
