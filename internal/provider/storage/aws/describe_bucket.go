package aws

import (
	"context"
	"errors"
	"fmt"
	"synkronus/internal/domain"
	"synkronus/internal/domain/storage"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	smithy "github.com/aws/smithy-go"
	"golang.org/x/sync/errgroup"
)

func (s *AWSStorage) DescribeBucket(ctx context.Context, bucketName string) (storage.Bucket, error) {
	s.logger.Debug("Starting AWS DescribeBucket operation", "bucket", bucketName)

	bucket := storage.Bucket{
		Name:       bucketName,
		Provider:   domain.AWS,
		UsageBytes: -1,
	}

	var (
		location     string
		versioning   *storage.Versioning
		encryption   *storage.Encryption
		lifecycle    []storage.LifecycleRule
		labels       map[string]string
		iamPolicy    *storage.IAMPolicy
		acls         []storage.ACLRule
		publicAccess string
		logging      *storage.Logging
		retention    *storage.RetentionPolicy
	)

	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		out, err := s.client.GetBucketLocation(egCtx, &s3.GetBucketLocationInput{Bucket: &bucketName})
		if err != nil {
			s.logger.Warn("Could not retrieve bucket location", "bucket", bucketName, "error", err)
			return nil
		}
		loc := string(out.LocationConstraint)
		if loc == "" {
			loc = "us-east-1"
		}
		location = loc
		return nil
	})

	eg.Go(func() error {
		out, err := s.client.GetBucketVersioning(egCtx, &s3.GetBucketVersioningInput{Bucket: &bucketName})
		if err != nil {
			s.logger.Warn("Could not retrieve versioning", "bucket", bucketName, "error", err)
			return nil
		}
		versioning = mapVersioning(out.Status)
		return nil
	})

	eg.Go(func() error {
		out, err := s.client.GetBucketEncryption(egCtx, &s3.GetBucketEncryptionInput{Bucket: &bucketName})
		if err != nil {
			if isS3NotFoundError(err) {
				return nil
			}
			s.logger.Warn("Could not retrieve encryption", "bucket", bucketName, "error", err)
			return nil
		}
		if out.ServerSideEncryptionConfiguration != nil {
			encryption = mapEncryption(out.ServerSideEncryptionConfiguration.Rules)
		}
		return nil
	})

	eg.Go(func() error {
		out, err := s.client.GetBucketLifecycleConfiguration(egCtx, &s3.GetBucketLifecycleConfigurationInput{Bucket: &bucketName})
		if err != nil {
			if isS3NotFoundError(err) {
				return nil
			}
			s.logger.Warn("Could not retrieve lifecycle configuration", "bucket", bucketName, "error", err)
			return nil
		}
		lifecycle = mapLifecycleRules(out.Rules)
		return nil
	})

	eg.Go(func() error {
		out, err := s.client.GetBucketTagging(egCtx, &s3.GetBucketTaggingInput{Bucket: &bucketName})
		if err != nil {
			if isS3NotFoundError(err) {
				return nil
			}
			s.logger.Warn("Could not retrieve tags", "bucket", bucketName, "error", err)
			return nil
		}
		labels = mapTags(out.TagSet)
		return nil
	})

	eg.Go(func() error {
		out, err := s.client.GetBucketPolicy(egCtx, &s3.GetBucketPolicyInput{Bucket: &bucketName})
		if err != nil {
			if isS3NotFoundError(err) {
				return nil
			}
			s.logger.Warn("Could not retrieve bucket policy", "bucket", bucketName, "error", err)
			return nil
		}
		if out.Policy != nil {
			statements, parseErr := parseBucketPolicy(*out.Policy)
			if parseErr != nil {
				s.logger.Warn("Could not parse bucket policy", "bucket", bucketName, "error", parseErr)
				return nil
			}
			iamPolicy = &storage.IAMPolicy{Statements: statements}
		}
		return nil
	})

	eg.Go(func() error {
		out, err := s.client.GetBucketAcl(egCtx, &s3.GetBucketAclInput{Bucket: &bucketName})
		if err != nil {
			s.logger.Warn("Could not retrieve ACLs", "bucket", bucketName, "error", err)
			return nil
		}
		acls = mapACLGrants(out.Owner, out.Grants)
		return nil
	})

	eg.Go(func() error {
		out, err := s.client.GetPublicAccessBlock(egCtx, &s3.GetPublicAccessBlockInput{Bucket: &bucketName})
		if err != nil {
			if isS3NotFoundError(err) {
				return nil
			}
			s.logger.Warn("Could not retrieve public access block", "bucket", bucketName, "error", err)
			return nil
		}
		publicAccess = mapPublicAccessBlock(out.PublicAccessBlockConfiguration)
		return nil
	})

	eg.Go(func() error {
		out, err := s.client.GetBucketLogging(egCtx, &s3.GetBucketLoggingInput{Bucket: &bucketName})
		if err != nil {
			s.logger.Warn("Could not retrieve logging", "bucket", bucketName, "error", err)
			return nil
		}
		logging = mapLogging(out.LoggingEnabled)
		return nil
	})

	eg.Go(func() error {
		out, err := s.client.GetObjectLockConfiguration(egCtx, &s3.GetObjectLockConfigurationInput{Bucket: &bucketName})
		if err != nil {
			if isS3NotFoundError(err) {
				return nil
			}
			s.logger.Warn("Could not retrieve object lock configuration", "bucket", bucketName, "error", err)
			return nil
		}
		retention = mapRetentionPolicy(out.ObjectLockConfiguration)
		return nil
	})

	if err := eg.Wait(); err != nil {
		return storage.Bucket{}, fmt.Errorf("error describing bucket: %w", err)
	}

	bucket.Location = location
	bucket.Versioning = versioning
	bucket.Encryption = encryption
	bucket.LifecycleRules = lifecycle
	bucket.Labels = labels
	bucket.IAMPolicy = iamPolicy
	bucket.ACLs = acls
	bucket.PublicAccessPrevention = publicAccess
	bucket.Logging = logging
	bucket.RetentionPolicy = retention

	return bucket, nil
}

// isS3NotFoundError checks if an error is an S3 "not configured" error
// (e.g., NoSuchTagSet, NoSuchBucketPolicy, NoSuchLifecycleConfiguration).
func isS3NotFoundError(err error) bool {
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
