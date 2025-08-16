package storage

import "context"

type Storage interface {
	ListBuckets(ctx context.Context) ([]Bucket, error)

	DescribeBucket(ctx context.Context, bucketName string) (Bucket, error)

	ProviderName() Provider

	Close() error
}
