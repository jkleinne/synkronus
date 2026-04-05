package aws

import (
	"context"
	"errors"
	"log/slog"
	"synkronus/internal/domain"
	"synkronus/internal/domain/storage"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	smithy "github.com/aws/smithy-go"
	"golang.org/x/sync/errgroup"
)

// s3DefaultRegion is the implied region when S3 returns an empty LocationConstraint.
const s3DefaultRegion = "us-east-1"

// bucketFetcher describes a single concurrent detail fetch for DescribeBucket.
type bucketFetcher struct {
	label                 string
	tolerateNotConfigured bool
	fetch                 func(ctx context.Context) error
}

// run wraps the fetch with centralized error handling and always returns nil.
// "Not configured" errors are silently tolerated when flagged; all other failures
// are logged as warnings. Errors do not propagate through errgroup — the
// corresponding bucket field is left at its zero value.
func (f bucketFetcher) run(ctx context.Context, logger *slog.Logger, bucketName string) func() error {
	return func() error {
		if err := f.fetch(ctx); err != nil {
			if f.tolerateNotConfigured && isS3NotConfiguredError(err) {
				return nil
			}
			logger.Warn("Could not retrieve "+f.label, "bucket", bucketName, "error", err)
		}
		return nil
	}
}

func (s *AWSStorage) DescribeBucket(ctx context.Context, bucketName string) (storage.Bucket, error) {
	s.logger.Debug("Starting AWS DescribeBucket operation", "bucket", bucketName)

	bucket := storage.Bucket{
		Name:       bucketName,
		Provider:   domain.AWS,
		UsageBytes: -1,
	}

	eg, egCtx := errgroup.WithContext(ctx)
	for _, f := range s.bucketFetchers(bucketName, &bucket) {
		eg.Go(f.run(egCtx, s.logger, bucketName))
	}

	// All fetchers log-warn-continue; no error is propagated.
	eg.Wait()

	return bucket, nil
}

func (s *AWSStorage) bucketFetchers(bucketName string, bucket *storage.Bucket) []bucketFetcher {
	return []bucketFetcher{
		{"location", false, func(ctx context.Context) error {
			out, err := s.client.GetBucketLocation(ctx, &s3.GetBucketLocationInput{Bucket: &bucketName})
			if err != nil {
				return err
			}
			bucket.Location = string(out.LocationConstraint)
			if bucket.Location == "" {
				bucket.Location = s3DefaultRegion
			}
			return nil
		}},
		{"versioning", false, func(ctx context.Context) error {
			out, err := s.client.GetBucketVersioning(ctx, &s3.GetBucketVersioningInput{Bucket: &bucketName})
			if err != nil {
				return err
			}
			bucket.Versioning = mapVersioning(out.Status)
			return nil
		}},
		{"encryption", true, func(ctx context.Context) error {
			out, err := s.client.GetBucketEncryption(ctx, &s3.GetBucketEncryptionInput{Bucket: &bucketName})
			if err != nil {
				return err
			}
			if out.ServerSideEncryptionConfiguration != nil {
				bucket.Encryption = mapEncryption(out.ServerSideEncryptionConfiguration.Rules)
			}
			return nil
		}},
		{"lifecycle configuration", true, func(ctx context.Context) error {
			out, err := s.client.GetBucketLifecycleConfiguration(ctx, &s3.GetBucketLifecycleConfigurationInput{Bucket: &bucketName})
			if err != nil {
				return err
			}
			bucket.LifecycleRules = mapLifecycleRules(out.Rules)
			return nil
		}},
		{"tags", true, func(ctx context.Context) error {
			out, err := s.client.GetBucketTagging(ctx, &s3.GetBucketTaggingInput{Bucket: &bucketName})
			if err != nil {
				return err
			}
			bucket.Labels = mapTags(out.TagSet)
			return nil
		}},
		{"bucket policy", true, func(ctx context.Context) error {
			out, err := s.client.GetBucketPolicy(ctx, &s3.GetBucketPolicyInput{Bucket: &bucketName})
			if err != nil {
				return err
			}
			if out.Policy != nil {
				statements, parseErr := parseBucketPolicy(*out.Policy)
				if parseErr != nil {
					return parseErr
				}
				bucket.IAMPolicy = &storage.IAMPolicy{Statements: statements}
			}
			return nil
		}},
		{"ACLs", false, func(ctx context.Context) error {
			out, err := s.client.GetBucketAcl(ctx, &s3.GetBucketAclInput{Bucket: &bucketName})
			if err != nil {
				return err
			}
			bucket.ACLs = mapACLGrants(out.Owner, out.Grants)
			return nil
		}},
		{"public access block", true, func(ctx context.Context) error {
			out, err := s.client.GetPublicAccessBlock(ctx, &s3.GetPublicAccessBlockInput{Bucket: &bucketName})
			if err != nil {
				return err
			}
			bucket.PublicAccessPrevention = mapPublicAccessBlock(out.PublicAccessBlockConfiguration)
			return nil
		}},
		{"logging", false, func(ctx context.Context) error {
			out, err := s.client.GetBucketLogging(ctx, &s3.GetBucketLoggingInput{Bucket: &bucketName})
			if err != nil {
				return err
			}
			bucket.Logging = mapLogging(out.LoggingEnabled)
			return nil
		}},
		{"object lock configuration", true, func(ctx context.Context) error {
			out, err := s.client.GetObjectLockConfiguration(ctx, &s3.GetObjectLockConfigurationInput{Bucket: &bucketName})
			if err != nil {
				return err
			}
			bucket.RetentionPolicy = mapRetentionPolicy(out.ObjectLockConfiguration)
			return nil
		}},
	}
}

// isS3NotConfiguredError checks if an error is an S3 "not configured" error
// (e.g., NoSuchTagSet, NoSuchBucketPolicy, NoSuchLifecycleConfiguration).
func isS3NotConfiguredError(err error) bool {
	var apiErr smithy.APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	switch apiErr.ErrorCode() {
	case "NoSuchTagSet",
		"NoSuchBucketPolicy",
		"NoSuchLifecycleConfiguration",
		"ServerSideEncryptionConfigurationNotFoundError",
		"ObjectLockConfigurationNotFoundError",
		"NoSuchPublicAccessBlockConfiguration":
		return true
	default:
		return false
	}
}
