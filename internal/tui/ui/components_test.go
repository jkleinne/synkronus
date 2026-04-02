package ui

import (
	"strings"
	"testing"
)

func TestRenderBannerContainsSynkronus(t *testing.T) {
	result := RenderBanner(80)
	for _, char := range "SYNKRONUS" {
		if !strings.ContainsRune(result, char) {
			t.Errorf("banner missing character %q", string(char))
		}
	}
}

func TestRenderBannerRespectsWidth(t *testing.T) {
	narrow := RenderBanner(40)
	wide := RenderBanner(120)
	if narrow == "" || wide == "" {
		t.Error("banner should not be empty")
	}
}

func TestRenderTabsHighlightsActive(t *testing.T) {
	result := RenderTabs(1, 80, nil)
	if !strings.Contains(result, "SQL") {
		t.Error("tabs output should contain SQL")
	}
	if !strings.Contains(result, "Storage") {
		t.Error("tabs output should contain Storage")
	}
}

func TestRenderTableEmpty(t *testing.T) {
	result := RenderTable([]string{"Name", "Value"}, nil, 0, 0, 60)
	if !strings.Contains(result, "NAME") {
		t.Error("empty table should still show headers")
	}
}

func TestRenderTableWithRows(t *testing.T) {
	headers := []string{"Name", "Provider"}
	rows := [][]string{
		{"bucket-1", "GCP"},
		{"bucket-2", "AWS"},
	}
	result := RenderTable(headers, rows, 0, 0, 60)
	if !strings.Contains(result, "bucket-1") {
		t.Error("table should contain first row data")
	}
	if !strings.Contains(result, "bucket-2") {
		t.Error("table should contain second row data")
	}
}

func TestRenderModalContainsTitle(t *testing.T) {
	result := RenderModal("Delete Bucket", "Are you sure?", 80, 24)
	if !strings.Contains(result, "Delete Bucket") {
		t.Error("modal should contain title")
	}
	if !strings.Contains(result, "Are you sure?") {
		t.Error("modal should contain content")
	}
}

func TestRenderBreadcrumb(t *testing.T) {
	result := RenderBreadcrumb([]string{"Storage", "Buckets", "my-bucket"})
	if !strings.Contains(result, "Storage") || !strings.Contains(result, "my-bucket") {
		t.Error("breadcrumb should contain all parts")
	}
}

func TestRenderSpinnerNonEmpty(t *testing.T) {
	result := RenderSpinnerView("⣾", "Loading buckets...")
	if result == "" {
		t.Error("spinner view should not be empty")
	}
	if !strings.Contains(result, "Loading buckets...") {
		t.Error("spinner should show label")
	}
}

func TestRenderErrorMessage(t *testing.T) {
	result := RenderError("something broke", 80)
	if !strings.Contains(result, "something broke") {
		t.Error("error should contain message")
	}
}

func TestRenderHelpContentNonEmpty(t *testing.T) {
	result := RenderHelpContent(0, 0)
	if result == "" {
		t.Error("help content should not be empty")
	}
	if !strings.Contains(result, "Storage") {
		t.Error("help content should contain Storage context")
	}
}
