package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"testing"
)

type mockClient struct {
	name   string
	closed bool
}

func (m *mockClient) Close() error {
	m.closed = true
	return nil
}

func TestConcurrentFanOut_AllSucceed(t *testing.T) {
	results, err := concurrentFanOut(
		context.Background(),
		[]string{"a", "b"},
		func(ctx context.Context, name string) (*mockClient, error) {
			return &mockClient{name: name}, nil
		},
		func(ctx context.Context, client *mockClient) ([]string, error) {
			return []string{client.name + "-1", client.name + "-2"}, nil
		},
		slog.Default(),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 4 {
		t.Fatalf("expected 4 results, got %d", len(results))
	}
}

func TestConcurrentFanOut_AllFail(t *testing.T) {
	results, err := concurrentFanOut(
		context.Background(),
		[]string{"a", "b"},
		func(ctx context.Context, name string) (*mockClient, error) {
			return nil, fmt.Errorf("init failed for %s", name)
		},
		func(ctx context.Context, client *mockClient) ([]string, error) {
			t.Fatal("listFn should not be called when getClient fails")
			return nil, nil
		},
		slog.Default(),
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if len(results) != 0 {
		t.Errorf("expected no results, got %d", len(results))
	}
	if !strings.Contains(err.Error(), "provider a") || !strings.Contains(err.Error(), "provider b") {
		t.Errorf("expected errors for both providers, got: %v", err)
	}
}

func TestConcurrentFanOut_PartialFailure(t *testing.T) {
	results, err := concurrentFanOut(
		context.Background(),
		[]string{"ok", "fail"},
		func(ctx context.Context, name string) (*mockClient, error) {
			return &mockClient{name: name}, nil
		},
		func(ctx context.Context, client *mockClient) ([]string, error) {
			if client.name == "fail" {
				return nil, fmt.Errorf("list failed")
			}
			return []string{"result-1"}, nil
		},
		slog.Default(),
	)

	if err == nil {
		t.Fatal("expected error for partial failure")
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result from successful provider, got %d", len(results))
	}
	if !strings.Contains(err.Error(), "provider fail") {
		t.Errorf("expected error to mention failing provider, got: %v", err)
	}
}

func TestConcurrentFanOut_EmptyProviders(t *testing.T) {
	results, err := concurrentFanOut(
		context.Background(),
		[]string{},
		func(ctx context.Context, name string) (*mockClient, error) {
			t.Fatal("getClient should not be called for empty providers")
			return nil, nil
		},
		func(ctx context.Context, client *mockClient) ([]string, error) {
			t.Fatal("listFn should not be called for empty providers")
			return nil, nil
		},
		slog.Default(),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected no results, got %d", len(results))
	}
}

func TestConcurrentFanOut_SingleProvider(t *testing.T) {
	results, err := concurrentFanOut(
		context.Background(),
		[]string{"only"},
		func(ctx context.Context, name string) (*mockClient, error) {
			return &mockClient{name: name}, nil
		},
		func(ctx context.Context, client *mockClient) ([]string, error) {
			return []string{"result-1"}, nil
		},
		slog.Default(),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestConcurrentFanOut_ListFnReturnsEmptySlice(t *testing.T) {
	results, err := concurrentFanOut(
		context.Background(),
		[]string{"empty-provider"},
		func(ctx context.Context, name string) (*mockClient, error) {
			return &mockClient{name: name}, nil
		},
		func(ctx context.Context, client *mockClient) ([]string, error) {
			return []string{}, nil
		},
		slog.Default(),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}
