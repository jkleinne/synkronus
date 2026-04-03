package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestObjectBasename(t *testing.T) {
	tests := []struct {
		name      string
		objectKey string
		want      string
		wantErr   string
	}{
		{"simple filename", "photo.jpg", "photo.jpg", ""},
		{"nested path", "dir/subdir/file.txt", "file.txt", ""},
		{"single segment no extension", "readme", "readme", ""},
		{"directory marker", "dir/", "", "cannot download directory marker"},
		{"root slash", "/", "", "cannot download directory marker"},
		{"dot only", ".", "", "cannot derive filename"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := objectBasename(tt.objectKey)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("error = %q, want containing %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("objectBasename(%q) = %q, want %q", tt.objectKey, got, tt.want)
			}
		})
	}
}

func TestResolveOutputPath(t *testing.T) {
	t.Run("existing directory appends basename", func(t *testing.T) {
		dir := t.TempDir()
		got, err := resolveOutputPath(dir, "bucket/data.csv")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := filepath.Join(dir, "data.csv")
		if got != want {
			t.Errorf("resolveOutputPath = %q, want %q", got, want)
		}
	})

	t.Run("trailing slash appends basename", func(t *testing.T) {
		got, err := resolveOutputPath("/tmp/nonexistent/", "dir/file.txt")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := filepath.Join("/tmp/nonexistent", "file.txt")
		if got != want {
			t.Errorf("resolveOutputPath = %q, want %q", got, want)
		}
	})

	t.Run("explicit file path used as-is", func(t *testing.T) {
		got, err := resolveOutputPath("/tmp/out.txt", "dir/file.txt")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "/tmp/out.txt" {
			t.Errorf("resolveOutputPath = %q, want %q", got, "/tmp/out.txt")
		}
	})

	t.Run("directory marker key propagates error", func(t *testing.T) {
		_, err := resolveOutputPath("/tmp/out.txt", "dir/")
		if err == nil {
			t.Fatal("expected error for directory marker key")
		}
		if !strings.Contains(err.Error(), "cannot download directory marker") {
			t.Errorf("error = %q, want containing 'cannot download directory marker'", err.Error())
		}
	})
}

func TestWriteToFile(t *testing.T) {
	t.Run("writes content successfully", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "output.txt")
		content := "hello world"

		err := writeToFile(path, strings.NewReader(content))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read written file: %v", err)
		}
		if string(data) != content {
			t.Errorf("file content = %q, want %q", string(data), content)
		}
	})

	t.Run("removes partial file on copy error", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "partial.txt")

		err := writeToFile(path, &failingReader{failAfter: 5})
		if err == nil {
			t.Fatal("expected error from failing reader")
		}

		if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
			t.Error("expected partial file to be removed after copy error")
		}
	})
}

// failingReader writes n bytes then returns an error.
type failingReader struct {
	failAfter int
	read      int
}

func (r *failingReader) Read(p []byte) (int, error) {
	if r.read >= r.failAfter {
		return 0, fmt.Errorf("simulated read error")
	}
	n := len(p)
	remaining := r.failAfter - r.read
	if n > remaining {
		n = remaining
	}
	for i := 0; i < n; i++ {
		p[i] = 'x'
	}
	r.read += n
	return n, nil
}
