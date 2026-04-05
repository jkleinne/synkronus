package main

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"synkronus/internal/domain"
	"synkronus/internal/domain/sql"
)

// Note: The commands in this package render output via output.Render(os.Stdout, ...)
// which writes to the real os.Stdout, not the cobra command's output writer.
// Tests therefore verify behavior (return value, error chain) rather than
// capturing rendered output through cmd.SetOut.

// --- describe-instance tests ---

func TestDescribeInstanceCmd_HappyPath_ReturnsInstance(t *testing.T) {
	want := sql.Instance{
		Name:            "prod-db",
		Provider:        domain.GCP,
		State:           "RUNNABLE",
		DatabaseVersion: "POSTGRES_15",
	}
	mock := &cmdMockSQL{instance: want}
	sqlFactory := &cmdSqlFactory{providers: map[string]sql.SQL{"gcp": mock}}
	app := newSqlTestApp(sqlFactory)

	var buf bytes.Buffer
	cmd := newDescribeInstanceCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetContext(app.ToContext(context.Background()))
	cmd.SetArgs([]string{"--provider", "gcp", "prod-db"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Output is rendered to os.Stdout (not the cobra writer), so we assert on
	// successful execution rather than captured output content.
}

func TestDescribeInstanceCmd_ServiceError_ReturnsWrappedError(t *testing.T) {
	serviceErr := errors.New("instance not found")
	mock := &cmdMockSQL{err: serviceErr}
	sqlFactory := &cmdSqlFactory{providers: map[string]sql.SQL{"gcp": mock}}
	app := newSqlTestApp(sqlFactory)

	var buf bytes.Buffer
	cmd := newDescribeInstanceCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetContext(app.ToContext(context.Background()))
	cmd.SetArgs([]string{"--provider", "gcp", "missing-db"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, serviceErr) {
		t.Errorf("expected service error in chain, got: %v", err)
	}
	// The service wraps with the instance name for context.
	if !strings.Contains(err.Error(), "missing-db") {
		t.Errorf("expected instance name in error message for context, got: %v", err)
	}
}

func TestDescribeInstanceCmd_MissingProviderFlag_ReturnsError(t *testing.T) {
	app := newSqlTestApp(&cmdSqlFactory{})

	var buf bytes.Buffer
	cmd := newDescribeInstanceCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetContext(app.ToContext(context.Background()))
	cmd.SetArgs([]string{"prod-db"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for missing --provider flag, got nil")
	}
}

func TestDescribeInstanceCmd_MissingInstanceNameArg_ReturnsError(t *testing.T) {
	app := newSqlTestApp(&cmdSqlFactory{})

	var buf bytes.Buffer
	cmd := newDescribeInstanceCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetContext(app.ToContext(context.Background()))
	cmd.SetArgs([]string{"--provider", "gcp"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for missing instance name argument, got nil")
	}
}
