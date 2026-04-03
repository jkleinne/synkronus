package aws

import (
	"context"
	"fmt"
	"log/slog"
	"synkronus/internal/config"
	"synkronus/internal/domain"
	"synkronus/internal/domain/storage"
	"synkronus/internal/provider/registry"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func init() {
	registry.RegisterProvider("aws", registry.Registration[storage.Storage]{
		ConfigCheck: isConfigured,
		Initializer: initialize,
	})
}

// isConfigured checks if the AWS configuration block is present and the region is set.
func isConfigured(cfg *config.Config) bool {
	return cfg.AWS != nil && cfg.AWS.Region != ""
}

// initialize creates an AWS storage client from the configuration.
func initialize(ctx context.Context, cfg *config.Config, logger *slog.Logger) (storage.Storage, error) {
	if !isConfigured(cfg) {
		return nil, fmt.Errorf("AWS configuration missing or incomplete")
	}
	return NewAWSStorage(ctx, cfg.AWS, logger)
}

// AWSStorage implements storage.Storage using the AWS S3 API.
type AWSStorage struct {
	client *s3.Client
	region string
	logger *slog.Logger
}

var _ storage.Storage = (*AWSStorage)(nil)

// NewAWSStorage creates a new S3 storage client. If cfg.Endpoint is set, the client
// targets that URL (e.g., LocalStack) instead of real AWS endpoints.
func NewAWSStorage(ctx context.Context, cfg *config.AWSConfig, logger *slog.Logger) (*AWSStorage, error) {
	sdkCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(cfg.Region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS SDK config: %w", err)
	}

	var s3Opts []func(*s3.Options)
	if cfg.Endpoint != "" {
		s3Opts = append(s3Opts, func(o *s3.Options) {
			o.BaseEndpoint = &cfg.Endpoint
			o.UsePathStyle = true // Required for LocalStack and most S3-compatible services
		})
	}

	client := s3.NewFromConfig(sdkCfg, s3Opts...)

	return &AWSStorage{
		client: client,
		region: cfg.Region,
		logger: logger,
	}, nil
}

func (s *AWSStorage) ProviderName() domain.Provider {
	return domain.AWS
}

func (s *AWSStorage) Close() error {
	// AWS SDK v2 clients don't require explicit cleanup
	return nil
}
