package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// setupIntegrationTest creates an isolated temp HOME directory with the
// synkronus config directory pre-created, then sets HOME for the duration
// of the test so the config manager reads/writes there instead of the real
// user home.
func setupIntegrationTest(t *testing.T) {
	t.Helper()
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "synkronus")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	t.Setenv("HOME", tmpDir)
}

// executeCommand builds a fresh root command, wires the output buffer, and
// runs the supplied arguments. It returns whatever Cobra wrote to its output
// writer and any error returned by Execute.
//
// NOTE: the config sub-commands print via fmt.Printf/fmt.Println to real
// os.Stdout, not to cmd.OutOrStdout(). Those tests therefore verify
// correctness via error returns and config round-trip semantics rather than
// captured output strings.
func executeCommand(args ...string) (string, error) {
	cmd := newRootCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return buf.String(), err
}

// TestIntegration_ConfigSetAndGet verifies that a key written by "config set"
// is readable by "config get" in the same process.
func TestIntegration_ConfigSetAndGet(t *testing.T) {
	setupIntegrationTest(t)

	_, err := executeCommand("config", "set", "gcp.project", "test-proj")
	if err != nil {
		t.Fatalf("config set failed: %v", err)
	}

	// config get prints to os.Stdout, not the cobra buffer; verify via err only.
	_, err = executeCommand("config", "get", "gcp.project")
	if err != nil {
		t.Fatalf("config get failed after set: %v", err)
	}
}

// TestIntegration_ConfigList verifies that "config list" succeeds after at
// least one key is set.
func TestIntegration_ConfigList(t *testing.T) {
	setupIntegrationTest(t)

	_, err := executeCommand("config", "set", "gcp.project", "my-project")
	if err != nil {
		t.Fatalf("config set failed: %v", err)
	}

	_, err = executeCommand("config", "list")
	if err != nil {
		t.Fatalf("config list failed: %v", err)
	}
}

// TestIntegration_ConfigDelete_RequiredFieldRejected verifies that "config delete"
// returns an error when asked to remove a required field (all current fields carry
// validate:"required"), and that the key is still readable afterwards.
func TestIntegration_ConfigDelete_RequiredFieldRejected(t *testing.T) {
	setupIntegrationTest(t)

	_, err := executeCommand("config", "set", "gcp.project", "test-proj")
	if err != nil {
		t.Fatalf("config set failed: %v", err)
	}

	// Deleting a required field must be rejected by the validation layer.
	_, err = executeCommand("config", "delete", "gcp.project")
	if err == nil {
		t.Fatal("expected error when deleting a required field, got nil")
	}

	// The key must still exist because the deletion was rolled back.
	_, err = executeCommand("config", "get", "gcp.project")
	if err != nil {
		t.Errorf("key should still be present after rejected deletion, got error: %v", err)
	}
}

// TestIntegration_ConfigDelete_NonExistentKey verifies that "config delete"
// returns an error when the requested key has never been set.
func TestIntegration_ConfigDelete_NonExistentKey(t *testing.T) {
	setupIntegrationTest(t)

	_, err := executeCommand("config", "delete", "gcp.project")
	if err == nil {
		t.Fatal("expected error when deleting a non-existent key, got nil")
	}
}

// TestIntegration_OutputFlagAccepted verifies that each supported --output
// format value is accepted without error against a benign command.
func TestIntegration_OutputFlagAccepted(t *testing.T) {
	setupIntegrationTest(t)

	for _, format := range []string{"json", "yaml", "table"} {
		t.Run(format, func(t *testing.T) {
			_, err := executeCommand("--output", format, "config", "list")
			if err != nil {
				t.Errorf("--output %s should be accepted, got error: %v", format, err)
			}
		})
	}
}

// TestIntegration_InvalidOutputFormat verifies that an unrecognised --output
// value causes Execute to return a non-nil error.
func TestIntegration_InvalidOutputFormat(t *testing.T) {
	setupIntegrationTest(t)

	_, err := executeCommand("--output", "xml", "config", "list")
	if err == nil {
		t.Fatal("expected error for unsupported output format 'xml', got nil")
	}
}

// TestIntegration_InvalidProvider verifies that requesting an unsupported
// provider via --providers returns a non-nil error.
func TestIntegration_InvalidProvider(t *testing.T) {
	setupIntegrationTest(t)

	_, err := executeCommand("config", "set", "gcp.project", "test-proj")
	if err != nil {
		t.Fatalf("config set failed: %v", err)
	}

	_, err = executeCommand("storage", "list-buckets", "--providers", "azure")
	if err == nil {
		t.Fatal("expected error for unsupported provider 'azure', got nil")
	}
}
