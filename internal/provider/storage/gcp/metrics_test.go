package gcp

import (
	"fmt"
	"testing"

	monitoringpb "cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
)

func TestExtractUsageValue_NilInput(t *testing.T) {
	result := extractUsageValue(nil)
	if result != 0 {
		t.Errorf("expected 0 for nil input, got %d", result)
	}
}

func TestExtractUsageValue_DoubleValue(t *testing.T) {
	tv := &monitoringpb.TypedValue{
		Value: &monitoringpb.TypedValue_DoubleValue{DoubleValue: 1234.6},
	}
	result := extractUsageValue(tv)
	if result != 1235 {
		t.Errorf("expected 1235 (rounded), got %d", result)
	}
}

func TestExtractUsageValue_Int64Value(t *testing.T) {
	tv := &monitoringpb.TypedValue{
		Value: &monitoringpb.TypedValue_Int64Value{Int64Value: 5000},
	}
	result := extractUsageValue(tv)
	if result != 5000 {
		t.Errorf("expected 5000, got %d", result)
	}
}

func TestExtractUsageValue_UnknownType(t *testing.T) {
	// TypedValue with no Value set (zero value)
	tv := &monitoringpb.TypedValue{}
	result := extractUsageValue(tv)
	if result != 0 {
		t.Errorf("expected 0 for unknown type, got %d", result)
	}
}

func TestBuildMetricsRequest_CorrectFields(t *testing.T) {
	projectID := "test-project"
	filter := `metric.type="storage.googleapis.com/storage/v2/total_bytes"`

	req := buildMetricsRequest(projectID, filter)

	expectedName := fmt.Sprintf("projects/%s", projectID)
	if req.Name != expectedName {
		t.Errorf("expected Name=%q, got %q", expectedName, req.Name)
	}
	if req.Filter != filter {
		t.Errorf("expected Filter=%q, got %q", filter, req.Filter)
	}
	if req.Interval == nil {
		t.Fatal("expected non-nil Interval")
	}
	if req.Aggregation == nil {
		t.Fatal("expected non-nil Aggregation")
	}
	if req.Aggregation.PerSeriesAligner != monitoringpb.Aggregation_ALIGN_MEAN {
		t.Errorf("expected ALIGN_MEAN, got %v", req.Aggregation.PerSeriesAligner)
	}
	if req.Aggregation.CrossSeriesReducer != monitoringpb.Aggregation_REDUCE_SUM {
		t.Errorf("expected REDUCE_SUM, got %v", req.Aggregation.CrossSeriesReducer)
	}
	if len(req.Aggregation.GroupByFields) != 1 || req.Aggregation.GroupByFields[0] != "resource.labels.bucket_name" {
		t.Errorf("expected GroupByFields=[resource.labels.bucket_name], got %v", req.Aggregation.GroupByFields)
	}
}
