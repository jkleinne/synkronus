// File: internal/service/storage_service.go
package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"synkronus/internal/domain/storage"
	"synkronus/internal/provider/factory"
)

type StorageService struct {
	providerFactory *factory.Factory
	logger          *slog.Logger
}

func NewStorageService(providerFactory *factory.Factory, logger *slog.Logger) *StorageService {
	return &StorageService{
		providerFactory: providerFactory,
		logger:          logger.With("service", "StorageService"),
	}
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

	client, err := s.getStorageClient(ctx, providerName)
	if err != nil {
		return storage.Bucket{}, err
	}
	defer client.Close()

	bucket, err := client.DescribeBucket(ctx, bucketName)
	if err != nil {
		s.logger.Error("Failed to describe bucket", "bucket", bucketName, "provider", providerName, "error", err)
		return storage.Bucket{}, fmt.Errorf("describing bucket %q on %s: %w", bucketName, providerName, err)
	}
	return bucket, nil
}

func (s *StorageService) CreateBucket(ctx context.Context, opts storage.CreateBucketOptions, providerName string) (storage.CreateBucketResult, error) {
	s.logger.Debug("Starting CreateBucket operation", "bucket", opts.Name, "provider", providerName, "location", opts.Location)

	client, err := s.getStorageClient(ctx, providerName)
	if err != nil {
		return storage.CreateBucketResult{}, err
	}
	defer client.Close()

	result, err := client.CreateBucket(ctx, opts)
	if err != nil {
		s.logger.Error("Failed to create bucket", "bucket", opts.Name, "provider", providerName, "error", err)
		return storage.CreateBucketResult{}, fmt.Errorf("creating bucket %q on %s: %w", opts.Name, providerName, err)
	}
	return result, nil
}

func (s *StorageService) DeleteBucket(ctx context.Context, bucketName, providerName string) error {
	s.logger.Debug("Starting DeleteBucket operation", "bucket", bucketName, "provider", providerName)

	client, err := s.getStorageClient(ctx, providerName)
	if err != nil {
		return err
	}
	defer client.Close()

	err = client.DeleteBucket(ctx, bucketName)
	if err != nil {
		s.logger.Error("Failed to delete bucket", "bucket", bucketName, "provider", providerName, "error", err)
		return fmt.Errorf("deleting bucket %q on %s: %w", bucketName, providerName, err)
	}
	return nil
}

// --- Object Operations ---

func (s *StorageService) ListObjects(ctx context.Context, bucketName, providerName, prefix string) (storage.ObjectList, error) {
	s.logger.Debug("Starting ListObjects operation", "bucket", bucketName, "provider", providerName, "prefix", prefix)

	client, err := s.getStorageClient(ctx, providerName)
	if err != nil {
		return storage.ObjectList{}, err
	}
	defer client.Close()

	objects, err := client.ListObjects(ctx, bucketName, prefix)
	if err != nil {
		s.logger.Error("Failed to list objects", "bucket", bucketName, "provider", providerName, "error", err)
		return storage.ObjectList{}, fmt.Errorf("listing objects in bucket %q on %s: %w", bucketName, providerName, err)
	}
	return objects, nil
}

func (s *StorageService) DescribeObject(ctx context.Context, bucketName, objectKey, providerName string) (storage.Object, error) {
	s.logger.Debug("Starting DescribeObject operation", "bucket", bucketName, "object", objectKey, "provider", providerName)

	client, err := s.getStorageClient(ctx, providerName)
	if err != nil {
		return storage.Object{}, err
	}
	defer client.Close()

	object, err := client.DescribeObject(ctx, bucketName, objectKey)
	if err != nil {
		s.logger.Error("Failed to describe object", "bucket", bucketName, "object", objectKey, "provider", providerName, "error", err)
		return storage.Object{}, fmt.Errorf("describing object %q in bucket %q on %s: %w", objectKey, bucketName, providerName, err)
	}
	return object, nil
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
		s.logger.Error("Failed to download object", "bucket", bucketName, "object", objectKey, "provider", providerName, "error", err)
		return nil, fmt.Errorf("downloading object %q from bucket %q on %s: %w", objectKey, bucketName, providerName, err)
	}

	return &readerWithCleanup{ReadCloser: reader, cleanup: client.Close}, nil
}

func (s *StorageService) UploadObject(ctx context.Context, opts storage.UploadObjectOptions, providerName string, reader io.Reader) error {
	s.logger.Debug("Starting UploadObject operation",
		"bucket", opts.BucketName, "key", opts.ObjectKey, "provider", providerName)

	client, err := s.getStorageClient(ctx, providerName)
	if err != nil {
		return err
	}
	defer client.Close()

	if err := client.UploadObject(ctx, opts, reader); err != nil {
		s.logger.Error("Failed to upload object",
			"bucket", opts.BucketName, "key", opts.ObjectKey, "provider", providerName, "error", err)
		return fmt.Errorf("uploading object %q to bucket %q on %s: %w", opts.ObjectKey, opts.BucketName, providerName, err)
	}
	return nil
}

func (s *StorageService) DeleteObject(ctx context.Context, bucketName, objectKey, providerName string) error {
	s.logger.Debug("Starting DeleteObject operation",
		"bucket", bucketName, "key", objectKey, "provider", providerName)

	client, err := s.getStorageClient(ctx, providerName)
	if err != nil {
		return err
	}
	defer client.Close()

	if err := client.DeleteObject(ctx, bucketName, objectKey); err != nil {
		s.logger.Error("Failed to delete object",
			"bucket", bucketName, "key", objectKey, "provider", providerName, "error", err)
		return fmt.Errorf("deleting object %q from bucket %q on %s: %w", objectKey, bucketName, providerName, err)
	}
	return nil
}

func (s *StorageService) CopyObject(ctx context.Context, srcBucket, srcKey, destBucket, destKey, providerName string) error {
	s.logger.Debug("Starting CopyObject operation",
		"srcBucket", srcBucket, "srcKey", srcKey,
		"destBucket", destBucket, "destKey", destKey, "provider", providerName)

	client, err := s.getStorageClient(ctx, providerName)
	if err != nil {
		return err
	}
	defer client.Close()

	if err := client.CopyObject(ctx, srcBucket, srcKey, destBucket, destKey); err != nil {
		s.logger.Error("Failed to copy object",
			"srcBucket", srcBucket, "srcKey", srcKey,
			"destBucket", destBucket, "destKey", destKey, "provider", providerName, "error", err)
		return fmt.Errorf("copying object %q/%q to %q/%q on %s: %w", srcBucket, srcKey, destBucket, destKey, providerName, err)
	}
	return nil
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

// Helper to initialize the storage client and handle common error logging
func (s *StorageService) getStorageClient(ctx context.Context, providerName string) (storage.Storage, error) {
	client, err := s.providerFactory.GetStorageProvider(ctx, providerName)
	if err != nil {
		s.logger.Error("Failed to initialize provider", "provider", providerName, "error", err)
		return nil, fmt.Errorf("error initializing provider: %w", err)
	}
	return client, nil
}
