package gcp

import (
	"synkronus/internal/domain"
	"synkronus/internal/domain/storage"
	"testing"
	"time"

	gcpstorage "cloud.google.com/go/storage"
)

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
			got := detectContentType(tt.objectKey)
			if got != tt.want {
				t.Errorf("detectContentType(%q) = %q, want %q", tt.objectKey, got, tt.want)
			}
		})
	}
}

func TestMapObjectAttributes_NilAttrs(t *testing.T) {
	result := mapObjectAttributes(nil, nil)
	if result.Key != "" {
		t.Errorf("expected empty Key for nil attrs, got %q", result.Key)
	}
}

func TestMapObjectAttributes_NilEncryption_DefaultsToGoogleManaged(t *testing.T) {
	attrs := &gcpstorage.ObjectAttrs{
		Name:   "test-object",
		Bucket: "test-bucket",
	}
	result := mapObjectAttributes(attrs, nil)
	if result.Encryption == nil {
		t.Fatal("expected non-nil encryption")
	}
	if result.Encryption.KmsKeyName != "Google-managed" {
		t.Errorf("expected Google-managed, got %q", result.Encryption.KmsKeyName)
	}
	if result.Encryption.Algorithm != "AES256" {
		t.Errorf("expected AES256, got %q", result.Encryption.Algorithm)
	}
}

func TestMapObjectAttributes_WithEncryption(t *testing.T) {
	attrs := &gcpstorage.ObjectAttrs{
		Name:   "test-object",
		Bucket: "test-bucket",
	}
	enc := &storage.Encryption{
		KmsKeyName: "projects/p/locations/l/keyRings/kr/cryptoKeys/k",
		Algorithm:  "AES256",
	}
	result := mapObjectAttributes(attrs, enc)
	if result.Encryption.KmsKeyName != enc.KmsKeyName {
		t.Errorf("expected %q, got %q", enc.KmsKeyName, result.Encryption.KmsKeyName)
	}
}

func TestMapObjectAttributes_AllFieldsMapped(t *testing.T) {
	now := time.Now()
	attrs := &gcpstorage.ObjectAttrs{
		Name:               "path/to/file.txt",
		Bucket:             "my-bucket",
		Size:               1024,
		StorageClass:       "STANDARD",
		Updated:            now,
		Created:            now,
		Etag:               "abc123",
		ContentType:        "text/plain",
		ContentEncoding:    "gzip",
		ContentLanguage:    "en",
		CacheControl:       "no-cache",
		ContentDisposition: "inline",
		Generation:         12345,
		Metageneration:     2,
		Metadata:           map[string]string{"key": "value"},
	}
	result := mapObjectAttributes(attrs, nil)

	if result.Key != "path/to/file.txt" {
		t.Errorf("Key: expected %q, got %q", "path/to/file.txt", result.Key)
	}
	if result.Bucket != "my-bucket" {
		t.Errorf("Bucket: expected %q, got %q", "my-bucket", result.Bucket)
	}
	if result.Provider != domain.GCP {
		t.Errorf("Provider: expected %q, got %q", domain.GCP, result.Provider)
	}
	if result.Size != 1024 {
		t.Errorf("Size: expected 1024, got %d", result.Size)
	}
	if result.StorageClass != "STANDARD" {
		t.Errorf("StorageClass: expected STANDARD, got %q", result.StorageClass)
	}
	if result.ContentType != "text/plain" {
		t.Errorf("ContentType: expected text/plain, got %q", result.ContentType)
	}
	if result.ContentEncoding != "gzip" {
		t.Errorf("ContentEncoding: expected gzip, got %q", result.ContentEncoding)
	}
	if result.Generation != 12345 {
		t.Errorf("Generation: expected 12345, got %d", result.Generation)
	}
	if result.Metadata["key"] != "value" {
		t.Errorf("Metadata: expected key=value, got %v", result.Metadata)
	}
}
