// File: pkg/storage/gcp/buckets.go
package gcp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"synkronus/pkg/common"
	"synkronus/pkg/storage"

	gcpstorage "cloud.google.com/go/storage"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iterator"
)

func (g *GCPStorage) ListBuckets(ctx context.Context) ([]storage.Bucket, error) {
	g.logger.Debug("Starting GCP ListBuckets operation")
	var buckets []storage.Bucket

	// 1. Fetch usage metrics for all buckets first (O(1) API calls)
	usageMap, err := g.getAllBucketUsages(ctx)
	if err != nil {
		// Propagate the error if metrics cannot be retrieved. The caller (StorageService) will handle this
		return nil, fmt.Errorf("failed to retrieve GCP bucket usage metrics: %w", err)
	}

	// 2. Fetch bucket metadata (O(N) API calls, paginated by SDK)
	it := g.client.Buckets(ctx, g.projectID)
	for {
		bucketAttrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error listing buckets metadata: %w", err)
		}

		// Default to -1 if a specific bucket wasn't found in the metrics response, although we expect it if usageMap is populated
		usage := int64(-1)
		if u, ok := usageMap[bucketAttrs.Name]; ok {
			usage = u
		}

		buckets = append(buckets, storage.Bucket{
			Name:         bucketAttrs.Name,
			Provider:     common.GCP,
			Location:     bucketAttrs.Location,
			StorageClass: bucketAttrs.StorageClass,
			CreatedAt:    bucketAttrs.Created,
			UpdatedAt:    bucketAttrs.Updated,
			UsageBytes:   usage,
			Labels:       bucketAttrs.Labels,
		})
	}

	return buckets, nil
}

func (g *GCPStorage) DescribeBucket(ctx context.Context, bucketName string) (storage.Bucket, error) {
	g.logger.Debug("Starting GCP DescribeBucket operation", "bucket", bucketName)

	bucketHandle := g.client.Bucket(bucketName)
	attrs, err := bucketHandle.Attrs(ctx)
	if err != nil {
		return storage.Bucket{}, fmt.Errorf("error getting bucket attributes: %w", err)
	}

	usage, err := g.getSingleBucketUsage(ctx, bucketName)
	if err != nil {
		logLevel := slog.LevelWarn
		logMsg := "Failed to retrieve usage metrics due to API error, usage will be reported as N/A"

		if errors.Is(err, ErrMetricsNotFound) {
			logLevel = slog.LevelInfo
			logMsg = "Usage metrics not yet available (bucket may be new), usage will be reported as N/A"
		}

		g.logger.Log(ctx, logLevel, logMsg, "bucket", bucketName, "error", err)
		usage = -1 // Set usage to unknown on failure
	}

	// Fetch ACLs (separate API call)
	aclRules, err := g.getACLs(ctx, bucketHandle)
	if err != nil {
		// Log a warning but don't fail the entire operation, as ACLs might not be readable
		g.logger.Warn("Could not retrieve ACLs for bucket", "bucket", bucketName, "error", err)
	}

	// Fetch IAM Policy (separate API call)
	iamPolicy, err := g.getIAMPolicy(ctx, bucketHandle)
	if err != nil {
		g.logger.Warn("Could not retrieve IAM policy for bucket. Requires 'storage.buckets.getIamPolicy' permission.", "bucket", bucketName, "error", err)
	}

	details := storage.Bucket{
		Name:                     attrs.Name,
		Provider:                 common.GCP,
		Location:                 attrs.Location,
		StorageClass:             attrs.StorageClass,
		CreatedAt:                attrs.Created,
		UpdatedAt:                attrs.Updated,
		UsageBytes:               usage,
		Labels:                   attrs.Labels,
		IAMPolicy:                iamPolicy,
		ACLs:                     aclRules,
		LifecycleRules:           mapLifecycleRules(attrs.Lifecycle.Rules),
		Logging:                  mapLogging(attrs.Logging),
		Versioning:               &storage.Versioning{Enabled: attrs.VersioningEnabled},
		SoftDeletePolicy:         mapSoftDeletePolicy(attrs.SoftDeletePolicy),
		UniformBucketLevelAccess: &storage.UniformBucketLevelAccess{Enabled: attrs.UniformBucketLevelAccess.Enabled},
		PublicAccessPrevention:   mapPublicAccessPrevention(attrs.PublicAccessPrevention),
	}

	return details, nil
}

// Fetches the bucket's IAM policy and maps it to the domain model
func (g *GCPStorage) getIAMPolicy(ctx context.Context, bucketHandle *gcpstorage.BucketHandle) (*storage.IAMPolicy, error) {
	policy, err := bucketHandle.IAM().Policy(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get IAM policy: %w", err)
	}

	// Map the SDK policy structure (which handles V1/V3) to the domain model
	var bindings []storage.IAMBinding
	hasConditions := false

	// Iterate over the Bindings field directly for the most accurate representation
	for _, binding := range policy.Bindings {
		// Skip bindings with conditions as they complicate the CLI view and are not yet supported in the detailed model
		// TODO: add support for conditional bindings in the future
		if binding.Condition != nil {
			g.logger.Debug("Skipping conditional IAM binding", "role", binding.Role, "condition", binding.Condition.Title)
			hasConditions = true
			continue
		}

		// Ensure principals are sorted for deterministic output
		principals := make([]string, len(binding.Members))
		copy(principals, binding.Members)
		sort.Strings(principals)

		bindings = append(bindings, storage.IAMBinding{
			Role:       binding.Role,
			Principals: principals,
		})
	}

	// Sort bindings by role name for deterministic output
	sort.Slice(bindings, func(i, j int) bool {
		return bindings[i].Role < bindings[j].Role
	})

	return &storage.IAMPolicy{
		Bindings:      bindings,
		HasConditions: hasConditions,
	}, nil
}

// getACLs fetches the bucket's ACLs
func (g *GCPStorage) getACLs(ctx context.Context, bucketHandle *gcpstorage.BucketHandle) ([]storage.ACLRule, error) {
	gcpAcls, err := bucketHandle.ACL().List(ctx)
	if err != nil {
		var gcsErr *googleapi.Error
		// If UBLA is enabled, GCP returns a 400 error when trying to list ACLs (treating as expected behavior)
		if errors.As(err, &gcsErr) && gcsErr.Code == 400 {
			return []storage.ACLRule{}, nil
		}
		return nil, fmt.Errorf("failed to list ACLs: %w", err)
	}

	var acls []storage.ACLRule
	for _, acl := range gcpAcls {
		acls = append(acls, storage.ACLRule{
			Entity: string(acl.Entity),
			Role:   string(acl.Role),
		})
	}
	return acls, nil
}

func (g *GCPStorage) CreateBucket(ctx context.Context, bucketName string, location string) error {
	bucket := g.client.Bucket(bucketName)
	attrs := &gcpstorage.BucketAttrs{
		Location: location,
	}
	if err := bucket.Create(ctx, g.projectID, attrs); err != nil {
		return fmt.Errorf("failed to create bucket: %w", err)
	}
	return nil
}

func (g *GCPStorage) DeleteBucket(ctx context.Context, bucketName string) error {
	bucket := g.client.Bucket(bucketName)
	if err := bucket.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete bucket: %w", err)
	}
	return nil
}
