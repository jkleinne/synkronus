package storage

import "testing"

func TestCreateBucketOptions_DefaultsAreNil(t *testing.T) {
	opts := CreateBucketOptions{}
	if opts.Versioning != nil {
		t.Error("zero-value Versioning should be nil")
	}
	if opts.UniformBucketLevelAccess != nil {
		t.Error("zero-value UniformBucketLevelAccess should be nil")
	}
	if opts.PublicAccessPrevention != nil {
		t.Error("zero-value PublicAccessPrevention should be nil")
	}
	if opts.Labels != nil {
		t.Error("zero-value Labels should be nil")
	}
	if opts.StorageClass != "" {
		t.Errorf("zero-value StorageClass = %q, want empty", opts.StorageClass)
	}
}

func TestPublicAccessPreventionConstants(t *testing.T) {
	if PublicAccessPreventionEnforced != "enforced" {
		t.Errorf("PublicAccessPreventionEnforced = %q, want enforced", PublicAccessPreventionEnforced)
	}
	if PublicAccessPreventionInherited != "inherited" {
		t.Errorf("PublicAccessPreventionInherited = %q, want inherited", PublicAccessPreventionInherited)
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected string
	}{
		{"Negative", -1, "N/A"},
		{"Zero", 0, "0 B"},
		{"OneByte", 1, "1 B"},
		{"SmallBytes", 500, "500 B"},
		{"JustBelowKB", 1023, "1023 B"},
		{"ExactlyOneKB", 1024, "1.0 KB"},
		{"OneAndHalfKB", 1536, "1.5 KB"},
		{"ExactlyOneMB", 1048576, "1.0 MB"},
		{"OneAndHalfMB", 1572864, "1.5 MB"},
		{"ExactlyOneGB", 1073741824, "1.0 GB"},
		{"ExactlyOneTB", 1099511627776, "1.0 TB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatBytes(tt.input)
			if result != tt.expected {
				t.Errorf("FormatBytes(%d) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
