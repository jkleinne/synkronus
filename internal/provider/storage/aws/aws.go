// File: internal/provider/storage/aws/aws.go
package aws

import (
	"context"
	"fmt"
	"log/slog"
	"synkronus/internal/config"
	"synkronus/internal/domain"
	"synkronus/internal/domain/storage"
	"synkronus/internal/provider/registry"
)

func init() {
	registry.RegisterProvider("aws", registry.Registration[storage.Storage]{
		ConfigCheck: isConfigured,
		Initializer: initialize,
	})
}

// Checks if the AWS configuration block is present and the region is set
func isConfigured(cfg *config.Config) bool {
	return cfg.AWS != nil && cfg.AWS.Region != ""
}

// Initializes the AWS storage client from the configuration
func initialize(ctx context.Context, cfg *config.Config, logger *slog.Logger) (storage.Storage, error) {
	if !isConfigured(cfg) {
		return nil, fmt.Errorf("AWS configuration missing or incomplete")
	}
	return NewAWSStorage(cfg.AWS.Region, logger), nil
}

type AWSStorage struct {
	region string
	logger *slog.Logger
}

var _ storage.Storage = (*AWSStorage)(nil)

func NewAWSStorage(region string, logger *slog.Logger) *AWSStorage {
	return &AWSStorage{
		region: region,
		logger: logger,
	}
}

func (s *AWSStorage) ProviderName() domain.Provider {
	return domain.AWS
}

func (s *AWSStorage) ListBuckets(ctx context.Context) ([]storage.Bucket, error) {
	return nil, fmt.Errorf("AWS ListBuckets is not yet implemented")
}

func (s *AWSStorage) DescribeBucket(ctx context.Context, bucketName string) (storage.Bucket, error) {
	return storage.Bucket{}, fmt.Errorf("AWS DescribeBucket is not yet implemented")
}

func (s *AWSStorage) CreateBucket(ctx context.Context, opts storage.CreateBucketOptions) error {
	return fmt.Errorf("AWS CreateBucket is not yet implemented")
}

func (s *AWSStorage) DeleteBucket(ctx context.Context, bucketName string) error {
	return fmt.Errorf("AWS DeleteBucket is not yet implemented")
}

func (s *AWSStorage) ListObjects(ctx context.Context, bucketName string, prefix string) (storage.ObjectList, error) {
	s.logger.Debug("Listing AWS objects (placeholder)", "bucket", bucketName, "prefix", prefix)
	return storage.ObjectList{}, fmt.Errorf("AWS ListObjects is not yet implemented")
}

func (s *AWSStorage) DescribeObject(ctx context.Context, bucketName string, objectKey string) (storage.Object, error) {
	s.logger.Debug("Describing AWS object (placeholder)", "bucket", bucketName, "objectKey", objectKey)
	return storage.Object{}, fmt.Errorf("AWS DescribeObject is not yet implemented")
}

func (s *AWSStorage) Close() error {
	// Nothing to close in the placeholder
	return nil
}
