package main

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"synkronus/internal/domain/storage"
)

// --- delete-object tests ---

func TestDeleteObjectCmd_ForceFlag_SkipsConfirmationAndSucceeds(t *testing.T) {
	mock := &cmdMockStorage{}
	factory := &cmdStorageFactory{providers: map[string]storage.Storage{"gcp": mock}}
	app := newStorageTestApp(factory, nil)

	var buf bytes.Buffer
	cmd := newDeleteObjectCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetContext(app.ToContext(context.Background()))
	cmd.SetArgs([]string{"--provider", "gcp", "--bucket", "my-bucket", "--force", "objects/file.txt"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !mock.closeCalled {
		t.Error("expected provider client Close to be called after delete")
	}
}

func TestDeleteObjectCmd_ServiceError_ReturnsError(t *testing.T) {
	serviceErr := errors.New("object not found")
	mock := &cmdMockStorage{err: serviceErr}
	factory := &cmdStorageFactory{providers: map[string]storage.Storage{"gcp": mock}}
	app := newStorageTestApp(factory, nil)

	var buf bytes.Buffer
	cmd := newDeleteObjectCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetContext(app.ToContext(context.Background()))
	cmd.SetArgs([]string{"--provider", "gcp", "--bucket", "my-bucket", "--force", "objects/file.txt"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, serviceErr) {
		t.Errorf("expected service error in chain, got: %v", err)
	}
}

func TestDeleteObjectCmd_MissingProviderFlag_ReturnsError(t *testing.T) {
	app := newStorageTestApp(&cmdStorageFactory{}, nil)

	var buf bytes.Buffer
	cmd := newDeleteObjectCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetContext(app.ToContext(context.Background()))
	cmd.SetArgs([]string{"--bucket", "my-bucket", "--force", "objects/file.txt"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for missing --provider flag, got nil")
	}
}

func TestDeleteObjectCmd_MissingBucketFlag_ReturnsError(t *testing.T) {
	app := newStorageTestApp(&cmdStorageFactory{}, nil)

	var buf bytes.Buffer
	cmd := newDeleteObjectCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetContext(app.ToContext(context.Background()))
	cmd.SetArgs([]string{"--provider", "gcp", "--force", "objects/file.txt"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for missing --bucket flag, got nil")
	}
}

func TestDeleteObjectCmd_MissingObjectKeyArg_ReturnsError(t *testing.T) {
	app := newStorageTestApp(&cmdStorageFactory{}, nil)

	var buf bytes.Buffer
	cmd := newDeleteObjectCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetContext(app.ToContext(context.Background()))
	cmd.SetArgs([]string{"--provider", "gcp", "--bucket", "my-bucket", "--force"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for missing object key argument, got nil")
	}
}

func TestDeleteObjectCmd_ConfirmDeclined_AbortsWithoutDelete(t *testing.T) {
	mock := &cmdMockStorage{}
	factory := &cmdStorageFactory{providers: map[string]storage.Storage{"gcp": mock}}
	prompter := &mockPrompter{confirmed: false}
	app := newStorageTestApp(factory, prompter)

	var buf bytes.Buffer
	cmd := newDeleteObjectCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetContext(app.ToContext(context.Background()))
	cmd.SetArgs([]string{"--provider", "gcp", "--bucket", "my-bucket", "objects/file.txt"})

	err := cmd.Execute()
	if !errors.Is(err, ErrOperationAborted) {
		t.Errorf("expected ErrOperationAborted, got: %v", err)
	}
	if mock.closeCalled {
		t.Error("provider client should not be called when confirmation is declined")
	}
}
