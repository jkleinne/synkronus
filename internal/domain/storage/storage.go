package storage

import (
	"context"
	"io"
	"synkronus/internal/domain"
)

// Storage defines the interface for interacting with cloud storage buckets
// and objects across providers. Each provider implementation must satisfy
// this interface.
type Storage interface {
	// --- Bucket Operations ---
	ListBuckets(ctx context.Context) ([]Bucket, error)

	DescribeBucket(ctx context.Context, bucketName string) (Bucket, error)

	CreateBucket(ctx context.Context, opts CreateBucketOptions) (CreateBucketResult, error)

	DeleteBucket(ctx context.Context, bucketName string) error

	// --- Object Operations ---
	ListObjects(ctx context.Context, bucketName string, prefix string) (ObjectList, error)

	DescribeObject(ctx context.Context, bucketName string, objectKey string) (Object, error)

	DownloadObject(ctx context.Context, bucketName string, objectKey string) (io.ReadCloser, error)

	UploadObject(ctx context.Context, opts UploadObjectOptions, reader io.Reader) error

	DeleteObject(ctx context.Context, bucketName, objectKey string) error

	CopyObject(ctx context.Context, srcBucket, srcKey, destBucket, destKey string) error

	ProviderName() domain.Provider

	Close() error
}
