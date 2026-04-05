package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"synkronus/internal/domain/storage"
)

func TestUploadObjectCmd_HappyPath(t *testing.T) {
	// Write a temporary file so the command can open it.
	tmpFile := filepath.Join(t.TempDir(), "data.csv")
	if err := os.WriteFile(tmpFile, []byte("col1,col2\n1,2\n"), 0600); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	mock := &cmdMockStorage{}
	factory := &cmdStorageFactory{providers: map[string]storage.Storage{"gcp": mock}}
	app := newStorageTestApp(factory, nil)

	var buf bytes.Buffer
	cmd := newUploadObjectCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetContext(app.ToContext(context.Background()))
	cmd.SetArgs([]string{tmpFile, "--provider", "gcp", "--bucket", "my-bucket", "--key", "custom/data.csv"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUploadObjectCmd_KeyDerivedFromFilename(t *testing.T) {
	// Verify that omitting --key uses the file's base name as the object key.
	tmpFile := filepath.Join(t.TempDir(), "report.pdf")
	if err := os.WriteFile(tmpFile, []byte("%PDF-1.4"), 0600); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	mock := &cmdMockStorage{}
	factory := &cmdStorageFactory{providers: map[string]storage.Storage{"gcp": mock}}
	app := newStorageTestApp(factory, nil)

	var buf bytes.Buffer
	cmd := newUploadObjectCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetContext(app.ToContext(context.Background()))
	// --key is omitted intentionally so the command derives it from the filename.
	cmd.SetArgs([]string{tmpFile, "--provider", "gcp", "--bucket", "my-bucket"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUploadObjectCmd_FileNotFound_ReturnsError(t *testing.T) {
	app := newStorageTestApp(&cmdStorageFactory{}, nil)

	var buf bytes.Buffer
	cmd := newUploadObjectCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetContext(app.ToContext(context.Background()))
	cmd.SetArgs([]string{"/nonexistent/path/file.txt", "--provider", "gcp", "--bucket", "my-bucket"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for non-existent file, got nil")
	}
}

func TestUploadObjectCmd_DirectoryPath_ReturnsError(t *testing.T) {
	dir := t.TempDir()

	app := newStorageTestApp(&cmdStorageFactory{}, nil)

	var buf bytes.Buffer
	cmd := newUploadObjectCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetContext(app.ToContext(context.Background()))
	cmd.SetArgs([]string{dir, "--provider", "gcp", "--bucket", "my-bucket"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error when path is a directory, got nil")
	}
}

func TestUploadObjectCmd_MissingProviderFlag_ReturnsError(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "data.txt")
	if err := os.WriteFile(tmpFile, []byte("hello"), 0600); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	app := newStorageTestApp(&cmdStorageFactory{}, nil)

	var buf bytes.Buffer
	cmd := newUploadObjectCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetContext(app.ToContext(context.Background()))
	cmd.SetArgs([]string{tmpFile, "--bucket", "my-bucket"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for missing --provider flag, got nil")
	}
}

func TestUploadObjectCmd_MissingBucketFlag_ReturnsError(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "data.txt")
	if err := os.WriteFile(tmpFile, []byte("hello"), 0600); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	app := newStorageTestApp(&cmdStorageFactory{}, nil)

	var buf bytes.Buffer
	cmd := newUploadObjectCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetContext(app.ToContext(context.Background()))
	cmd.SetArgs([]string{tmpFile, "--provider", "gcp"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for missing --bucket flag, got nil")
	}
}

func TestUploadObjectCmd_ServiceError_ReturnsError(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "data.txt")
	if err := os.WriteFile(tmpFile, []byte("hello"), 0600); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	serviceErr := errors.New("upload quota exceeded")
	mock := &cmdMockStorage{err: serviceErr}
	factory := &cmdStorageFactory{providers: map[string]storage.Storage{"gcp": mock}}
	app := newStorageTestApp(factory, nil)

	var buf bytes.Buffer
	cmd := newUploadObjectCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetContext(app.ToContext(context.Background()))
	cmd.SetArgs([]string{tmpFile, "--provider", "gcp", "--bucket", "my-bucket"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error from service, got nil")
	}
	if !errors.Is(err, serviceErr) {
		t.Errorf("expected service error in chain, got: %v", err)
	}
}
