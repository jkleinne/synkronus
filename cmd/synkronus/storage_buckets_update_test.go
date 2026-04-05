package main

import (
	"strings"
	"testing"
)

func TestParseSetLabels(t *testing.T) {
	// Verify Cobra's StringToString parsing handles key=value pairs
	cmd := newUpdateBucketCmd()
	cmd.SetArgs([]string{"my-bucket", "--provider", "gcp", "--set-labels", "env=prod,team=data"})

	// We can't fully execute (no app context), but we can verify flags parse
	err := cmd.ParseFlags([]string{"--provider", "gcp", "--set-labels", "env=prod,team=data", "--versioning=true"})
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	labels, err := cmd.Flags().GetStringToString("set-labels")
	if err != nil {
		t.Fatalf("unexpected error getting set-labels: %v", err)
	}
	if labels["env"] != "prod" || labels["team"] != "data" {
		t.Errorf("expected env=prod,team=data, got %v", labels)
	}
}

func TestUpdateBucketCmd_NoFlags(t *testing.T) {
	cmd := newUpdateBucketCmd()
	cmd.SetArgs([]string{"my-bucket", "--provider", "gcp"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when no mutation flags provided")
	}
	if !strings.Contains(err.Error(), "at least one of") {
		t.Errorf("expected 'at least one of' error, got: %v", err)
	}
}
