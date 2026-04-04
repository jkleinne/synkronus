package main

import "testing"

func TestCopyObject_DestKeyDefault(t *testing.T) {
	tests := []struct {
		name       string
		srcKey     string
		flagDstKey string
		wantKey    string
	}{
		{"explicit dest key", "src/file.txt", "dst/file.txt", "dst/file.txt"},
		{"defaults to src key", "src/file.txt", "", "src/file.txt"},
		{"preserves nested path", "a/b/c/data.json", "", "a/b/c/data.json"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			destKey := tt.flagDstKey
			if destKey == "" {
				destKey = tt.srcKey
			}
			if destKey != tt.wantKey {
				t.Errorf("destKey = %q, want %q", destKey, tt.wantKey)
			}
		})
	}
}
