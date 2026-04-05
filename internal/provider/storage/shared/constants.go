package shared

// Display-facing status strings used in output views.
const (
	StatusEnabled  = "Enabled"
	StatusDisabled = "Disabled"
)

// URI prefixes for bucket paths in output views.
const (
	GCSBucketURIPrefix = "gs://"
	S3BucketURIPrefix  = "s3://"
)

// DirectoryMarker is the display string for directory entries in object lists.
const DirectoryMarker = "(DIR)"

// TimeNotAvailable is used when a timestamp cannot be retrieved.
const TimeNotAvailable = "N/A"

// DefaultEncryptionAlgorithm is the default server-side encryption algorithm.
const DefaultEncryptionAlgorithm = "AES256"

// EncryptionProviderManaged describes encryption managed by the cloud provider.
const EncryptionProviderManaged = "Provider-managed"

// EncryptionGoogleManaged describes encryption managed by Google.
const EncryptionGoogleManaged = "Google-managed"
