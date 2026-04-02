package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

type testData struct {
	Name  string `json:"name" yaml:"name"`
	Value int    `json:"value" yaml:"value"`
}

type testTableData struct {
	items []testData
}

func (t testTableData) RenderTable() string {
	table := NewTable([]string{"NAME", "VALUE"})
	for _, item := range t.items {
		table.AddRow([]string{item.Name, "42"})
	}
	return table.String()
}

func TestRender_JSON(t *testing.T) {
	data := testData{Name: "test", Value: 42}
	var buf bytes.Buffer
	err := Render(&buf, FormatJSON, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var decoded testData
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if decoded.Name != "test" || decoded.Value != 42 {
		t.Errorf("JSON round-trip failed: got %+v", decoded)
	}
}

func TestRender_YAML(t *testing.T) {
	data := testData{Name: "test", Value: 42}
	var buf bytes.Buffer
	err := Render(&buf, FormatYAML, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var decoded testData
	if err := yaml.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid YAML: %v", err)
	}
	if decoded.Name != "test" || decoded.Value != 42 {
		t.Errorf("YAML round-trip failed: got %+v", decoded)
	}
}

func TestRender_Table(t *testing.T) {
	data := testTableData{items: []testData{{Name: "alice", Value: 1}}}
	var buf bytes.Buffer
	err := Render(&buf, FormatTable, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "alice") {
		t.Error("expected table output to contain data")
	}
}

func TestRender_Table_NonTableRenderer(t *testing.T) {
	data := testData{Name: "test", Value: 42}
	var buf bytes.Buffer
	err := Render(&buf, FormatTable, data)
	if err == nil {
		t.Fatal("expected error for non-TableRenderer with table format")
	}
}

func TestRender_UnknownFormat(t *testing.T) {
	var buf bytes.Buffer
	err := Render(&buf, Format("xml"), testData{})
	if err == nil {
		t.Fatal("expected error for unknown format")
	}
}
