// File: pkg/storage/storage.go
package storage

import (
	"context"
	"synkronus/pkg/common"
)

type Storage interface {
	ListBuckets(ctx context.Context) ([]Bucket, error)

	DescribeBucket(ctx context.Context, bucketName string) (Bucket, error)

	CreateBucket(ctx context.Context, bucketName string, location string) error

	DeleteBucket(ctx context.Context, bucketName string) error

	ProviderName() common.Provider

	Close() error
}
