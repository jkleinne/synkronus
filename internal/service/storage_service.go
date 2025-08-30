// File: internal/service/storage_service.go
package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"synkronus/internal/provider/factory"
	"synkronus/pkg/storage"
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

	var allBuckets []storage.Bucket
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, pName := range providerNames {
		wg.Add(1)
		go func(pName string) {
			defer wg.Done()

			client, err := s.providerFactory.GetStorageProvider(ctx, pName)
			if err != nil {
				s.logger.Error("Failed to initialize provider client", "provider", pName, "error", err)
				return
			}
			defer client.Close()

			buckets, err := client.ListBuckets(ctx)
			if err != nil {
				s.logger.Error("Failed to list buckets from provider", "provider", pName, "error", err)
				return
			}

			// Safely append successful results
			mu.Lock()
			allBuckets = append(allBuckets, buckets...)
			mu.Unlock()

			s.logger.Debug("Successfully fetched buckets", "provider", pName, "count", len(buckets))
		}(pName)
	}

	wg.Wait()

	// The operation itself succeeded, even if some providers failed
	return allBuckets, nil
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
		return storage.Bucket{}, err
	}
	return bucket, nil
}

func (s *StorageService) CreateBucket(ctx context.Context, bucketName, providerName, location string) error {
	s.logger.Debug("Starting CreateBucket operation", "bucket", bucketName, "provider", providerName, "location", location)

	client, err := s.getStorageClient(ctx, providerName)
	if err != nil {
		return err
	}
	defer client.Close()

	err = client.CreateBucket(ctx, bucketName, location)
	if err != nil {
		s.logger.Error("Failed to create bucket", "bucket", bucketName, "provider", providerName, "error", err)
		return err
	}
	return nil
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
		return err
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
		return storage.ObjectList{}, err
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
		return storage.Object{}, err
	}
	return object, nil
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
