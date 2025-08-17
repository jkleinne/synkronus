// File: pkg/storage/storage.go
package storage

import (
	"context"
	"synkronus/pkg/common"
)

type Storage interface {
	ListBuckets(ctx context.Context) ([]Bucket, error)

	DescribeBucket(ctx context.Context, bucketName string) (Bucket, error)

	ProviderName() common.Provider

	Close() error
}
