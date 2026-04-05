package service

import (
	"context"
	"errors"
	"testing"

	"synkronus/internal/domain/storage"
)

// TestWithClientResult_HappyPath verifies that on success the fn return value
// is propagated and the client is closed afterward.
func TestWithClientResult_HappyPath(t *testing.T) {
	want := storage.Bucket{Name: "test-bucket"}
	mock := &mockStorage{bucket: want}

	getClient := func(_ context.Context, _ string) (storage.Storage, error) {
		return mock, nil
	}

	got, err := withClientResult(context.Background(), getClient, "gcp", func(client storage.Storage) (storage.Bucket, error) {
		return client.DescribeBucket(context.Background(), "test-bucket")
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != want.Name {
		t.Errorf("got bucket name %q, want %q", got.Name, want.Name)
	}
	if !mock.closeCalled {
		t.Error("expected Close to be called on success")
	}
}

// TestWithClientResult_ClientError verifies that when getClient fails fn is
// never called and the zero value with the error is returned.
func TestWithClientResult_ClientError(t *testing.T) {
	clientErr := errors.New("provider unavailable")
	fnCalled := false

	getClient := func(_ context.Context, _ string) (storage.Storage, error) {
		return nil, clientErr
	}

	_, err := withClientResult(context.Background(), getClient, "gcp", func(_ storage.Storage) (storage.Bucket, error) {
		fnCalled = true
		return storage.Bucket{}, nil
	})

	if err == nil {
		t.Fatal("expected error from client acquisition, got nil")
	}
	if !errors.Is(err, clientErr) {
		t.Errorf("error chain should wrap client error; got: %v", err)
	}
	if fnCalled {
		t.Error("fn should not be called when getClient fails")
	}
}

// TestWithClientResult_FnError verifies that when fn returns an error the
// client is still closed and the error is propagated.
func TestWithClientResult_FnError(t *testing.T) {
	fnErr := errors.New("operation failed")
	mock := &mockStorage{}

	getClient := func(_ context.Context, _ string) (storage.Storage, error) {
		return mock, nil
	}

	_, err := withClientResult(context.Background(), getClient, "gcp", func(_ storage.Storage) (storage.Bucket, error) {
		return storage.Bucket{}, fnErr
	})

	if err == nil {
		t.Fatal("expected error from fn, got nil")
	}
	if !errors.Is(err, fnErr) {
		t.Errorf("error chain should wrap fn error; got: %v", err)
	}
	if !mock.closeCalled {
		t.Error("expected Close to be called even when fn returns an error")
	}
}
