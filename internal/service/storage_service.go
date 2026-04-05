// File: internal/service/storage_service.go
package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"synkronus/internal/domain/storage"
)

type StorageService struct {
	providerFactory StorageProviderFactory
	logger          *slog.Logger
}

func NewStorageService(providerFactory StorageProviderFactory, logger *slog.Logger) *StorageService {
	return &StorageService{
		providerFactory: providerFactory,
		logger:          logger.With("service", "StorageService"),
	}
}

// withClient acquires a storage provider client, calls fn, and ensures the
// client is closed afterward. Used by service methods that return only an error.
func (s *StorageService) withClient(ctx context.Context, providerName string, fn func(client storage.Storage) error) error {
	client, err := s.getStorageClient(ctx, providerName)
	if err != nil {
		return err
	}
	defer client.Close()
	return fn(client)
}

// --- Bucket Operations ---

func (s *StorageService) ListAllBuckets(ctx context.Context, providerNames []string) ([]storage.Bucket, error) {
	if len(providerNames) == 0 {
		return nil, nil
	}

	s.logger.Debug("Starting ListAllBuckets operation", "providers", providerNames)

	return concurrentFanOut(
		ctx,
		providerNames,
		s.providerFactory.GetStorageProvider,
		func(ctx context.Context, client storage.Storage) ([]storage.Bucket, error) {
			return client.ListBuckets(ctx)
		},
		s.logger,
	)
}

func (s *StorageService) DescribeBucket(ctx context.Context, bucketName, providerName string) (storage.Bucket, error) {
	s.logger.Debug("Starting DescribeBucket operation", "bucket", bucketName, "provider", providerName)
	return withClientResult(ctx, s.getStorageClient, providerName, func(client storage.Storage) (storage.Bucket, error) {
		bucket, err := client.DescribeBucket(ctx, bucketName)
		if err != nil {
			return storage.Bucket{}, fmt.Errorf("describing bucket %q on %s: %w", bucketName, providerName, err)
		}
		return bucket, nil
	})
}

func (s *StorageService) CreateBucket(ctx context.Context, opts storage.CreateBucketOptions, providerName string) (storage.CreateBucketResult, error) {
	s.logger.Debug("Starting CreateBucket operation", "bucket", opts.Name, "provider", providerName, "location", opts.Location)
	return withClientResult(ctx, s.getStorageClient, providerName, func(client storage.Storage) (storage.CreateBucketResult, error) {
		result, err := client.CreateBucket(ctx, opts)
		if err != nil {
			return storage.CreateBucketResult{}, fmt.Errorf("creating bucket %q on %s: %w", opts.Name, providerName, err)
		}
		return result, nil
	})
}

func (s *StorageService) DeleteBucket(ctx context.Context, bucketName, providerName string) error {
	s.logger.Debug("Starting DeleteBucket operation", "bucket", bucketName, "provider", providerName)
	return s.withClient(ctx, providerName, func(client storage.Storage) error {
		if err := client.DeleteBucket(ctx, bucketName); err != nil {
			return fmt.Errorf("deleting bucket %q on %s: %w", bucketName, providerName, err)
		}
		return nil
	})
}

// --- Object Operations ---

func (s *StorageService) ListObjects(ctx context.Context, bucketName, providerName, prefix string) (storage.ObjectList, error) {
	s.logger.Debug("Starting ListObjects operation", "bucket", bucketName, "provider", providerName, "prefix", prefix)
	return withClientResult(ctx, s.getStorageClient, providerName, func(client storage.Storage) (storage.ObjectList, error) {
		objects, err := client.ListObjects(ctx, bucketName, prefix)
		if err != nil {
			return storage.ObjectList{}, fmt.Errorf("listing objects in bucket %q on %s: %w", bucketName, providerName, err)
		}
		return objects, nil
	})
}

func (s *StorageService) DescribeObject(ctx context.Context, bucketName, objectKey, providerName string) (storage.Object, error) {
	s.logger.Debug("Starting DescribeObject operation", "bucket", bucketName, "object", objectKey, "provider", providerName)
	return withClientResult(ctx, s.getStorageClient, providerName, func(client storage.Storage) (storage.Object, error) {
		object, err := client.DescribeObject(ctx, bucketName, objectKey)
		if err != nil {
			return storage.Object{}, fmt.Errorf("describing object %q in bucket %q on %s: %w", objectKey, bucketName, providerName, err)
		}
		return object, nil
	})
}

func (s *StorageService) DownloadObject(ctx context.Context, bucketName, objectKey, providerName string) (io.ReadCloser, error) {
	s.logger.Debug("Starting DownloadObject operation", "bucket", bucketName, "object", objectKey, "provider", providerName)

	client, err := s.getStorageClient(ctx, providerName)
	if err != nil {
		return nil, err
	}

	reader, err := client.DownloadObject(ctx, bucketName, objectKey)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("downloading object %q from bucket %q on %s: %w", objectKey, bucketName, providerName, err)
	}

	return &readerWithCleanup{ReadCloser: reader, cleanup: client.Close}, nil
}

func (s *StorageService) UploadObject(ctx context.Context, opts storage.UploadObjectOptions, providerName string, reader io.Reader) error {
	s.logger.Debug("Starting UploadObject operation",
		"bucket", opts.BucketName, "key", opts.ObjectKey, "provider", providerName)
	return s.withClient(ctx, providerName, func(client storage.Storage) error {
		if err := client.UploadObject(ctx, opts, reader); err != nil {
			return fmt.Errorf("uploading object %q to bucket %q on %s: %w", opts.ObjectKey, opts.BucketName, providerName, err)
		}
		return nil
	})
}

func (s *StorageService) DeleteObject(ctx context.Context, bucketName, objectKey, providerName string) error {
	s.logger.Debug("Starting DeleteObject operation",
		"bucket", bucketName, "key", objectKey, "provider", providerName)
	return s.withClient(ctx, providerName, func(client storage.Storage) error {
		if err := client.DeleteObject(ctx, bucketName, objectKey); err != nil {
			return fmt.Errorf("deleting object %q from bucket %q on %s: %w", objectKey, bucketName, providerName, err)
		}
		return nil
	})
}

func (s *StorageService) CopyObject(ctx context.Context, srcBucket, srcKey, destBucket, destKey, providerName string) error {
	s.logger.Debug("Starting CopyObject operation",
		"srcBucket", srcBucket, "srcKey", srcKey,
		"destBucket", destBucket, "destKey", destKey, "provider", providerName)
	return s.withClient(ctx, providerName, func(client storage.Storage) error {
		if err := client.CopyObject(ctx, srcBucket, srcKey, destBucket, destKey); err != nil {
			return fmt.Errorf("copying object %q/%q to %q/%q on %s: %w", srcBucket, srcKey, destBucket, destKey, providerName, err)
		}
		return nil
	})
}

// readerWithCleanup wraps an io.ReadCloser to run a cleanup function (e.g., client.Close)
// when the reader is closed. This ensures the provider client outlives the reader.
type readerWithCleanup struct {
	io.ReadCloser
	cleanup func() error
}

func (r *readerWithCleanup) Close() error {
	readErr := r.ReadCloser.Close()
	cleanupErr := r.cleanup()
	return errors.Join(readErr, cleanupErr)
}

func (s *StorageService) getStorageClient(ctx context.Context, providerName string) (storage.Storage, error) {
	client, err := s.providerFactory.GetStorageProvider(ctx, providerName)
	if err != nil {
		return nil, fmt.Errorf("initializing storage provider %s: %w", providerName, err)
	}
	return client, nil
}
