package gcp

import (
	"encoding/base64"
	"testing"
	"time"

	gcpstorage "cloud.google.com/go/storage"
)

func TestMapLifecycleRules_NilSlice(t *testing.T) {
	result := mapLifecycleRules(nil)
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestMapLifecycleRules_EmptySlice(t *testing.T) {
	result := mapLifecycleRules([]gcpstorage.LifecycleRule{})
	if result != nil {
		t.Errorf("expected nil for empty slice, got %v", result)
	}
}

func TestMapLifecycleRules_DeleteAction(t *testing.T) {
	rules := []gcpstorage.LifecycleRule{
		{
			Action:    gcpstorage.LifecycleAction{Type: "Delete"},
			Condition: gcpstorage.LifecycleCondition{AgeInDays: 30},
		},
	}
	result := mapLifecycleRules(rules)
	if len(result) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(result))
	}
	if result[0].Action != "Delete" {
		t.Errorf("expected action 'Delete', got %q", result[0].Action)
	}
	if result[0].Condition.Age != 30 {
		t.Errorf("expected age 30, got %d", result[0].Condition.Age)
	}
}

func TestMapLifecycleRules_SetStorageClassAction(t *testing.T) {
	rules := []gcpstorage.LifecycleRule{
		{
			Action: gcpstorage.LifecycleAction{
				Type:         "SetStorageClass",
				StorageClass: "NEARLINE",
			},
			Condition: gcpstorage.LifecycleCondition{AgeInDays: 90},
		},
	}
	result := mapLifecycleRules(rules)
	if len(result) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(result))
	}
	expected := "SetStorageClass to NEARLINE"
	if result[0].Action != expected {
		t.Errorf("expected action %q, got %q", expected, result[0].Action)
	}
}

func TestMapLifecycleRules_AllConditions(t *testing.T) {
	createdBefore := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	rules := []gcpstorage.LifecycleRule{
		{
			Action: gcpstorage.LifecycleAction{Type: "Delete"},
			Condition: gcpstorage.LifecycleCondition{
				AgeInDays:             60,
				CreatedBefore:         createdBefore,
				MatchesStorageClasses: []string{"STANDARD", "NEARLINE"},
				NumNewerVersions:      3,
			},
		},
	}
	result := mapLifecycleRules(rules)
	c := result[0].Condition
	if c.Age != 60 {
		t.Errorf("expected Age=60, got %d", c.Age)
	}
	if !c.CreatedBefore.Equal(createdBefore) {
		t.Errorf("expected CreatedBefore=%v, got %v", createdBefore, c.CreatedBefore)
	}
	if len(c.MatchesStorageClass) != 2 {
		t.Errorf("expected 2 storage classes, got %d", len(c.MatchesStorageClass))
	}
	if c.NumNewerVersions != 3 {
		t.Errorf("expected NumNewerVersions=3, got %d", c.NumNewerVersions)
	}
}

func TestMapLogging_Nil(t *testing.T) {
	if mapLogging(nil) != nil {
		t.Error("expected nil for nil input")
	}
}

func TestMapLogging_Valid(t *testing.T) {
	input := &gcpstorage.BucketLogging{
		LogBucket:       "log-bucket",
		LogObjectPrefix: "prefix/",
	}
	result := mapLogging(input)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.LogBucket != "log-bucket" {
		t.Errorf("expected LogBucket=%q, got %q", "log-bucket", result.LogBucket)
	}
	if result.LogObjectPrefix != "prefix/" {
		t.Errorf("expected LogObjectPrefix=%q, got %q", "prefix/", result.LogObjectPrefix)
	}
}

func TestMapSoftDeletePolicy_Nil(t *testing.T) {
	if mapSoftDeletePolicy(nil) != nil {
		t.Error("expected nil for nil input")
	}
}

func TestMapSoftDeletePolicy_Valid(t *testing.T) {
	input := &gcpstorage.SoftDeletePolicy{
		RetentionDuration: 7 * 24 * time.Hour,
	}
	result := mapSoftDeletePolicy(input)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.RetentionDuration != 7*24*time.Hour {
		t.Errorf("expected 7 days, got %v", result.RetentionDuration)
	}
}

func TestMapPublicAccessPrevention(t *testing.T) {
	tests := []struct {
		name     string
		input    gcpstorage.PublicAccessPrevention
		expected string
	}{
		{"Enforced", gcpstorage.PublicAccessPreventionEnforced, "Enforced"},
		{"Inherited", gcpstorage.PublicAccessPreventionInherited, "Inherited"},
		{"Unknown", gcpstorage.PublicAccessPrevention(99), "Unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapPublicAccessPrevention(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestMapBucketEncryption_Nil(t *testing.T) {
	if mapBucketEncryption(nil) != nil {
		t.Error("expected nil for nil input")
	}
}

func TestMapBucketEncryption_EmptyKeyName(t *testing.T) {
	input := &gcpstorage.BucketEncryption{DefaultKMSKeyName: ""}
	if mapBucketEncryption(input) != nil {
		t.Error("expected nil for empty key name")
	}
}

func TestMapBucketEncryption_ValidKmsKey(t *testing.T) {
	keyName := "projects/p/locations/l/keyRings/kr/cryptoKeys/k"
	input := &gcpstorage.BucketEncryption{DefaultKMSKeyName: keyName}
	result := mapBucketEncryption(input)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.KmsKeyName != keyName {
		t.Errorf("expected KmsKeyName=%q, got %q", keyName, result.KmsKeyName)
	}
	if result.Algorithm != "AES256" {
		t.Errorf("expected Algorithm=AES256, got %q", result.Algorithm)
	}
}

func TestMapRetentionPolicy_Nil(t *testing.T) {
	if mapRetentionPolicy(nil) != nil {
		t.Error("expected nil for nil input")
	}
}

func TestMapRetentionPolicy_Unlocked(t *testing.T) {
	input := &gcpstorage.RetentionPolicy{
		RetentionPeriod: 30 * 24 * time.Hour,
		IsLocked:        false,
	}
	result := mapRetentionPolicy(input)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.IsLocked {
		t.Error("expected IsLocked=false")
	}
	if result.RetentionPeriod != 30*24*time.Hour {
		t.Errorf("expected 30 days, got %v", result.RetentionPeriod)
	}
}

func TestMapRetentionPolicy_Locked(t *testing.T) {
	input := &gcpstorage.RetentionPolicy{
		RetentionPeriod: 90 * 24 * time.Hour,
		IsLocked:        true,
	}
	result := mapRetentionPolicy(input)
	if !result.IsLocked {
		t.Error("expected IsLocked=true")
	}
}

func TestFormatMD5_EmptyHash(t *testing.T) {
	if formatMD5(nil) != "" {
		t.Error("expected empty string for nil hash")
	}
	if formatMD5([]byte{}) != "" {
		t.Error("expected empty string for empty hash")
	}
}

func TestFormatMD5_ValidHash(t *testing.T) {
	hash := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
	result := formatMD5(hash)
	expected := base64.StdEncoding.EncodeToString(hash)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFormatCRC32C_Zero(t *testing.T) {
	if formatCRC32C(0) != "" {
		t.Error("expected empty string for zero CRC32C")
	}
}

func TestFormatCRC32C_ValidValue(t *testing.T) {
	result := formatCRC32C(0xDEADBEEF)
	// Big-endian bytes: 0xDE, 0xAD, 0xBE, 0xEF
	expected := base64.StdEncoding.EncodeToString([]byte{0xDE, 0xAD, 0xBE, 0xEF})
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}
