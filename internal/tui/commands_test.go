package tui

import (
	"testing"

	"synkronus/internal/tui/ui"
)

func TestMessageTypes(t *testing.T) {
	// Compile-time check that all message types exist with expected fields
	_ = BucketsLoadedMsg{Err: nil}
	_ = BucketDetailMsg{Err: nil}
	_ = ObjectsLoadedMsg{Err: nil}
	_ = ObjectDetailMsg{Err: nil}
	_ = InstancesLoadedMsg{Err: nil}
	_ = InstanceDetailMsg{Err: nil}
	_ = ConfigLoadedMsg{Err: nil}
	_ = BucketCreatedMsg{Err: nil}
	_ = BucketDeletedMsg{Err: nil}
	_ = ConfigUpdatedMsg{Err: nil}
	_ = ConfigDeletedMsg{Err: nil}
	_ = StatusClearMsg{}
}

func TestDefaultTimeout(t *testing.T) {
	if defaultTimeout.Seconds() != 30 {
		t.Errorf("expected 30s default timeout, got %v", defaultTimeout)
	}
}

func TestFlattenSettings(t *testing.T) {
	settings := map[string]interface{}{
		"gcp": map[string]interface{}{
			"project": "my-project",
		},
		"aws": map[string]interface{}{
			"region": "us-east-1",
		},
	}
	entries := flattenSettings(settings, "")
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	found := make(map[string]string)
	for _, e := range entries {
		found[e.Key] = e.Value
	}
	if found["gcp.project"] != "my-project" {
		t.Errorf("expected gcp.project=my-project, got %q", found["gcp.project"])
	}
	if found["aws.region"] != "us-east-1" {
		t.Errorf("expected aws.region=us-east-1, got %q", found["aws.region"])
	}
}

func TestFlattenSettingsEmpty(t *testing.T) {
	entries := flattenSettings(map[string]interface{}{}, "")
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

func TestFlattenSettingsNested(t *testing.T) {
	settings := map[string]interface{}{
		"a": map[string]interface{}{
			"b": map[string]interface{}{
				"c": "deep",
			},
		},
	}
	entries := flattenSettings(settings, "")
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Key != "a.b.c" || entries[0].Value != "deep" {
		t.Errorf("expected a.b.c=deep, got %s=%s", entries[0].Key, entries[0].Value)
	}
}

func TestConfigEntryType(t *testing.T) {
	// Verify ConfigEntry from ui package is usable
	entry := ui.ConfigEntry{Key: "test", Value: "value"}
	if entry.Key != "test" || entry.Value != "value" {
		t.Error("ConfigEntry should hold key-value pairs")
	}
}
