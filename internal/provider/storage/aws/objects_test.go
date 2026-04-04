package aws

import "testing"

func TestDetectContentType(t *testing.T) {
	tests := []struct {
		name      string
		objectKey string
		want      string
	}{
		{"json file", "data/config.json", "application/json"},
		{"text file", "logs/output.txt", "text/plain; charset=utf-8"},
		{"no extension", "README", ""},
		{"unknown extension", "file.xyz123", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectContentType(tt.objectKey)
			if got != tt.want {
				t.Errorf("detectContentType(%q) = %q, want %q", tt.objectKey, got, tt.want)
			}
		})
	}
}
