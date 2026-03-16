package aws

import (
	"context"
	"log/slog"
	"strings"
	"testing"
)

func newTestStorage() *AWSStorage {
	return NewAWSStorage("us-east-1", slog.Default())
}

func TestAllMethodsReturnNotImplemented(t *testing.T) {
	s := newTestStorage()
	ctx := context.Background()

	tests := []struct {
		name string
		fn   func() error
	}{
		{"ListBuckets", func() error { _, err := s.ListBuckets(ctx); return err }},
		{"DescribeBucket", func() error { _, err := s.DescribeBucket(ctx, "b"); return err }},
		{"CreateBucket", func() error { return s.CreateBucket(ctx, "b", "us-east-1") }},
		{"DeleteBucket", func() error { return s.DeleteBucket(ctx, "b") }},
		{"ListObjects", func() error { _, err := s.ListObjects(ctx, "b", ""); return err }},
		{"DescribeObject", func() error { _, err := s.DescribeObject(ctx, "b", "k"); return err }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), "not yet implemented") {
				t.Errorf("expected 'not yet implemented' error, got: %v", err)
			}
		})
	}
}
