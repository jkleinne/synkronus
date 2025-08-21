// File: pkg/storage/gcp/metrics.go
package gcp

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	monitoringpb "cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	"google.golang.org/api/iterator"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const metricTimeWindow = 72 * time.Hour

// ErrMetricsNotFound indicates that the usage metrics could not be found within the queried time range
// This often happens for new buckets that haven't reported metrics yet
var ErrMetricsNotFound = errors.New("usage metrics not found in the monitoring window")

func (g *GCPStorage) getAllBucketUsages(ctx context.Context) (map[string]int64, error) {
	g.logger.Debug("Fetching GCP bucket usage metrics via Monitoring API (Aggregated)")
	client, err := monitoring.NewMetricClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create monitoring client: %w", err)
	}
	defer client.Close()

	endTime := time.Now()
	startTime := endTime.Add(-metricTimeWindow)

	req := &monitoringpb.ListTimeSeriesRequest{
		Name: fmt.Sprintf("projects/%s", g.projectID),
		// Request metrics for all buckets
		Filter: `metric.type="storage.googleapis.com/storage/v2/total_bytes"`,
		Interval: &monitoringpb.TimeInterval{
			StartTime: timestamppb.New(startTime),
			EndTime:   timestamppb.New(endTime),
		},
		Aggregation: &monitoringpb.Aggregation{
			AlignmentPeriod:    durationpb.New(metricTimeWindow),
			PerSeriesAligner:   monitoringpb.Aggregation_ALIGN_MEAN,
			CrossSeriesReducer: monitoringpb.Aggregation_REDUCE_SUM,
			GroupByFields:      []string{"resource.labels.bucket_name"},
		},
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
			g.logger.Warn("Aggregated metric response missing 'bucket_name' label")
			continue
		}

		// Get the aggregated data point
		if len(resp.GetPoints()) > 0 {
			pointValue := resp.GetPoints()[0].GetValue()
			usageMap[bucketName] = extractUsageValue(pointValue)
		}
	}

	return usageMap, nil
}

func (g *GCPStorage) getSingleBucketUsage(ctx context.Context, bucketName string) (int64, error) {
	g.logger.Debug("Fetching single GCP bucket usage metric via Monitoring API (Aggregated)", "bucket", bucketName)
	client, err := monitoring.NewMetricClient(ctx)
	if err != nil {
		return -1, fmt.Errorf("failed to create monitoring client: %w", err)
	}
	defer client.Close()

	endTime := time.Now()
	startTime := endTime.Add(-metricTimeWindow)

	req := &monitoringpb.ListTimeSeriesRequest{
		Name: fmt.Sprintf("projects/%s", g.projectID),
		// Request metrics for a single bucket
		Filter: fmt.Sprintf(`metric.type="storage.googleapis.com/storage/v2/total_bytes" AND resource.labels.bucket_name="%s"`, bucketName),
		Interval: &monitoringpb.TimeInterval{
			StartTime: timestamppb.New(startTime),
			EndTime:   timestamppb.New(endTime),
		},
		Aggregation: &monitoringpb.Aggregation{
			AlignmentPeriod:    durationpb.New(metricTimeWindow),
			PerSeriesAligner:   monitoringpb.Aggregation_ALIGN_MEAN,
			CrossSeriesReducer: monitoringpb.Aggregation_REDUCE_SUM,
			GroupByFields:      []string{"resource.labels.bucket_name"},
		},
	}

	it := client.ListTimeSeries(ctx, req)

	// Since we aggregated everything into a single point and summed across series,
	// we expect exactly one time series in the response
	resp, err := it.Next()

	if err == iterator.Done {
		return -1, ErrMetricsNotFound
	}
	if err != nil {
		return -1, fmt.Errorf("error getting metric data for bucket %s: %w", bucketName, err)
	}

	// Check if the time series has the aggregated data point
	if len(resp.GetPoints()) > 0 {
		pointValue := resp.GetPoints()[0].GetValue()
		return extractUsageValue(pointValue), nil
	}

	return -1, ErrMetricsNotFound
}

func extractUsageValue(pointValue *monitoringpb.TypedValue) int64 {
	if pointValue == nil {
		return 0
	}

	switch v := pointValue.Value.(type) {
	case *monitoringpb.TypedValue_DoubleValue:
		return int64(math.Round(v.DoubleValue))
	case *monitoringpb.TypedValue_Int64Value:
		return v.Int64Value
	default:
		return 0
	}
}
