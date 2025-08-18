// File: internal/service/storage_service.go
package service

import (
	"context"
	"fmt"
	"sync"

	"golang.org/x/sync/errgroup"
	"synkronus/internal/provider"
	"synkronus/pkg/storage"
)

type StorageService struct {
	providerFactory *provider.Factory
}

func NewStorageService(providerFactory *provider.Factory) *StorageService {
	return &StorageService{
		providerFactory: providerFactory,
	}
}

func (s *StorageService) ListAllBuckets(ctx context.Context, useGCP, useAWS bool) ([]storage.Bucket, error) {
	providersToQuery := s.providerFactory.GetStorageProviders(useGCP, useAWS)

	if len(providersToQuery) == 0 {
		return nil, nil
	}

	var allBuckets []storage.Bucket
	var mu sync.Mutex
	g, gCtx := errgroup.WithContext(ctx)

	for _, pName := range providersToQuery {
		pName := pName // Capture pName for the goroutine
		g.Go(func() error {
			client, err := s.providerFactory.GetStorageProvider(gCtx, pName)
			if err != nil {
				return fmt.Errorf("initializing client for %s: %w", pName, err)
			}
			defer client.Close()

			buckets, err := client.ListBuckets(gCtx)
			if err != nil {
				return fmt.Errorf("listing buckets from %s: %w", pName, err)
			}

			mu.Lock()
			allBuckets = append(allBuckets, buckets...)
			mu.Unlock()

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return allBuckets, nil
}

func (s *StorageService) DescribeBucket(ctx context.Context, bucketName, providerName string) (storage.Bucket, error) {
	client, err := s.providerFactory.GetStorageProvider(ctx, providerName)
	if err != nil {
		return storage.Bucket{}, fmt.Errorf("error initializing provider: %w", err)
	}
	defer client.Close()

	return client.DescribeBucket(ctx, bucketName)
}

func (s *StorageService) CreateBucket(ctx context.Context, bucketName, providerName, location string) error {
	client, err := s.providerFactory.GetStorageProvider(ctx, providerName)
	if err != nil {
		return fmt.Errorf("error initializing provider: %w", err)
	}
	defer client.Close()

	return client.CreateBucket(ctx, bucketName, location)
}

func (s *StorageService) DeleteBucket(ctx context.Context, bucketName, providerName string) error {
	client, err := s.providerFactory.GetStorageProvider(ctx, providerName)
	if err != nil {
		return fmt.Errorf("error initializing provider: %w", err)
	}
	defer client.Close()

	return client.DeleteBucket(ctx, bucketName)
}
