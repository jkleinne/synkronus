package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"testing"

	"synkronus/internal/domain"
	"synkronus/internal/domain/storage"
)

// --- readerWithCleanup tests ---

func TestReaderWithCleanup_CloseBothSucceed(t *testing.T) {
	inner := io.NopCloser(strings.NewReader("data"))
	r := &readerWithCleanup{ReadCloser: inner, cleanup: func() error { return nil }}

	if err := r.Close(); err != nil {
		t.Errorf("expected nil error, got: %v", err)
	}
}

func TestReaderWithCleanup_ReaderCloseError(t *testing.T) {
	r := &readerWithCleanup{
		ReadCloser: &errorCloser{readErr: errors.New("reader fail")},
		cleanup:    func() error { return nil },
	}

	err := r.Close()
	if err == nil || !strings.Contains(err.Error(), "reader fail") {
		t.Errorf("expected 'reader fail' error, got: %v", err)
	}
}

func TestReaderWithCleanup_CleanupError(t *testing.T) {
	inner := io.NopCloser(strings.NewReader("data"))
	r := &readerWithCleanup{
		ReadCloser: inner,
		cleanup:    func() error { return errors.New("cleanup fail") },
	}

	err := r.Close()
	if err == nil || !strings.Contains(err.Error(), "cleanup fail") {
		t.Errorf("expected 'cleanup fail' error, got: %v", err)
	}
}

func TestReaderWithCleanup_BothErrors(t *testing.T) {
	r := &readerWithCleanup{
		ReadCloser: &errorCloser{readErr: errors.New("reader fail")},
		cleanup:    func() error { return errors.New("cleanup fail") },
	}

	err := r.Close()
	if err == nil {
		t.Fatal("expected combined error, got nil")
	}
	if !strings.Contains(err.Error(), "reader fail") {
		t.Errorf("error should contain 'reader fail', got: %v", err)
	}
	if !strings.Contains(err.Error(), "cleanup fail") {
		t.Errorf("error should contain 'cleanup fail', got: %v", err)
	}
}

// errorCloser is an io.ReadCloser whose Close always returns an error.
type errorCloser struct {
	readErr error
}

func (e *errorCloser) Read(p []byte) (int, error) { return 0, io.EOF }
func (e *errorCloser) Close() error               { return e.readErr }

// --- Mock types ---

// mockStorageFactory satisfies StorageProviderFactory. If err is set it is
// returned for every GetStorageProvider call; otherwise the call is dispatched
// to the providers map by name.
type mockStorageFactory struct {
	providers map[string]storage.Storage
	err       error
}

func (m *mockStorageFactory) GetStorageProvider(_ context.Context, name string) (storage.Storage, error) {
	if m.err != nil {
		return nil, m.err
	}
	client, ok := m.providers[name]
	if !ok {
		return nil, fmt.Errorf("unsupported provider: %s", name)
	}
	return client, nil
}

// mockStorage is a controllable implementation of storage.Storage. All methods
// return the configured result fields; errors are returned when err is non-nil.
type mockStorage struct {
	providerName domain.Provider
	buckets      []storage.Bucket
	bucket       storage.Bucket
	objects      storage.ObjectList
	object       storage.Object
	createResult storage.CreateBucketResult
	reader       io.ReadCloser
	err          error
	closeCalled  bool
}

func (m *mockStorage) ListBuckets(_ context.Context) ([]storage.Bucket, error) {
	return m.buckets, m.err
}

func (m *mockStorage) DescribeBucket(_ context.Context, _ string) (storage.Bucket, error) {
	return m.bucket, m.err
}

func (m *mockStorage) CreateBucket(_ context.Context, _ storage.CreateBucketOptions) (storage.CreateBucketResult, error) {
	return m.createResult, m.err
}

func (m *mockStorage) DeleteBucket(_ context.Context, _ string) error {
	return m.err
}

func (m *mockStorage) ListObjects(_ context.Context, _ string, _ string) (storage.ObjectList, error) {
	return m.objects, m.err
}

func (m *mockStorage) DescribeObject(_ context.Context, _ string, _ string) (storage.Object, error) {
	return m.object, m.err
}

func (m *mockStorage) DownloadObject(_ context.Context, _ string, _ string) (io.ReadCloser, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.reader != nil {
		return m.reader, nil
	}
	return io.NopCloser(strings.NewReader("object-data")), nil
}

func (m *mockStorage) UploadObject(_ context.Context, _ storage.UploadObjectOptions, _ io.Reader) error {
	return m.err
}

func (m *mockStorage) DeleteObject(_ context.Context, _ string, _ string) error {
	return m.err
}

func (m *mockStorage) CopyObject(_ context.Context, _, _, _, _ string) error {
	return m.err
}

func (m *mockStorage) ProviderName() domain.Provider {
	return m.providerName
}

func (m *mockStorage) Close() error {
	m.closeCalled = true
	return nil
}

// newTestLogger returns a logger that discards all output, suitable for tests.
func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// newStorageService builds a StorageService wired to the supplied factory.
func newStorageService(factory StorageProviderFactory) *StorageService {
	return NewStorageService(factory, newTestLogger())
}

// --- StorageService tests ---

func TestStorageService_DescribeBucket_HappyPath(t *testing.T) {
	want := storage.Bucket{Name: "my-bucket", Provider: domain.GCP}
	mock := &mockStorage{bucket: want}
	svc := newStorageService(&mockStorageFactory{providers: map[string]storage.Storage{"gcp": mock}})

	got, err := svc.DescribeBucket(context.Background(), "my-bucket", "gcp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != want.Name {
		t.Errorf("got bucket name %q, want %q", got.Name, want.Name)
	}
}

func TestStorageService_DescribeBucket_ProviderNotFound(t *testing.T) {
	svc := newStorageService(&mockStorageFactory{providers: map[string]storage.Storage{}})

	_, err := svc.DescribeBucket(context.Background(), "my-bucket", "unknown")
	if err == nil {
		t.Fatal("expected error for unknown provider, got nil")
	}
	if !strings.Contains(err.Error(), "unknown") {
		t.Errorf("error should mention provider name, got: %v", err)
	}
}

func TestStorageService_DescribeBucket_ProviderError(t *testing.T) {
	providerErr := errors.New("describe failed")
	mock := &mockStorage{err: providerErr}
	svc := newStorageService(&mockStorageFactory{providers: map[string]storage.Storage{"gcp": mock}})

	_, err := svc.DescribeBucket(context.Background(), "my-bucket", "gcp")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, providerErr) {
		t.Errorf("error chain should wrap provider error; got: %v", err)
	}
}

func TestStorageService_DeleteBucket_HappyPath(t *testing.T) {
	mock := &mockStorage{}
	svc := newStorageService(&mockStorageFactory{providers: map[string]storage.Storage{"gcp": mock}})

	if err := svc.DeleteBucket(context.Background(), "my-bucket", "gcp"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !mock.closeCalled {
		t.Error("expected Close to be called on the provider client")
	}
}

func TestStorageService_CreateBucket_HappyPath(t *testing.T) {
	want := storage.CreateBucketResult{Warnings: []string{"uniform bucket-level access not set"}}
	mock := &mockStorage{createResult: want}
	svc := newStorageService(&mockStorageFactory{providers: map[string]storage.Storage{"gcp": mock}})

	opts := storage.CreateBucketOptions{Name: "new-bucket", Location: "us-east1"}
	got, err := svc.CreateBucket(context.Background(), opts, "gcp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Warnings) != len(want.Warnings) {
		t.Errorf("got %d warnings, want %d", len(got.Warnings), len(want.Warnings))
	}
}

func TestStorageService_CreateBucket_Error(t *testing.T) {
	createErr := errors.New("bucket already exists")
	mock := &mockStorage{err: createErr}
	svc := newStorageService(&mockStorageFactory{providers: map[string]storage.Storage{"gcp": mock}})

	_, err := svc.CreateBucket(context.Background(), storage.CreateBucketOptions{Name: "dup-bucket"}, "gcp")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, createErr) {
		t.Errorf("error chain should wrap provider error; got: %v", err)
	}
}

func TestStorageService_ListObjects_HappyPath(t *testing.T) {
	want := storage.ObjectList{
		BucketName: "my-bucket",
		Objects:    []storage.Object{{Key: "file.txt"}},
	}
	mock := &mockStorage{objects: want}
	svc := newStorageService(&mockStorageFactory{providers: map[string]storage.Storage{"gcp": mock}})

	got, err := svc.ListObjects(context.Background(), "my-bucket", "gcp", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Objects) != len(want.Objects) {
		t.Errorf("got %d objects, want %d", len(got.Objects), len(want.Objects))
	}
}

func TestStorageService_DescribeObject_HappyPath(t *testing.T) {
	want := storage.Object{Key: "file.txt", Bucket: "my-bucket"}
	mock := &mockStorage{object: want}
	svc := newStorageService(&mockStorageFactory{providers: map[string]storage.Storage{"gcp": mock}})

	got, err := svc.DescribeObject(context.Background(), "my-bucket", "file.txt", "gcp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Key != want.Key {
		t.Errorf("got key %q, want %q", got.Key, want.Key)
	}
}

func TestStorageService_DownloadObject_HappyPath(t *testing.T) {
	content := "hello, world"
	mock := &mockStorage{
		reader: io.NopCloser(strings.NewReader(content)),
	}
	svc := newStorageService(&mockStorageFactory{providers: map[string]storage.Storage{"gcp": mock}})

	rc, err := svc.DownloadObject(context.Background(), "my-bucket", "file.txt", "gcp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, rc); err != nil {
		t.Fatalf("reading from returned reader: %v", err)
	}
	if buf.String() != content {
		t.Errorf("got content %q, want %q", buf.String(), content)
	}

	// Close triggers cleanup (client.Close)
	if err := rc.Close(); err != nil {
		t.Errorf("unexpected close error: %v", err)
	}
	if !mock.closeCalled {
		t.Error("expected provider client Close to be called when reader is closed")
	}
}

func TestStorageService_UploadObject_HappyPath(t *testing.T) {
	mock := &mockStorage{}
	svc := newStorageService(&mockStorageFactory{providers: map[string]storage.Storage{"gcp": mock}})

	opts := storage.UploadObjectOptions{BucketName: "my-bucket", ObjectKey: "file.txt"}
	if err := svc.UploadObject(context.Background(), opts, "gcp", strings.NewReader("data")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStorageService_DeleteObject_HappyPath(t *testing.T) {
	mock := &mockStorage{}
	svc := newStorageService(&mockStorageFactory{providers: map[string]storage.Storage{"gcp": mock}})

	if err := svc.DeleteObject(context.Background(), "my-bucket", "file.txt", "gcp"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStorageService_CopyObject_HappyPath(t *testing.T) {
	mock := &mockStorage{}
	svc := newStorageService(&mockStorageFactory{providers: map[string]storage.Storage{"gcp": mock}})

	if err := svc.CopyObject(context.Background(), "src-bucket", "src-key", "dst-bucket", "dst-key", "gcp"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStorageService_ListAllBuckets_PartialFailure(t *testing.T) {
	successBuckets := []storage.Bucket{
		{Name: "bucket-a", Provider: domain.GCP},
	}
	failErr := errors.New("aws not reachable")

	goodMock := &mockStorage{buckets: successBuckets}
	badMock := &mockStorage{err: failErr}

	svc := newStorageService(&mockStorageFactory{
		providers: map[string]storage.Storage{
			"gcp": goodMock,
			"aws": badMock,
		},
	})

	results, err := svc.ListAllBuckets(context.Background(), []string{"gcp", "aws"})
	if err == nil {
		t.Fatal("expected a partial error, got nil")
	}
	if !errors.Is(err, failErr) {
		t.Errorf("error chain should wrap provider error; got: %v", err)
	}
	// Successful provider results must still be present
	if len(results) != len(successBuckets) {
		t.Errorf("got %d results, want %d (partial success should be preserved)", len(results), len(successBuckets))
	}
}
