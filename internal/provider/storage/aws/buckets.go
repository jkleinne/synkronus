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

func (s *AWSStorage) CreateBucket(ctx context.Context, opts storage.CreateBucketOptions) error {
	s.logger.Debug("Starting AWS CreateBucket operation", "bucket", opts.Name)

	input := &s3.CreateBucketInput{
		Bucket: &opts.Name,
	}

	// us-east-1 is the default region and must not specify a LocationConstraint
	if s.region != "us-east-1" {
		input.CreateBucketConfiguration = &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(s.region),
		}
	}

	if _, err := s.client.CreateBucket(ctx, input); err != nil {
		return fmt.Errorf("failed to create S3 bucket: %w", err)
	}

	s.applyPostCreateOptions(ctx, opts)

	return nil
}

// applyPostCreateOptions applies optional settings after bucket creation.
// Failures are logged as warnings, not returned as errors.
func (s *AWSStorage) applyPostCreateOptions(ctx context.Context, opts storage.CreateBucketOptions) {
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
			s.logger.Warn("Failed to set versioning on new bucket", "bucket", opts.Name, "error", err)
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
			s.logger.Warn("Failed to set tags on new bucket", "bucket", opts.Name, "error", err)
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
			s.logger.Warn("Failed to set public access block on new bucket", "bucket", opts.Name, "error", err)
		}
	}
	// opts.UniformBucketLevelAccess is GCP-only — ignored for AWS
}

func (s *AWSStorage) DeleteBucket(ctx context.Context, bucketName string) error {
	s.logger.Debug("Starting AWS DeleteBucket operation", "bucket", bucketName)

	if _, err := s.client.DeleteBucket(ctx, &s3.DeleteBucketInput{Bucket: &bucketName}); err != nil {
		return fmt.Errorf("failed to delete S3 bucket: %w", err)
	}
	return nil
}

func (s *AWSStorage) UpdateBucket(ctx context.Context, opts storage.UpdateBucketOptions) error {
	s.logger.Debug("Starting AWS UpdateBucket operation", "bucket", opts.Name)

	if err := s.updateBucketTags(ctx, opts); err != nil {
		return err
	}

	if opts.Versioning != nil {
		status := types.BucketVersioningStatusSuspended
		if *opts.Versioning {
			status = types.BucketVersioningStatusEnabled
		}
		if _, err := s.client.PutBucketVersioning(ctx, &s3.PutBucketVersioningInput{
			Bucket: &opts.Name,
			VersioningConfiguration: &types.VersioningConfiguration{
				Status: status,
			},
		}); err != nil {
			return fmt.Errorf("failed to update versioning: %w", err)
		}
	}

	return nil
}

// updateBucketTags performs a read-modify-write on bucket tags to support
// SetLabels/RemoveLabels without replacing unrelated tags.
func (s *AWSStorage) updateBucketTags(ctx context.Context, opts storage.UpdateBucketOptions) error {
	if len(opts.SetLabels) == 0 && len(opts.RemoveLabels) == 0 {
		return nil
	}

	// Read existing tags
	existing := make(map[string]string)
	out, err := s.client.GetBucketTagging(ctx, &s3.GetBucketTaggingInput{Bucket: &opts.Name})
	if err != nil {
		if !isS3NotConfiguredError(err) {
			return fmt.Errorf("failed to get bucket tags: %w", err)
		}
		// NoSuchTagSet — start from empty
	} else {
		for _, tag := range out.TagSet {
			existing[derefString(tag.Key)] = derefString(tag.Value)
		}
	}

	// Apply changes
	for k, v := range opts.SetLabels {
		existing[k] = v
	}
	for _, k := range opts.RemoveLabels {
		delete(existing, k)
	}

	// Write back
	if len(existing) == 0 {
		_, err := s.client.DeleteBucketTagging(ctx, &s3.DeleteBucketTaggingInput{Bucket: &opts.Name})
		if err != nil {
			return fmt.Errorf("failed to delete bucket tags: %w", err)
		}
		return nil
	}

	tags := make([]types.Tag, 0, len(existing))
	for k, v := range existing {
		tags = append(tags, types.Tag{Key: strPtr(k), Value: strPtr(v)})
	}
	if _, err := s.client.PutBucketTagging(ctx, &s3.PutBucketTaggingInput{
		Bucket:  &opts.Name,
		Tagging: &types.Tagging{TagSet: tags},
	}); err != nil {
		return fmt.Errorf("failed to update bucket tags: %w", err)
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
