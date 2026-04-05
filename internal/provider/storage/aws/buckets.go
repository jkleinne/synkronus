package aws

import (
	"context"
	"fmt"
	"synkronus/internal/domain"
	"synkronus/internal/domain/storage"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func (s *AWSStorage) ListBuckets(ctx context.Context) ([]storage.Bucket, error) {
	s.logger.Debug("Starting AWS ListBuckets operation")

	input := &s3.ListBucketsInput{
		BucketRegion: &s.region,
	}

	var buckets []storage.Bucket
	paginator := s3.NewListBucketsPaginator(s.client, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list S3 buckets: %w", err)
		}

		for _, b := range page.Buckets {
			bucket := storage.Bucket{
				Name:         derefString(b.Name),
				Provider:     domain.AWS,
				Location:     s.region,
				StorageClass: "STANDARD",
				UsageBytes:   -1,
			}
			if b.CreationDate != nil {
				bucket.CreatedAt = *b.CreationDate
			}
			buckets = append(buckets, bucket)
		}
	}

	return buckets, nil
}

func (s *AWSStorage) CreateBucket(ctx context.Context, opts storage.CreateBucketOptions) (storage.CreateBucketResult, error) {
	s.logger.Debug("Starting AWS CreateBucket operation", "bucket", opts.Name)

	input := &s3.CreateBucketInput{
		Bucket: &opts.Name,
	}

	// us-east-1 is the default region and must not specify a LocationConstraint
	if s.region != s3DefaultRegion {
		input.CreateBucketConfiguration = &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(s.region),
		}
	}

	if _, err := s.client.CreateBucket(ctx, input); err != nil {
		return storage.CreateBucketResult{}, fmt.Errorf("failed to create S3 bucket: %w", err)
	}

	warnings := s.applyPostCreateOptions(ctx, opts)

	return storage.CreateBucketResult{Warnings: warnings}, nil
}

// applyPostCreateOptions applies optional settings after bucket creation.
// Returns warning messages for any settings that failed to apply.
func (s *AWSStorage) applyPostCreateOptions(ctx context.Context, opts storage.CreateBucketOptions) []string {
	var warnings []string

	if opts.Versioning != nil {
		status := types.BucketVersioningStatusSuspended
		if *opts.Versioning {
			status = types.BucketVersioningStatusEnabled
		}
		_, err := s.client.PutBucketVersioning(ctx, &s3.PutBucketVersioningInput{
			Bucket: &opts.Name,
			VersioningConfiguration: &types.VersioningConfiguration{
				Status: status,
			},
		})
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("failed to set versioning: %v", err))
		}
	}

	if len(opts.Labels) > 0 {
		tags := make([]types.Tag, 0, len(opts.Labels))
		for k, v := range opts.Labels {
			tags = append(tags, types.Tag{Key: strPtr(k), Value: strPtr(v)})
		}
		_, err := s.client.PutBucketTagging(ctx, &s3.PutBucketTaggingInput{
			Bucket:  &opts.Name,
			Tagging: &types.Tagging{TagSet: tags},
		})
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("failed to set tags: %v", err))
		}
	}

	if opts.PublicAccessPrevention != nil {
		blockAll := *opts.PublicAccessPrevention == storage.PublicAccessPreventionEnforced
		_, err := s.client.PutPublicAccessBlock(ctx, &s3.PutPublicAccessBlockInput{
			Bucket: &opts.Name,
			PublicAccessBlockConfiguration: &types.PublicAccessBlockConfiguration{
				BlockPublicAcls:       &blockAll,
				BlockPublicPolicy:     &blockAll,
				IgnorePublicAcls:      &blockAll,
				RestrictPublicBuckets: &blockAll,
			},
		})
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("failed to set public access block: %v", err))
		}
	}
	// opts.UniformBucketLevelAccess is GCP-only — ignored for AWS

	return warnings
}

func (s *AWSStorage) DeleteBucket(ctx context.Context, bucketName string) error {
	s.logger.Debug("Starting AWS DeleteBucket operation", "bucket", bucketName)

	if _, err := s.client.DeleteBucket(ctx, &s3.DeleteBucketInput{Bucket: &bucketName}); err != nil {
		return fmt.Errorf("failed to delete S3 bucket: %w", err)
	}
	return nil
}

// derefString safely dereferences a string pointer, returning "" if nil.
func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// strPtr returns a pointer to the given string.
func strPtr(s string) *string {
	return &s
}
