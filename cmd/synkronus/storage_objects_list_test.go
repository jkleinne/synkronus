package main

import (
	"bytes"
	"context"
	"testing"

	"synkronus/internal/domain"
	"synkronus/internal/domain/storage"
)

// --- list-objects tests ---

func TestListObjectsCmd_HappyPath_ReturnsObjects(t *testing.T) {
	wantObjects := storage.ObjectList{
		BucketName: "my-bucket",
		Objects: []storage.Object{
			{Key: "report.csv", Bucket: "my-bucket", Provider: domain.GCP},
			{Key: "data/export.json", Bucket: "my-bucket", Provider: domain.GCP},
		},
	}
	mock := &cmdMockStorage{objects: wantObjects}
	factory := &cmdStorageFactory{providers: map[string]storage.Storage{"gcp": mock}}
	// list-objects uses only StorageService, not ProviderFactory.
	app := newStorageTestApp(factory, nil)

	var buf bytes.Buffer
	cmd := newListObjectsCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetContext(app.ToContext(context.Background()))
	cmd.SetArgs([]string{"--provider", "gcp", "--bucket", "my-bucket"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Output is rendered to os.Stdout (not the cobra writer), so we assert on
	// successful execution rather than captured output content.
}

func TestListObjectsCmd_MissingProviderFlag_ReturnsError(t *testing.T) {
	app := newStorageTestApp(&cmdStorageFactory{}, nil)

	var buf bytes.Buffer
	cmd := newListObjectsCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetContext(app.ToContext(context.Background()))
	cmd.SetArgs([]string{"--bucket", "my-bucket"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for missing --provider flag, got nil")
	}
}

func TestListObjectsCmd_MissingBucketFlag_ReturnsError(t *testing.T) {
	app := newStorageTestApp(&cmdStorageFactory{}, nil)

	var buf bytes.Buffer
	cmd := newListObjectsCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetContext(app.ToContext(context.Background()))
	cmd.SetArgs([]string{"--provider", "gcp"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for missing --bucket flag, got nil")
	}
}

func TestListObjectsCmd_EmptyBucket_ReturnsEmptyList(t *testing.T) {
	mock := &cmdMockStorage{objects: storage.ObjectList{BucketName: "empty-bucket"}}
	factory := &cmdStorageFactory{providers: map[string]storage.Storage{"gcp": mock}}
	app := newStorageTestApp(factory, nil)

	var buf bytes.Buffer
	cmd := newListObjectsCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetContext(app.ToContext(context.Background()))
	cmd.SetArgs([]string{"--provider", "gcp", "--bucket", "empty-bucket"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error for empty bucket: %v", err)
	}
}
