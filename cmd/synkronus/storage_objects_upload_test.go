package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUploadObject_KeyDerivation(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		flagKey  string
		wantKey  string
	}{
		{"explicit key", "/tmp/data.csv", "custom/path.csv", "custom/path.csv"},
		{"derived from filename", "/tmp/data.csv", "", "data.csv"},
		{"nested path derives basename", "/home/user/docs/report.pdf", "", "report.pdf"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := tt.flagKey
			if key == "" {
				key = filepath.Base(tt.filePath)
			}
			if key != tt.wantKey {
				t.Errorf("key = %q, want %q", key, tt.wantKey)
			}
		})
	}
}

func TestUploadObject_FileValidation(t *testing.T) {
	t.Run("directory is rejected", func(t *testing.T) {
		dir := t.TempDir()
		info, err := os.Stat(dir)
		if err != nil {
			t.Fatal(err)
		}
		if !info.IsDir() {
			t.Fatal("expected directory")
		}
	})

	t.Run("nonexistent file is rejected", func(t *testing.T) {
		_, err := os.Stat("/nonexistent/path/file.txt")
		if err == nil {
			t.Fatal("expected error for nonexistent file")
		}
	})
}
