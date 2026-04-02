package ui

import (
	"strings"
	"testing"
)

func TestRenderConfigListWithEntries(t *testing.T) {
	entries := []ConfigEntry{
		{Key: "gcp.project", Value: "my-project"},
		{Key: "aws.region", Value: "us-east-1"},
	}
	result := RenderConfigList(entries, 0, 0, 80)
	if !strings.Contains(result, "gcp.project") {
		t.Error("config list should show key")
	}
	if !strings.Contains(result, "my-project") {
		t.Error("config list should show value")
	}
}

func TestRenderConfigListEmpty(t *testing.T) {
	result := RenderConfigList(nil, 0, 0, 80)
	if !strings.Contains(result, "No configuration") {
		t.Error("empty config should show placeholder message")
	}
}

func TestRenderConfigEdit(t *testing.T) {
	result := RenderConfigEdit("gcp.project", "new-project", false, 80, 24)
	if !strings.Contains(result, "gcp.project") {
		t.Error("edit form should show key being edited")
	}
}

func TestRenderConfigDeleteConfirm(t *testing.T) {
	result := RenderConfigDeleteConfirm("gcp.project")
	if !strings.Contains(result, "gcp") {
		t.Error("delete confirm should show provider name")
	}
	if !strings.Contains(result, "Remove") {
		t.Error("delete confirm should say Remove")
	}
}

func TestRenderConfigListShowsHeaders(t *testing.T) {
	entries := []ConfigEntry{
		{Key: "gcp.project", Value: "my-project"},
	}
	result := RenderConfigList(entries, 0, 0, 120)
	if !strings.Contains(result, "KEY") {
		t.Error("config list should show KEY header")
	}
	if !strings.Contains(result, "VALUE") {
		t.Error("config list should show VALUE header")
	}
}

func TestRenderConfigListCursorHighlight(t *testing.T) {
	entries := []ConfigEntry{
		{Key: "gcp.project", Value: "my-project"},
		{Key: "aws.region", Value: "us-east-1"},
	}
	result := RenderConfigList(entries, 0, 0, 120)
	if !strings.Contains(result, "▸") {
		t.Error("selected row should have cursor indicator")
	}
}

func TestRenderConfigEditNewEntry(t *testing.T) {
	result := RenderConfigEdit("new.key", "", true, 80, 24)
	if !strings.Contains(result, "Add new entry") {
		t.Error("new entry form should show 'Add new entry' header")
	}
}

func TestRenderConfigEditExistingEntry(t *testing.T) {
	result := RenderConfigEdit("gcp.project", "old-project", false, 80, 24)
	if !strings.Contains(result, "Edit entry") {
		t.Error("edit form for existing entry should show 'Edit entry' header")
	}
}

func TestRenderConfigDeleteConfirmMessage(t *testing.T) {
	result := RenderConfigDeleteConfirm("aws.region")
	if !strings.Contains(result, "Enter") {
		t.Error("delete confirm should mention Enter key")
	}
	if !strings.Contains(result, "Esc") {
		t.Error("delete confirm should mention Esc key")
	}
}
