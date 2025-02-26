package gcp

import (
	"context"
	"fmt"
	"time"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	monitoringpb "cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GCPStorage struct {
	client     *storage.Client
	projectID  string
	bucketName string
}

func NewGCPStorage(projectID, bucketName string) (*GCPStorage, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCP storage client: %w", err)
	}

	return &GCPStorage{
		client:     client,
		projectID:  projectID,
		bucketName: bucketName,
	}, nil
}
func (g *GCPStorage) List() ([]string, error) {
	ctx := context.Background()
	var buckets []string

	it := g.client.Buckets(ctx, g.projectID)
	for {
		bucketAttrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error listing buckets: %w", err)
		}
		buckets = append(buckets, bucketAttrs.Name)
	}

	return buckets, nil
}

func (g *GCPStorage) getBucketUsage(ctx context.Context, bucketName string) (int64, error) {
	client, err := monitoring.NewMetricClient(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to create monitoring client: %w", err)
	}
	defer client.Close()

	endTime := time.Now()
	startTime := endTime.Add(-5 * time.Minute)

	req := &monitoringpb.ListTimeSeriesRequest{
		Name:   fmt.Sprintf("projects/%s", g.projectID),
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
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("error getting metric data: %w", err)
	}

	if len(resp.GetPoints()) == 0 {
		return 0, nil
	}

	latestPoint := resp.GetPoints()[0]
	value := latestPoint.GetValue().GetInt64Value()

	return value, nil
}

func formatBytes(bytes int64) string {
	if bytes == 0 {
		return "0 B"
	}

	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	sizes := []string{"KB", "MB", "GB", "TB"}
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), sizes[exp])
}

// DescribeBucket returns details about a specific bucket
func (g *GCPStorage) DescribeBucket(bucketName string) (map[string]interface{}, error) {
	ctx := context.Background()

	bucket := g.client.Bucket(bucketName)
	attrs, err := bucket.Attrs(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting bucket attributes: %w", err)
	}

	usage, err := g.getBucketUsage(ctx, bucketName)

	var usageFormatted string

	if err != nil {
		usageFormatted = "N/A"
	} else {
		usageFormatted = formatBytes(usage)
	}

	details := map[string]interface{}{
		"name":              attrs.Name,
		"Location / Region": attrs.Location,
		"storageClass":      attrs.StorageClass,
		"created":           attrs.Created,
		"updated":           attrs.Updated,
		"Provider":          "Google Cloud",
		"Usage":             usageFormatted,
	}

	return details, nil
}

// Close closes the GCP storage client
func (g *GCPStorage) Close() error {
	if g.client != nil {
		return g.client.Close()
	}
	return nil
}
