package config

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestConfig(t *testing.T) (*ConfigManager, string) {
	t.Helper()
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", ConfigDirName)
	if err := os.MkdirAll(configDir, ConfigDirPermissions); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	configPath := filepath.Join(configDir, ConfigFileName)
	content := `{"gcp": {"project": "test-project"}}`
	if err := os.WriteFile(configPath, []byte(content), ConfigFilePermissions); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}
	t.Setenv("HOME", tmpDir)
	cm, err := NewConfigManager()
	if err != nil {
		t.Fatalf("failed to create config manager: %v", err)
	}
	return cm, tmpDir
}

func TestLoadConfig_Valid(t *testing.T) {
	cm, _ := setupTestConfig(t)
	cfg, err := cm.LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.GCP == nil || cfg.GCP.Project != "test-project" {
		t.Errorf("expected GCP project 'test-project', got %+v", cfg.GCP)
	}
}

func TestLoadConfig_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	cm, err := NewConfigManager()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cfg, err := cm.LoadConfig()
	if err != nil {
		t.Fatalf("expected no error for missing config, got: %v", err)
	}
	if cfg.GCP != nil {
		t.Error("expected nil GCP config for missing file")
	}
}

func TestLoadConfig_MalformedJSON(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", ConfigDirName)
	if err := os.MkdirAll(configDir, ConfigDirPermissions); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	configPath := filepath.Join(configDir, ConfigFileName)
	if err := os.WriteFile(configPath, []byte(`{invalid json`), ConfigFilePermissions); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}
	t.Setenv("HOME", tmpDir)
	_, err := NewConfigManager()
	if err == nil {
		t.Fatal("expected error for malformed JSON config file")
	}
}

func TestSetValue_GetValue_RoundTrip(t *testing.T) {
	cm, _ := setupTestConfig(t)
	if err := cm.SetValue("aws.region", "us-east-1"); err != nil {
		t.Fatalf("SetValue failed: %v", err)
	}
	val, exists := cm.GetValue("aws.region")
	if !exists || val != "us-east-1" {
		t.Errorf("expected 'us-east-1', got %q (exists=%v)", val, exists)
	}
}

// TestDeleteValue_RequiredField verifies that deleting a required field is rejected
// with a validation error and returns deleted=false.
func TestDeleteValue_RequiredField(t *testing.T) {
	cm, _ := setupTestConfig(t)
	deleted, err := cm.DeleteValue("gcp.project")
	if err == nil {
		t.Fatal("expected error when deleting a required field, got nil")
	}
	if deleted {
		t.Error("expected deleted=false when deletion is rejected")
	}
}

// TestDeleteValue_AfterSet verifies that attempting to delete the only required field
// in a section is correctly rejected with a validation error.
func TestDeleteValue_AfterSet(t *testing.T) {
	cm, _ := setupTestConfig(t)

	if err := cm.SetValue("aws.region", "us-east-1"); err != nil {
		t.Fatalf("SetValue failed: %v", err)
	}

	// aws.region is marked required, so emptying it fails validation.
	deleted, err := cm.DeleteValue("aws.region")
	if err == nil {
		t.Fatal("expected error when deleting the only required field in a section")
	}
	if deleted {
		t.Error("expected deleted=false for rejected deletion")
	}
}

func TestDeleteValue_NonExistent(t *testing.T) {
	cm, _ := setupTestConfig(t)
	deleted, err := cm.DeleteValue("nonexistent.key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deleted {
		t.Error("expected false for non-existent key")
	}
}

func TestSetValue_UnrecognizedKey(t *testing.T) {
	cm, _ := setupTestConfig(t)
	err := cm.SetValue("invalid.unknown", "value")
	if err == nil {
		t.Fatal("expected error for unrecognized key")
	}
}

func TestSaveConfig_FilePermissions(t *testing.T) {
	cm, tmpDir := setupTestConfig(t)
	if err := cm.SaveConfig(); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}
	configPath := filepath.Join(tmpDir, ".config", ConfigDirName, ConfigFileName)
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("failed to stat config file: %v", err)
	}
	if info.Mode().Perm() != ConfigFilePermissions {
		t.Errorf("expected file permissions %o, got %o", ConfigFilePermissions, info.Mode().Perm())
	}
}

func TestRemoveProvider_Existing(t *testing.T) {
	cm, _ := setupTestConfig(t)

	// Add AWS provider first
	if err := cm.SetValue("aws.region", "us-east-1"); err != nil {
		t.Fatalf("SetValue failed: %v", err)
	}

	// Remove the entire AWS provider
	removed, err := cm.RemoveProvider("aws")
	if err != nil {
		t.Fatalf("RemoveProvider failed: %v", err)
	}
	if !removed {
		t.Error("expected removed=true for existing provider")
	}

	// Verify it's gone
	_, exists := cm.GetValue("aws.region")
	if exists {
		t.Error("aws.region should not exist after removing provider")
	}

	// GCP should still be intact
	cfg, err := cm.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if cfg.GCP == nil || cfg.GCP.Project != "test-project" {
		t.Error("GCP config should be untouched after removing AWS")
	}
	if cfg.AWS != nil {
		t.Error("AWS config should be nil after removal")
	}
}

func TestRemoveProvider_NonExistent(t *testing.T) {
	cm, _ := setupTestConfig(t)
	removed, err := cm.RemoveProvider("azure")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if removed {
		t.Error("expected removed=false for non-existent provider")
	}
}

func TestSaveConfig_DirectoryPermissions(t *testing.T) {
	cm, tmpDir := setupTestConfig(t)
	if err := cm.SaveConfig(); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}
	configDir := filepath.Join(tmpDir, ".config", ConfigDirName)
	info, err := os.Stat(configDir)
	if err != nil {
		t.Fatalf("failed to stat config dir: %v", err)
	}
	if info.Mode().Perm() != ConfigDirPermissions {
		t.Errorf("expected dir permissions %o, got %o", ConfigDirPermissions, info.Mode().Perm())
	}
}
