package output

import (
	"strings"
	"testing"
)

func TestTable_EmptyTable(t *testing.T) {
	table := NewTable([]string{"NAME", "VALUE"})
	result := table.String()
	if !strings.Contains(result, "NAME") {
		t.Error("expected headers in output")
	}
	if strings.Count(result, "\n") != 3 {
		t.Errorf("expected 3 lines (top border, headers, bottom border), got %d", strings.Count(result, "\n"))
	}
}

func TestTable_SingleRow(t *testing.T) {
	table := NewTable([]string{"NAME", "AGE"})
	table.AddRow([]string{"Alice", "30"})
	result := table.String()
	if !strings.Contains(result, "Alice") {
		t.Error("expected row data in output")
	}
	if !strings.Contains(result, "30") {
		t.Error("expected row data in output")
	}
}

func TestTable_MultiRow_ColumnWidths(t *testing.T) {
	table := NewTable([]string{"N", "V"})
	table.AddRow([]string{"short", "x"})
	table.AddRow([]string{"a", "longer-value"})
	result := table.String()
	if !strings.Contains(result, "longer-value") {
		t.Error("expected long value in output")
	}
	lines := strings.Split(result, "\n")
	if len(lines[0]) != len(lines[2]) {
		t.Error("border lines should be equal length")
	}
}

func TestTable_RowShorterThanHeaders(t *testing.T) {
	table := NewTable([]string{"A", "B", "C"})
	table.AddRow([]string{"1"})
	result := table.String()
	if !strings.Contains(result, "1") {
		t.Error("expected partial row data in output")
	}
}

func TestFormatHeaderSection(t *testing.T) {
	result := FormatHeaderSection("Test Title")
	if !strings.Contains(result, "Test Title") {
		t.Error("expected title in header section")
	}
	if !strings.Contains(result, "=") {
		t.Error("expected border characters in header section")
	}
}

func TestFormatSectionTitle(t *testing.T) {
	result := FormatSectionTitle("My Section")
	expected := "-- My Section --"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}
