package main

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"synkronus/internal/domain/storage"
)

func TestCreateBucketCmd_HappyPath(t *testing.T) {
	mock := &cmdMockStorage{
		createResult: storage.CreateBucketResult{},
	}
	factory := &cmdStorageFactory{providers: map[string]storage.Storage{"gcp": mock}}
	app := newStorageTestApp(factory, nil)

	var buf bytes.Buffer
	cmd := newCreateBucketCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetContext(app.ToContext(context.Background()))
	cmd.SetArgs([]string{"my-bucket", "--provider", "gcp", "--location", "us-central1"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateBucketCmd_MissingProviderFlag_ReturnsError(t *testing.T) {
	app := newStorageTestApp(&cmdStorageFactory{}, nil)

	var buf bytes.Buffer
	cmd := newCreateBucketCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetContext(app.ToContext(context.Background()))
	cmd.SetArgs([]string{"my-bucket", "--location", "us-central1"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for missing --provider flag, got nil")
	}
}

func TestCreateBucketCmd_MissingLocationFlag_ReturnsError(t *testing.T) {
	app := newStorageTestApp(&cmdStorageFactory{}, nil)

	var buf bytes.Buffer
	cmd := newCreateBucketCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetContext(app.ToContext(context.Background()))
	cmd.SetArgs([]string{"my-bucket", "--provider", "gcp"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for missing --location flag, got nil")
	}
}

func TestCreateBucketCmd_UniformAccessNonGCP_ReturnsError(t *testing.T) {
	mock := &cmdMockStorage{}
	factory := &cmdStorageFactory{providers: map[string]storage.Storage{"aws": mock}}
	app := newStorageTestApp(factory, nil)

	var buf bytes.Buffer
	cmd := newCreateBucketCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetContext(app.ToContext(context.Background()))
	cmd.SetArgs([]string{"my-bucket", "--provider", "aws", "--location", "us-east-1", "--uniform-access"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for --uniform-access on non-GCP provider, got nil")
	}
}

func TestCreateBucketCmd_ServiceError_ReturnsError(t *testing.T) {
	serviceErr := errors.New("bucket already exists")
	mock := &cmdMockStorage{err: serviceErr}
	factory := &cmdStorageFactory{providers: map[string]storage.Storage{"gcp": mock}}
	app := newStorageTestApp(factory, nil)

	var buf bytes.Buffer
	cmd := newCreateBucketCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetContext(app.ToContext(context.Background()))
	cmd.SetArgs([]string{"my-bucket", "--provider", "gcp", "--location", "us-central1"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error from service, got nil")
	}
	if !errors.Is(err, serviceErr) {
		t.Errorf("expected service error in chain, got: %v", err)
	}
}

func TestCreateBucketCmd_InvalidPublicAccessPrevention_ReturnsError(t *testing.T) {
	app := newStorageTestApp(&cmdStorageFactory{}, nil)

	var buf bytes.Buffer
	cmd := newCreateBucketCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetContext(app.ToContext(context.Background()))
	cmd.SetArgs([]string{"my-bucket", "--provider", "gcp", "--location", "us-central1", "--public-access-prevention", "blocked"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for invalid --public-access-prevention value, got nil")
	}
}
