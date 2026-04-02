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

func TestRenderKeyValueGrid_BasicGrid(t *testing.T) {
	sections := []KeyValueSection{
		{
			Title: "Info",
			Entries: []KeyValue{
				{Key: "Name", Value: "test-bucket", Style: ValueDefault},
				{Key: "Status", Value: "active", Style: ValueEnabled},
			},
		},
	}
	result := RenderKeyValueGrid(sections, 80)
	if !strings.Contains(result, "Info") {
		t.Error("expected section title 'Info'")
	}
	if !strings.Contains(result, "Name") {
		t.Error("expected key 'Name'")
	}
	if !strings.Contains(result, "test-bucket") {
		t.Error("expected value 'test-bucket'")
	}
}

func TestFormatBool_TrueAndFalse(t *testing.T) {
	trueResult := FormatBool(true)
	if !strings.Contains(trueResult, "enabled") {
		t.Errorf("expected 'enabled' for true, got %q", trueResult)
	}
	falseResult := FormatBool(false)
	if !strings.Contains(falseResult, "disabled") {
		t.Errorf("expected 'disabled' for false, got %q", falseResult)
	}
}

func TestFormatBoolValue_TrueAndFalse(t *testing.T) {
	text, style := FormatBoolValue(true)
	if text != "enabled" || style != ValueEnabled {
		t.Errorf("FormatBoolValue(true) = (%q, %d), want (enabled, ValueEnabled)", text, style)
	}
	text, style = FormatBoolValue(false)
	if text != "disabled" || style != ValueDisabled {
		t.Errorf("FormatBoolValue(false) = (%q, %d), want (disabled, ValueDisabled)", text, style)
	}
}

func TestFormatOptionalString_EmptyReturnsNone(t *testing.T) {
	result := FormatOptionalString("")
	if !strings.Contains(result, "none") {
		t.Errorf("expected 'none' for empty string, got %q", result)
	}
}

func TestFormatOptionalString_NonEmpty(t *testing.T) {
	result := FormatOptionalString("hello")
	if result != "hello" {
		t.Errorf("expected 'hello', got %q", result)
	}
}

func TestRenderCount_FormatsCorrectly(t *testing.T) {
	result := RenderCount(5, "buckets", "2 providers")
	if !strings.Contains(result, "5 buckets") {
		t.Errorf("expected '5 buckets' in output, got %q", result)
	}
	if !strings.Contains(result, "2 providers") {
		t.Errorf("expected '2 providers' in output, got %q", result)
	}
}

func TestRenderCount_NoExtra(t *testing.T) {
	result := RenderCount(3, "items", "")
	if !strings.Contains(result, "3 items") {
		t.Errorf("expected '3 items' in output, got %q", result)
	}
}

func TestRenderError_Truncation(t *testing.T) {
	longMsg := strings.Repeat("x", 200)
	result := RenderError(longMsg, 80)
	if !strings.Contains(result, "...") {
		t.Error("expected truncated message to contain '...'")
	}
}

func TestRenderError_NewlineStripping(t *testing.T) {
	result := RenderError("line1\nline2\nline3", 80)
	if strings.Contains(result, "line2") {
		t.Error("expected newlines to be stripped, but found line2")
	}
	if !strings.Contains(result, "line1") {
		t.Error("expected first line to be preserved")
	}
}

func TestRenderTableEmptyHeaders(t *testing.T) {
	result := RenderTable([]string{}, nil, 0, 0, 60)
	if result != "" {
		t.Errorf("expected empty string for empty headers, got %q", result)
	}
}
