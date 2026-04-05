package shared

import "testing"

func TestDetectContentType(t *testing.T) {
	tests := []struct {
		name      string
		objectKey string
		want      string
	}{
		{"json file", "data/config.json", "application/json"},
		{"text file", "logs/output.txt", "text/plain; charset=utf-8"},
		{"csv file", "exports/data.csv", "text/csv; charset=utf-8"},
		{"png image", "images/logo.png", "image/png"},
		{"no extension", "README", ""},
		{"unknown extension", "file.xyz123", ""},
		{"nested path with extension", "a/b/c/report.pdf", "application/pdf"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectContentType(tt.objectKey)
			if got != tt.want {
				t.Errorf("DetectContentType(%q) = %q, want %q", tt.objectKey, got, tt.want)
			}
		})
	}
}
