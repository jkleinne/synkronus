package aws

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"testing"

	smithy "github.com/aws/smithy-go"
)

// logRecord captures a single slog emission for test assertions.
type logRecord struct {
	Level   slog.Level
	Message string
}

// capturingHandler is a minimal slog.Handler that records log emissions.
type capturingHandler struct {
	mu      sync.Mutex
	records []logRecord
}

func (h *capturingHandler) Enabled(context.Context, slog.Level) bool { return true }
func (h *capturingHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.records = append(h.records, logRecord{Level: r.Level, Message: r.Message})
	return nil
}
func (h *capturingHandler) WithAttrs([]slog.Attr) slog.Handler { return h }
func (h *capturingHandler) WithGroup(string) slog.Handler       { return h }

func (h *capturingHandler) count() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.records)
}

func TestBucketFetcherRun_Success(t *testing.T) {
	handler := &capturingHandler{}
	logger := slog.New(handler)

	f := bucketFetcher{
		label: "test-field",
		fetch: func(ctx context.Context) error { return nil },
	}

	err := f.run(context.Background(), logger, "my-bucket")()

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if handler.count() != 0 {
		t.Errorf("expected no log records, got %d", handler.count())
	}
}

func TestBucketFetcherRun_NotConfiguredError_Tolerated(t *testing.T) {
	handler := &capturingHandler{}
	logger := slog.New(handler)

	f := bucketFetcher{
		label:                 "tags",
		tolerateNotConfigured: true,
		fetch: func(ctx context.Context) error {
			return &smithy.GenericAPIError{Code: "NoSuchTagSet", Message: "not configured"}
		},
	}

	err := f.run(context.Background(), logger, "my-bucket")()

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if handler.count() != 0 {
		t.Errorf("expected no log records for tolerated not-configured error, got %d", handler.count())
	}
}

func TestBucketFetcherRun_NotConfiguredError_NotTolerated(t *testing.T) {
	handler := &capturingHandler{}
	logger := slog.New(handler)

	f := bucketFetcher{
		label:                 "tags",
		tolerateNotConfigured: false,
		fetch: func(ctx context.Context) error {
			return &smithy.GenericAPIError{Code: "NoSuchTagSet", Message: "not configured"}
		},
	}

	err := f.run(context.Background(), logger, "my-bucket")()

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if handler.count() != 1 {
		t.Fatalf("expected 1 warning log record, got %d", handler.count())
	}
	if handler.records[0].Level != slog.LevelWarn {
		t.Errorf("expected Warn level, got %v", handler.records[0].Level)
	}
}

func TestBucketFetcherRun_OtherError_Tolerated(t *testing.T) {
	handler := &capturingHandler{}
	logger := slog.New(handler)

	f := bucketFetcher{
		label:                 "encryption",
		tolerateNotConfigured: true,
		fetch: func(ctx context.Context) error {
			return errors.New("access denied")
		},
	}

	err := f.run(context.Background(), logger, "my-bucket")()

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if handler.count() != 1 {
		t.Fatalf("expected 1 warning log record, got %d", handler.count())
	}
	if handler.records[0].Level != slog.LevelWarn {
		t.Errorf("expected Warn level, got %v", handler.records[0].Level)
	}
}

func TestBucketFetcherRun_OtherError_NotTolerated(t *testing.T) {
	handler := &capturingHandler{}
	logger := slog.New(handler)

	f := bucketFetcher{
		label:                 "location",
		tolerateNotConfigured: false,
		fetch: func(ctx context.Context) error {
			return errors.New("network timeout")
		},
	}

	err := f.run(context.Background(), logger, "my-bucket")()

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if handler.count() != 1 {
		t.Fatalf("expected 1 warning log record, got %d", handler.count())
	}
	if handler.records[0].Level != slog.LevelWarn {
		t.Errorf("expected Warn level, got %v", handler.records[0].Level)
	}
}
