package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"synkronus/internal/config"
	"synkronus/internal/provider/storage/shared"
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
	flat := config.FlattenSettings(settings)
	if len(flat) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(flat))
	}
	if flat["gcp.project"] != "my-project" {
		t.Errorf("expected gcp.project=my-project, got %q", flat["gcp.project"])
	}
	if flat["aws.region"] != "us-east-1" {
		t.Errorf("expected aws.region=us-east-1, got %q", flat["aws.region"])
	}
}

func TestFlattenSettingsEmpty(t *testing.T) {
	flat := config.FlattenSettings(map[string]interface{}{})
	if len(flat) != 0 {
		t.Errorf("expected 0 entries, got %d", len(flat))
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
	flat := config.FlattenSettings(settings)
	if len(flat) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(flat))
	}
	if flat["a.b.c"] != "deep" {
		t.Errorf("expected a.b.c=deep, got %s", flat["a.b.c"])
	}
}

func TestConfigEntryType(t *testing.T) {
	// Verify ConfigEntry from ui package is usable
	entry := ui.ConfigEntry{Key: "test", Value: "value"}
	if entry.Key != "test" || entry.Value != "value" {
		t.Error("ConfigEntry should hold key-value pairs")
	}
}

func TestObjectBasename_ValidKeys(t *testing.T) {
	tests := []struct {
		name string
		key  string
		want string
	}{
		{"simple file", "photo.jpg", "photo.jpg"},
		{"nested path", "dir/subdir/file.txt", "file.txt"},
		{"no extension", "readme", "readme"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := shared.ObjectBasename(tt.key)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("shared.ObjectBasename(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

func TestObjectBasename_DirectoryMarker(t *testing.T) {
	_, err := shared.ObjectBasename("dir/")
	if err == nil {
		t.Fatal("expected error for directory marker")
	}
	if !strings.Contains(err.Error(), "cannot download directory marker") {
		t.Errorf("error = %q, want containing 'cannot download directory marker'", err.Error())
	}
}

func TestObjectBasename_Degenerate(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{"dot only", "."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := shared.ObjectBasename(tt.key)
			if err == nil {
				t.Fatalf("expected error for degenerate key %q", tt.key)
			}
			if !strings.Contains(err.Error(), "cannot derive filename") {
				t.Errorf("error = %q, want containing 'cannot derive filename'", err.Error())
			}
		})
	}
}

func TestExpandTilde(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("cannot get home dir: %v", err)
	}

	tests := []struct {
		name string
		path string
		want string
	}{
		{"bare tilde", "~", home},
		{"tilde with subdir", "~/Downloads", filepath.Join(home, "Downloads")},
		{"tilde nested", "~/a/b/c", filepath.Join(home, "a/b/c")},
		{"relative path unchanged", "./", "./"},
		{"absolute path unchanged", "/tmp/data", "/tmp/data"},
		{"tilde in middle unchanged", "/foo/~/bar", "/foo/~/bar"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := expandTilde(tt.path)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("expandTilde(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}
