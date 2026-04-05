package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"synkronus/internal/domain"
	"synkronus/internal/domain/storage"
	"synkronus/internal/output"
	"synkronus/internal/service"
)

// --- CLI-layer mock types ---
// These types are defined here and shared across the cmd/synkronus test files.
// They provide controllable implementations of provider interfaces without
// duplicating any service-layer logic.

// cmdStorageFactory is a test double for service.StorageProviderFactory.
type cmdStorageFactory struct {
	providers map[string]storage.Storage
	err       error
}

func (f *cmdStorageFactory) GetStorageProvider(_ context.Context, name string) (storage.Storage, error) {
	if f.err != nil {
		return nil, f.err
	}
	client, ok := f.providers[name]
	if !ok {
		return nil, errors.New("unsupported provider: " + name)
	}
	return client, nil
}

// cmdMockStorage is a controllable implementation of storage.Storage for CLI tests.
type cmdMockStorage struct {
	buckets      []storage.Bucket
	bucket       storage.Bucket
	objects      storage.ObjectList
	object       storage.Object
	createResult storage.CreateBucketResult
	err          error
	closeCalled  bool
}

func (m *cmdMockStorage) ListBuckets(_ context.Context) ([]storage.Bucket, error) {
	return m.buckets, m.err
}
func (m *cmdMockStorage) DescribeBucket(_ context.Context, _ string) (storage.Bucket, error) {
	return m.bucket, m.err
}
func (m *cmdMockStorage) CreateBucket(_ context.Context, _ storage.CreateBucketOptions) (storage.CreateBucketResult, error) {
	return m.createResult, m.err
}
func (m *cmdMockStorage) DeleteBucket(_ context.Context, _ string) error {
	return m.err
}
func (m *cmdMockStorage) ListObjects(_ context.Context, _ string, _ string) (storage.ObjectList, error) {
	return m.objects, m.err
}
func (m *cmdMockStorage) DescribeObject(_ context.Context, _ string, _ string) (storage.Object, error) {
	return m.object, m.err
}
func (m *cmdMockStorage) DownloadObject(_ context.Context, _ string, _ string) (io.ReadCloser, error) {
	return nil, m.err
}
func (m *cmdMockStorage) UploadObject(_ context.Context, _ storage.UploadObjectOptions, _ io.Reader) error {
	return m.err
}
func (m *cmdMockStorage) DeleteObject(_ context.Context, _ string, _ string) error {
	return m.err
}
func (m *cmdMockStorage) CopyObject(_ context.Context, _, _, _, _ string) error {
	return m.err
}
func (m *cmdMockStorage) ProviderName() domain.Provider { return domain.GCP }
func (m *cmdMockStorage) Close() error {
	m.closeCalled = true
	return nil
}

// cmdTestLogger returns a logger that discards all output, suitable for CLI tests.
func cmdTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// newStorageTestApp builds a minimal appContainer wired to the supplied storage
// factory and prompter. ProviderFactory is nil because single-resource commands
// (delete, describe) do not enumerate providers.
func newStorageTestApp(factory service.StorageProviderFactory, prompter *mockPrompter) *appContainer {
	logger := cmdTestLogger()
	return &appContainer{
		StorageService: service.NewStorageService(factory, logger),
		Prompter:       prompter,
		OutputFormat:   output.FormatTable,
		Logger:         logger,
	}
}

// --- delete-bucket tests ---

func TestDeleteBucketCmd_ForceFlag_SkipsConfirmationAndSucceeds(t *testing.T) {
	mock := &cmdMockStorage{}
	factory := &cmdStorageFactory{providers: map[string]storage.Storage{"gcp": mock}}
	app := newStorageTestApp(factory, nil)

	cmd := newDeleteBucketCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetContext(app.ToContext(context.Background()))
	cmd.SetArgs([]string{"--provider", "gcp", "--force", "my-bucket"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !mock.closeCalled {
		t.Error("expected provider client Close to be called after delete")
	}
}

func TestDeleteBucketCmd_ServiceError_ReturnsError(t *testing.T) {
	serviceErr := errors.New("bucket not empty")
	mock := &cmdMockStorage{err: serviceErr}
	factory := &cmdStorageFactory{providers: map[string]storage.Storage{"gcp": mock}}
	app := newStorageTestApp(factory, nil)

	var buf bytes.Buffer
	cmd := newDeleteBucketCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetContext(app.ToContext(context.Background()))
	cmd.SetArgs([]string{"--provider", "gcp", "--force", "my-bucket"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, serviceErr) {
		t.Errorf("expected service error in chain, got: %v", err)
	}
}

func TestDeleteBucketCmd_MissingProviderFlag_ReturnsError(t *testing.T) {
	app := newStorageTestApp(&cmdStorageFactory{}, nil)

	var buf bytes.Buffer
	cmd := newDeleteBucketCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetContext(app.ToContext(context.Background()))
	cmd.SetArgs([]string{"--force", "my-bucket"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for missing --provider flag, got nil")
	}
}

func TestDeleteBucketCmd_MissingBucketArg_ReturnsError(t *testing.T) {
	app := newStorageTestApp(&cmdStorageFactory{}, nil)

	var buf bytes.Buffer
	cmd := newDeleteBucketCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetContext(app.ToContext(context.Background()))
	cmd.SetArgs([]string{"--provider", "gcp", "--force"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for missing bucket argument, got nil")
	}
}

func TestDeleteBucketCmd_ConfirmDeclined_AbortsWithoutDelete(t *testing.T) {
	mock := &cmdMockStorage{}
	factory := &cmdStorageFactory{providers: map[string]storage.Storage{"gcp": mock}}
	prompter := &mockPrompter{confirmed: false}
	app := newStorageTestApp(factory, prompter)

	var buf bytes.Buffer
	cmd := newDeleteBucketCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetContext(app.ToContext(context.Background()))
	cmd.SetArgs([]string{"--provider", "gcp", "my-bucket"})

	err := cmd.Execute()
	if !errors.Is(err, ErrOperationAborted) {
		t.Errorf("expected ErrOperationAborted, got: %v", err)
	}
	if mock.closeCalled {
		t.Error("provider client should not be called when confirmation is declined")
	}
}
