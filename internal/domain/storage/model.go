package storage

import (
	"fmt"
	"synkronus/internal/domain"
	"time"
)

type Bucket struct {
	Name         string          `json:"name" yaml:"name"`
	Provider     domain.Provider `json:"provider" yaml:"provider"`
	Location     string          `json:"location" yaml:"location"`
	LocationType string          `json:"location_type,omitempty" yaml:"location_type,omitempty"`
	StorageClass string          `json:"storage_class" yaml:"storage_class"`
	CreatedAt    time.Time       `json:"created_at" yaml:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at" yaml:"updated_at"`
	// A value of -1 indicates that the usage is unknown or could not be retrieved
	UsageBytes    int64             `json:"usage_bytes" yaml:"usage_bytes"`
	RequesterPays bool              `json:"requester_pays" yaml:"requester_pays"`
	Labels        map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`

	Autoclass                *Autoclass                `json:"autoclass,omitempty" yaml:"autoclass,omitempty"`
	IAMPolicy                *IAMPolicy                `json:"iam_policy,omitempty" yaml:"iam_policy,omitempty"`
	ACLs                     []ACLRule                 `json:"acls,omitempty" yaml:"acls,omitempty"`
	LifecycleRules           []LifecycleRule           `json:"lifecycle_rules,omitempty" yaml:"lifecycle_rules,omitempty"`
	Logging                  *Logging                  `json:"logging,omitempty" yaml:"logging,omitempty"`
	Versioning               *Versioning               `json:"versioning,omitempty" yaml:"versioning,omitempty"`
	SoftDeletePolicy         *SoftDeletePolicy         `json:"soft_delete_policy,omitempty" yaml:"soft_delete_policy,omitempty"`
	UniformBucketLevelAccess *UniformBucketLevelAccess `json:"uniform_bucket_level_access,omitempty" yaml:"uniform_bucket_level_access,omitempty"`
	PublicAccessPrevention   string                    `json:"public_access_prevention,omitempty" yaml:"public_access_prevention,omitempty"`
	Encryption               *Encryption               `json:"encryption,omitempty" yaml:"encryption,omitempty"`
	RetentionPolicy          *RetentionPolicy          `json:"retention_policy,omitempty" yaml:"retention_policy,omitempty"`
}

// ObjectList represents the results of a ListObjects operation using delimiters (simulating directories)
type ObjectList struct {
	BucketName     string   `json:"bucket_name" yaml:"bucket_name"`
	Prefix         string   `json:"prefix,omitempty" yaml:"prefix,omitempty"`
	Objects        []Object `json:"objects,omitempty" yaml:"objects,omitempty"`
	CommonPrefixes []string `json:"common_prefixes,omitempty" yaml:"common_prefixes,omitempty"`
	IsTruncated    bool     `json:"is_truncated,omitempty" yaml:"is_truncated,omitempty"`
}

// Object represents a single object (file) within a storage bucket
type Object struct {
	Key          string          `json:"key" yaml:"key"`
	Bucket       string          `json:"bucket" yaml:"bucket"`
	Provider     domain.Provider `json:"provider" yaml:"provider"`
	Size         int64           `json:"size" yaml:"size"`
	StorageClass string          `json:"storage_class" yaml:"storage_class"`
	LastModified time.Time       `json:"last_modified" yaml:"last_modified"`
	CreatedAt    time.Time       `json:"created_at" yaml:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at" yaml:"updated_at"`
	ETag         string          `json:"etag" yaml:"etag"`

	ContentType        string `json:"content_type,omitempty" yaml:"content_type,omitempty"`
	ContentEncoding    string `json:"content_encoding,omitempty" yaml:"content_encoding,omitempty"`
	ContentLanguage    string `json:"content_language,omitempty" yaml:"content_language,omitempty"`
	CacheControl       string `json:"cache_control,omitempty" yaml:"cache_control,omitempty"`
	ContentDisposition string `json:"content_disposition,omitempty" yaml:"content_disposition,omitempty"`

	// Checksums (Base64 encoded strings)
	MD5Hash string `json:"md5_hash,omitempty" yaml:"md5_hash,omitempty"`
	CRC32C  string `json:"crc32c,omitempty" yaml:"crc32c,omitempty"` // GCP specific

	// Versioning information
	Generation     int64  `json:"generation,omitempty" yaml:"generation,omitempty"`         // GCP specific
	Metageneration int64  `json:"metageneration,omitempty" yaml:"metageneration,omitempty"` // GCP specific
	VersionID      string `json:"version_id,omitempty" yaml:"version_id,omitempty"`         // AWS specific

	Encryption *Encryption `json:"encryption,omitempty" yaml:"encryption,omitempty"`

	Metadata map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

type Autoclass struct {
	Enabled bool `json:"enabled" yaml:"enabled"`
}

type Versioning struct {
	Enabled bool `json:"enabled" yaml:"enabled"`
}

type Logging struct {
	LogBucket       string `json:"log_bucket" yaml:"log_bucket"`
	LogObjectPrefix string `json:"log_object_prefix" yaml:"log_object_prefix"`
}

type SoftDeletePolicy struct {
	RetentionDuration time.Duration `json:"retention_duration" yaml:"retention_duration"`
}

type UniformBucketLevelAccess struct {
	Enabled bool `json:"enabled" yaml:"enabled"`
}

type Encryption struct {
	// The name or ARN of the KMS key used for encryption (CMEK)
	KmsKeyName string `json:"kms_key_name,omitempty" yaml:"kms_key_name,omitempty"`
	// The algorithm used (e.g., AES256)
	Algorithm string `json:"algorithm,omitempty" yaml:"algorithm,omitempty"`
}

type RetentionPolicy struct {
	RetentionPeriod time.Duration `json:"retention_period" yaml:"retention_period"`
	IsLocked        bool          `json:"is_locked" yaml:"is_locked"`
}

// IAMPolicy represents the IAM policy attached to a resource
type IAMPolicy struct {
	// GCP: associates a list of principals with a role
	Bindings []IAMBinding `json:"bindings,omitempty" yaml:"bindings,omitempty"`
	// AWS: S3 bucket policy statements
	Statements []PolicyStatement `json:"statements,omitempty" yaml:"statements,omitempty"`
}

// PolicyStatement represents a single statement in an AWS S3 bucket policy
type PolicyStatement struct {
	Effect     string                         `json:"effect" yaml:"effect"`
	Principals []string                       `json:"principals" yaml:"principals"`
	Actions    []string                       `json:"actions" yaml:"actions"`
	Resources  []string                       `json:"resources" yaml:"resources"`
	Conditions map[string]map[string][]string `json:"conditions,omitempty" yaml:"conditions,omitempty"`
}

// IAMCondition represents a conditional expression on an IAM binding.
type IAMCondition struct {
	Title       string `json:"title" yaml:"title"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Expression  string `json:"expression" yaml:"expression"`
}

// IAMBinding represents a single binding in an IAM policy
type IAMBinding struct {
	Role       string        `json:"role" yaml:"role"`
	Principals []string      `json:"principals,omitempty" yaml:"principals,omitempty"`
	Condition  *IAMCondition `json:"condition,omitempty" yaml:"condition,omitempty"`
}

type ACLRule struct {
	Entity string `json:"entity" yaml:"entity"` // e.g., "user-email@example.com", "allUsers"
	Role   string `json:"role" yaml:"role"`
}

type LifecycleRule struct {
	Action    string             `json:"action" yaml:"action"`
	Condition LifecycleCondition `json:"condition" yaml:"condition"`
}

type LifecycleCondition struct {
	Age                 int       `json:"age" yaml:"age"`
	CreatedBefore       time.Time `json:"created_before,omitempty" yaml:"created_before,omitempty"`
	MatchesStorageClass []string  `json:"matches_storage_class,omitempty" yaml:"matches_storage_class,omitempty"`
	NumNewerVersions    int       `json:"num_newer_versions" yaml:"num_newer_versions"`
	Prefix              string    `json:"prefix,omitempty" yaml:"prefix,omitempty"` // S3-specific
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

// PublicAccessPrevention values for CreateBucketOptions.
const (
	PublicAccessPreventionEnforced  = "enforced"
	PublicAccessPreventionInherited = "inherited"
)

// CreateBucketOptions contains the parameters for creating a new storage bucket.
// Optional fields use pointer types to distinguish "not set" (nil) from explicit values.
type CreateBucketOptions struct {
	Name                     string
	Location                 string
	StorageClass             string
	Labels                   map[string]string
	Versioning               *bool
	UniformBucketLevelAccess *bool
	PublicAccessPrevention   *string // "enforced", "inherited", or nil (provider default)
}

// UploadObjectOptions contains the parameters for uploading an object to a storage bucket.
// ContentType is optional — providers auto-detect from the object key extension if empty.
type UploadObjectOptions struct {
	BucketName  string
	ObjectKey   string
	ContentType string            // optional — auto-detected from key extension if empty
	Metadata    map[string]string // optional user-defined metadata
}

// DefaultMaxResults is the default cap for ListObjects when no explicit limit is given.
const DefaultMaxResults = 1000

// UpdateBucketOptions contains the parameters for updating an existing bucket.
// Nil pointer fields are left unchanged; non-nil fields are applied.
type UpdateBucketOptions struct {
	Name         string
	SetLabels    map[string]string // labels to add or overwrite
	RemoveLabels []string          // label keys to remove
	Versioning   *bool             // nil = don't change, true/false = set
}
