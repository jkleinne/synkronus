// File: internal/provider/storage/gcp/buckets.go
package gcp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"synkronus/internal/domain"
	"synkronus/internal/domain/storage"

	gcpstorage "cloud.google.com/go/storage"
	"golang.org/x/sync/errgroup"
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
		if errors.Is(err, iterator.Done) {
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
			Provider:     domain.GCP,
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

	// Fetch supplementary data concurrently — each is best-effort.
	var (
		usage     int64 = -1
		aclRules  []storage.ACLRule
		iamPolicy *storage.IAMPolicy
	)

	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		u, err := g.getSingleBucketUsage(egCtx, bucketName)
		if err != nil {
			logLevel := slog.LevelWarn
			logMsg := "Failed to retrieve usage metrics due to API error, usage will be reported as N/A"

			if errors.Is(err, ErrMetricsNotFound) {
				logLevel = slog.LevelInfo
				logMsg = "Usage metrics not yet available (bucket may be new), usage will be reported as N/A"
			}

			g.logger.Log(egCtx, logLevel, logMsg, "bucket", bucketName, "error", err)
			return nil
		}
		usage = u
		return nil
	})

	eg.Go(func() error {
		acls, err := g.getACLs(egCtx, bucketHandle)
		if err != nil {
			g.logger.Warn("Could not retrieve ACLs for bucket", "bucket", bucketName, "error", err)
			return nil
		}
		aclRules = acls
		return nil
	})

	eg.Go(func() error {
		iam, err := g.getIAMPolicy(egCtx, bucketHandle)
		if err != nil {
			g.logger.Warn("Could not retrieve IAM policy for bucket. Requires 'storage.buckets.getIamPolicy' permission.", "bucket", bucketName, "error", err)
			return nil
		}
		iamPolicy = iam
		return nil
	})

	eg.Wait()

	details := storage.Bucket{
		Name:                     attrs.Name,
		Provider:                 domain.GCP,
		Location:                 attrs.Location,
		LocationType:             attrs.LocationType,
		StorageClass:             attrs.StorageClass,
		CreatedAt:                attrs.Created,
		UpdatedAt:                attrs.Updated,
		UsageBytes:               usage,
		RequesterPays:            attrs.RequesterPays,
		Labels:                   attrs.Labels,
		Autoclass:                &storage.Autoclass{Enabled: attrs.Autoclass.Enabled},
		IAMPolicy:                iamPolicy,
		ACLs:                     aclRules,
		LifecycleRules:           mapLifecycleRules(attrs.Lifecycle.Rules),
		Logging:                  mapLogging(attrs.Logging),
		Versioning:               &storage.Versioning{Enabled: attrs.VersioningEnabled},
		SoftDeletePolicy:         mapSoftDeletePolicy(attrs.SoftDeletePolicy),
		UniformBucketLevelAccess: &storage.UniformBucketLevelAccess{Enabled: attrs.UniformBucketLevelAccess.Enabled},
		PublicAccessPrevention:   mapPublicAccessPrevention(attrs.PublicAccessPrevention),
		Encryption:               mapBucketEncryption(attrs.Encryption),
		RetentionPolicy:          mapRetentionPolicy(attrs.RetentionPolicy),
	}

	return details, nil
}

// Fetches the bucket's IAM policy (using V3) and maps it to the domain model
func (g *GCPStorage) getIAMPolicy(ctx context.Context, bucketHandle *gcpstorage.BucketHandle) (*storage.IAMPolicy, error) {
	policy, err := bucketHandle.IAM().V3().Policy(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get V3 IAM policy: %w", err)
	}

	var bindings []storage.IAMBinding

	for _, binding := range policy.Bindings {
		principals := slices.Clone(binding.Members)
		slices.Sort(principals)

		b := storage.IAMBinding{
			Role:       binding.Role,
			Principals: principals,
		}

		if binding.Condition != nil {
			b.Condition = &storage.IAMCondition{
				Title:       binding.Condition.Title,
				Description: binding.Condition.Description,
				Expression:  binding.Condition.Expression,
			}
		}

		bindings = append(bindings, b)
	}

	slices.SortFunc(bindings, func(a, b storage.IAMBinding) int {
		if cmp := strings.Compare(a.Role, b.Role); cmp != 0 {
			return cmp
		}
		titleA, titleB := "", ""
		if a.Condition != nil {
			titleA = a.Condition.Title
		}
		if b.Condition != nil {
			titleB = b.Condition.Title
		}
		return strings.Compare(titleA, titleB)
	})

	return &storage.IAMPolicy{
		Bindings: bindings,
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

func (g *GCPStorage) CreateBucket(ctx context.Context, opts storage.CreateBucketOptions) (storage.CreateBucketResult, error) {
	bucket := g.client.Bucket(opts.Name)
	attrs := &gcpstorage.BucketAttrs{
		Location: opts.Location,
	}
	if opts.StorageClass != "" {
		attrs.StorageClass = opts.StorageClass
	}
	if opts.Labels != nil {
		attrs.Labels = opts.Labels
	}
	if opts.Versioning != nil {
		attrs.VersioningEnabled = *opts.Versioning
	}
	if opts.UniformBucketLevelAccess != nil {
		attrs.UniformBucketLevelAccess = gcpstorage.UniformBucketLevelAccess{
			Enabled: *opts.UniformBucketLevelAccess,
		}
	}
	if opts.PublicAccessPrevention != nil {
		switch *opts.PublicAccessPrevention {
		case storage.PublicAccessPreventionEnforced:
			attrs.PublicAccessPrevention = gcpstorage.PublicAccessPreventionEnforced
		case storage.PublicAccessPreventionInherited:
			attrs.PublicAccessPrevention = gcpstorage.PublicAccessPreventionInherited
		}
	}
	if err := bucket.Create(ctx, g.projectID, attrs); err != nil {
		return storage.CreateBucketResult{}, fmt.Errorf("failed to create bucket: %w", err)
	}
	return storage.CreateBucketResult{}, nil
}

func (g *GCPStorage) DeleteBucket(ctx context.Context, bucketName string) error {
	bucket := g.client.Bucket(bucketName)
	if err := bucket.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete bucket: %w", err)
	}
	return nil
}
