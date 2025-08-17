package gcp

import (
	"context"
	"fmt"
	"log"
	"synkronus/pkg/common"
	"time"

	"synkronus/pkg/storage"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	monitoringpb "cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	gcpstorage "cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GCPStorage struct {
	client    *gcpstorage.Client
	projectID string
}

var _ storage.Storage = (*GCPStorage)(nil)

func NewGCPStorage(ctx context.Context, projectID string) (*GCPStorage, error) {
	client, err := gcpstorage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCP storage client: %w", err)
	}

	return &GCPStorage{
		client:    client,
		projectID: projectID,
	}, nil
}

func (g *GCPStorage) ProviderName() common.Provider {
	return common.GCP
}

func (g *GCPStorage) ListBuckets(ctx context.Context) ([]storage.Bucket, error) {
	var buckets []storage.Bucket

	// 1. Fetch usage metrics for all buckets first (O(1) API calls)
	usageMap, err := g.getAllBucketUsages(ctx)
	if err != nil {
		log.Printf("Warning: Failed to retrieve GCP bucket usage metrics: %v\n", err)
	}

	// 2. Fetch bucket metadata (O(N) API calls, paginated by SDK)
	// Use the provided context for the listing operation
	it := g.client.Buckets(ctx, g.projectID)
	for {
		bucketAttrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error listing buckets: %w", err)
		}

		usage := int64(-1) // Default to unknown
		if usageMap != nil {
			if u, ok := usageMap[bucketAttrs.Name]; ok {
				usage = u
			}
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

func (g *GCPStorage) DescribeBucket(ctx context.Context, bucketName string) (storage.Bucket, error) {
	bucketHandle := g.client.Bucket(bucketName)
	attrs, err := bucketHandle.Attrs(ctx)
	if err != nil {
		return storage.Bucket{}, fmt.Errorf("error getting bucket attributes: %w", err)
	}

	usageMap, err := g.getAllBucketUsages(ctx)

	usage := int64(-1)
	if err != nil {
		log.Printf("Warning: Failed to retrieve usage for bucket %s: %v\n", bucketName, err)
	} else if u, ok := usageMap[bucketName]; ok {
		usage = u
	}

	details := storage.Bucket{
		Name:         attrs.Name,
		Provider:     common.GCP,
		Location:     attrs.Location,
		StorageClass: attrs.StorageClass,
		CreatedAt:    attrs.Created,
		UpdatedAt:    attrs.Updated,
		UsageBytes:   usage,
		Labels:       attrs.Labels,
	}

	return details, nil
}

func (g *GCPStorage) Close() error {
	if g.client != nil {
		return g.client.Close()
	}
	return nil
}
