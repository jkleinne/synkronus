package main

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"synkronus/internal/config"
	"synkronus/internal/domain"
	"synkronus/internal/domain/storage"
	"synkronus/internal/output"
	"synkronus/internal/provider/factory"
	"synkronus/internal/service"
)

// newBucketListTestApp builds an appContainer suitable for list-buckets tests.
// It wires a real factory (for provider resolution) with a mock StorageService
// (for the actual list operation). The config marks GCP as configured so that
// the provider resolver returns ["gcp"] when no --providers flag is given.
func newBucketListTestApp(storageFactory service.StorageProviderFactory) *appContainer {
	logger := cmdTestLogger()
	gcpCfg := &config.Config{GCP: &config.GCPConfig{Project: "test-project"}}
	providerFactory := factory.NewFactory(gcpCfg, logger)
	return &appContainer{
		ProviderFactory: providerFactory,
		StorageService:  service.NewStorageService(storageFactory, logger),
		OutputFormat:    output.FormatJSON, // JSON avoids table-rendering dependencies
		Logger:          logger,
	}
}

// --- list-buckets tests ---

func TestListBucketsCmd_HappyPath_ReturnsBuckets(t *testing.T) {
	wantBuckets := []storage.Bucket{
		{Name: "alpha", Provider: domain.GCP},
		{Name: "beta", Provider: domain.GCP},
	}
	mock := &cmdMockStorage{buckets: wantBuckets}
	factory := &cmdStorageFactory{providers: map[string]storage.Storage{"gcp": mock}}
	app := newBucketListTestApp(factory)

	var buf bytes.Buffer
	cmd := newListBucketsCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetContext(app.ToContext(context.Background()))
	cmd.SetArgs([]string{})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Output is rendered to os.Stdout (not the cobra writer), so we assert on
	// successful execution rather than captured output content.
}

func TestListBucketsCmd_EmptyResults_PrintsMessage(t *testing.T) {
	mock := &cmdMockStorage{buckets: []storage.Bucket{}}
	factory := &cmdStorageFactory{providers: map[string]storage.Storage{"gcp": mock}}
	app := newBucketListTestApp(factory)

	var buf bytes.Buffer
	cmd := newListBucketsCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetContext(app.ToContext(context.Background()))
	cmd.SetArgs([]string{})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The command prints "No buckets found." to stdout when no buckets exist.
	// Note: fmt.Println writes to os.Stdout, not the cobra output writer, so
	// we only assert no error rather than capturing that specific message.
}

func TestListBucketsCmd_InvalidProvider_ReturnsError(t *testing.T) {
	app := newBucketListTestApp(&cmdStorageFactory{})

	var buf bytes.Buffer
	cmd := newListBucketsCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetContext(app.ToContext(context.Background()))
	cmd.SetArgs([]string{"--providers", "azure"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for unsupported provider 'azure', got nil")
	}
	if !strings.Contains(err.Error(), "unsupported") {
		t.Errorf("expected 'unsupported' in error message, got: %v", err)
	}
}

func TestListBucketsCmd_ServiceError_AllProvidersFail_ReturnsError(t *testing.T) {
	serviceErr := errors.New("network timeout")
	mock := &cmdMockStorage{err: serviceErr}
	factory := &cmdStorageFactory{providers: map[string]storage.Storage{"gcp": mock}}
	app := newBucketListTestApp(factory)

	var buf bytes.Buffer
	cmd := newListBucketsCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetContext(app.ToContext(context.Background()))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when all providers fail, got nil")
	}
}
