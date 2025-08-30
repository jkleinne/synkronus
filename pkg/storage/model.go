// File: pkg/storage/model.go
package storage

import (
	"fmt"
	"synkronus/pkg/common"
	"time"
)

type Bucket struct {
	Name         string
	Provider     common.Provider
	Location     string
	LocationType string
	StorageClass string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	// A value of -1 indicates that the usage is unknown or could not be retrieved
	UsageBytes    int64
	RequesterPays bool
	Labels        map[string]string

	Autoclass                *Autoclass
	IAMPolicy                *IAMPolicy
	ACLs                     []ACLRule
	LifecycleRules           []LifecycleRule
	Logging                  *Logging
	Versioning               *Versioning
	SoftDeletePolicy         *SoftDeletePolicy
	UniformBucketLevelAccess *UniformBucketLevelAccess
	PublicAccessPrevention   string
	Encryption               *Encryption
	RetentionPolicy          *RetentionPolicy
}

// Represents the results of a ListObjects operation using delimiters (simulating directories)
type ObjectList struct {
	BucketName     string
	Prefix         string
	Objects        []Object
	CommonPrefixes []string
}

// Represents a single object (file) within a storage bucket
type Object struct {
	Key          string
	Bucket       string
	Provider     common.Provider
	Size         int64
	StorageClass string
	LastModified time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
	ETag         string

	ContentType        string
	ContentEncoding    string
	ContentLanguage    string
	CacheControl       string
	ContentDisposition string

	// Checksums (Base64 encoded strings)
	MD5Hash string
	CRC32C  string // GCP specific

	// Versioning information
	Generation     int64 // GCP specific
	Metageneration int64 // GCP specific

	Encryption *Encryption

	Metadata map[string]string
}

type Autoclass struct {
	Enabled bool
}

type Versioning struct {
	Enabled bool
}

type Logging struct {
	LogBucket       string
	LogObjectPrefix string
}

type SoftDeletePolicy struct {
	RetentionDuration time.Duration
}

type UniformBucketLevelAccess struct {
	Enabled bool
}

type Encryption struct {
	// The name or ARN of the KMS key used for encryption (CMEK)
	KmsKeyName string
	// The algorithm used (e.g., AES256)
	Algorithm string
}

type RetentionPolicy struct {
	RetentionPeriod time.Duration
	IsLocked        bool
}

// Represents the IAM policy attached to a resource
type IAMPolicy struct {
	// Associates a list of principals with a role
	Bindings []IAMBinding
	// Indicates if the policy contains conditional bindings that are not displayed
	HasConditions bool
}

// Represents a single binding in an IAM policy
type IAMBinding struct {
	Role       string
	Principals []string
}

type ACLRule struct {
	Entity string // e.g., "user-email@example.com", "allUsers"
	Role   string
}

type LifecycleRule struct {
	Action    string
	Condition LifecycleCondition
}

type LifecycleCondition struct {
	Age                 int
	CreatedBefore       time.Time
	MatchesStorageClass []string
	NumNewerVersions    int
}

func FormatBytes(bytes int64) string {
	if bytes < 0 {
		return "N/A"
	}
	if bytes == 0 {
		return "0 B"
	}

	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	sizes := []string{"KB", "MB", "GB", "TB", "PB", "EB"}
	if exp >= len(sizes) {
		return fmt.Sprintf("%d B", bytes) // Fallback if extremely large
	}
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), sizes[exp])
}
