// File: pkg/storage/gcp/gcp.go
package gcp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"synkronus/pkg/common"
	"time"

	"synkronus/pkg/storage"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	monitoringpb "cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	gcpstorage "cloud.google.com/go/storage"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iterator"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GCPStorage struct {
	client    *gcpstorage.Client
	projectID string
	logger    *slog.Logger
}

var _ storage.Storage = (*GCPStorage)(nil)

func NewGCPStorage(ctx context.Context, projectID string, logger *slog.Logger) (*GCPStorage, error) {
	client, err := gcpstorage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCP storage client: %w", err)
	}

	return &GCPStorage{
		client:    client,
		projectID: projectID,
		logger:    logger,
	}, nil
}

func (g *GCPStorage) ProviderName() common.Provider {
	return common.GCP
}

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

// Fetches storage metrics for all buckets in the project in one request
func (g *GCPStorage) getAllBucketUsages(ctx context.Context) (map[string]int64, error) {
	g.logger.Debug("Fetching GCP bucket usage metrics via Monitoring API")
	client, err := monitoring.NewMetricClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create monitoring client: %w", err)
	}
	defer client.Close()

	endTime := time.Now()
	startTime := endTime.Add(-15 * time.Minute)

	req := &monitoringpb.ListTimeSeriesRequest{
		Name: fmt.Sprintf("projects/%s", g.projectID),
		// Request metrics for all buckets
		Filter: `metric.type="storage.googleapis.com/storage/total_bytes"`,
		Interval: &monitoringpb.TimeInterval{
			StartTime: timestamppb.New(startTime),
			EndTime:   timestamppb.New(endTime),
		},
		View: monitoringpb.ListTimeSeriesRequest_FULL,
	}

	usageMap := make(map[string]int64)
	it := client.ListTimeSeries(ctx, req)

	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error getting metric data: %w", err)
		}

		bucketName, ok := resp.GetResource().GetLabels()["bucket_name"]
		if !ok {
			continue
		}

		// Get the latest data point
		if len(resp.GetPoints()) > 0 {
			latestPoint := resp.GetPoints()[0]
			// Ensure we only store the latest point if multiple series somehow match
			if _, exists := usageMap[bucketName]; !exists {
				usageMap[bucketName] = latestPoint.GetValue().GetInt64Value()
			}
		}
	}

	return usageMap, nil
}

func (g *GCPStorage) getSingleBucketUsage(ctx context.Context, bucketName string) (int64, error) {
	g.logger.Debug("Fetching single GCP bucket usage metric via Monitoring API", "bucket", bucketName)
	client, err := monitoring.NewMetricClient(ctx)
	if err != nil {
		return -1, fmt.Errorf("failed to create monitoring client: %w", err)
	}
	defer client.Close()

	endTime := time.Now()
	startTime := endTime.Add(-15 * time.Minute)

	req := &monitoringpb.ListTimeSeriesRequest{
		Name: fmt.Sprintf("projects/%s", g.projectID),
		// Request metrics for a single bucket
		Filter: fmt.Sprintf(`metric.type="storage.googleapis.com/storage/total_bytes" AND resource.labels.bucket_name="%s"`, bucketName),
		Interval: &monitoringpb.TimeInterval{
			StartTime: timestamppb.New(startTime),
			EndTime:   timestamppb.New(endTime),
		},
		View: monitoringpb.ListTimeSeriesRequest_FULL,
	}

	it := client.ListTimeSeries(ctx, req)
	resp, err := it.Next()

	if err == iterator.Done {
		return -1, fmt.Errorf("no usage metrics found for bucket %s", bucketName)
	}
	if err != nil {
		return -1, fmt.Errorf("error getting metric data for bucket %s: %w", bucketName, err)
	}

	if len(resp.GetPoints()) > 0 {
		return resp.GetPoints()[0].GetValue().GetInt64Value(), nil
	}

	return -1, fmt.Errorf("metric response for bucket %s contained no data points", bucketName)
}

func (g *GCPStorage) DescribeBucket(ctx context.Context, bucketName string) (storage.Bucket, error) {
	g.logger.Debug("Starting GCP DescribeBucket operation", "bucket", bucketName)

	bucketHandle := g.client.Bucket(bucketName)
	attrs, err := bucketHandle.Attrs(ctx)
	if err != nil {
		return storage.Bucket{}, fmt.Errorf("error getting bucket attributes: %w", err)
	}

	// Fetch usage metrics for only this bucket
	usage, err := g.getSingleBucketUsage(ctx, bucketName)
	if err != nil {
		g.logger.Warn("Failed to retrieve usage metrics, usage will be reported as N/A", "bucket", bucketName, "error", err)
		usage = -1 // Set usage to unknown on failure
	}

	// Fetch ACLs (separate API call)
	aclRules, err := g.getACLs(ctx, bucketHandle)
	if err != nil {
		// Log a warning but don't fail the entire operation, as ACLs might not be readable.
		g.logger.Warn("Could not retrieve ACLs for bucket", "bucket", bucketName, "error", err)
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

// getACLs fetches the bucket's ACLs.
func (g *GCPStorage) getACLs(ctx context.Context, bucketHandle *gcpstorage.BucketHandle) ([]storage.ACLRule, error) {
	gcpAcls, err := bucketHandle.ACL().List(ctx)
	if err != nil {
		var gcsErr *googleapi.Error
		// If Uniform Bucket-Level Access is enabled, this call fails with a 400.
		// We can check for this specific error and return an empty list.
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

func mapLifecycleRules(rules []gcpstorage.LifecycleRule) []storage.LifecycleRule {
	if len(rules) == 0 {
		return nil
	}
	var result []storage.LifecycleRule
	for _, r := range rules {
		var actionStr string
		// Refine action string for better readability
		if r.Action.StorageClass != "" {
			actionStr = fmt.Sprintf("%s to %s", r.Action.Type, r.Action.StorageClass)
		} else {
			actionStr = r.Action.Type
		}

		result = append(result, storage.LifecycleRule{
			Action: actionStr,
			Condition: storage.LifecycleCondition{
				Age:                 int(r.Condition.AgeInDays),
				CreatedBefore:       r.Condition.CreatedBefore,
				MatchesStorageClass: r.Condition.MatchesStorageClasses,
				NumNewerVersions:    int(r.Condition.NumNewerVersions),
			},
		})
	}
	return result
}

func mapLogging(l *gcpstorage.BucketLogging) *storage.Logging {
	if l == nil {
		return nil
	}
	return &storage.Logging{
		LogBucket:       l.LogBucket,
		LogObjectPrefix: l.LogObjectPrefix,
	}
}

func mapSoftDeletePolicy(sdp *gcpstorage.SoftDeletePolicy) *storage.SoftDeletePolicy {
	if sdp == nil {
		return nil
	}
	return &storage.SoftDeletePolicy{
		RetentionDuration: sdp.RetentionDuration,
	}
}

func mapPublicAccessPrevention(pap gcpstorage.PublicAccessPrevention) string {
	switch pap {
	case gcpstorage.PublicAccessPreventionEnforced:
		return "Enforced"
	case gcpstorage.PublicAccessPreventionInherited:
		return "Inherited"
	default:
		return "Unknown"
	}
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

func (g *GCPStorage) Close() error {
	if g.client != nil {
		return g.client.Close()
	}
	return nil
}
