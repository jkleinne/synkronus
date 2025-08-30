// File: pkg/formatter/storage_formatter.go
package formatter

import (
	"fmt"
	"strings"
	"synkronus/pkg/common"
	"synkronus/pkg/storage"
	"time"
)

const (
	// Used to indicate a directory/prefix in the object list
	directoryMarker = "(DIR)"
	// Used when a specific timestamp is not available (e.g., S3 object creation time)
	timeNotAvailable = "N/A"
)

type StorageFormatter struct{}

func NewStorageFormatter() *StorageFormatter {
	return &StorageFormatter{}
}

// --- Bucket Formatters ---

func (f *StorageFormatter) FormatBucketList(buckets []storage.Bucket) string {
	table := NewTable([]string{"BUCKET NAME", "PROVIDER", "LOCATION", "USAGE", "STORAGE CLASS", "CREATED"})

	for _, bucket := range buckets {
		usageFormatted := storage.FormatBytes(bucket.UsageBytes)
		createdFormatted := bucket.CreatedAt.Format("2006-01-02")

		table.AddRow([]string{
			bucket.Name,
			string(bucket.Provider),
			bucket.Location,
			usageFormatted,
			bucket.StorageClass,
			createdFormatted,
		})
	}

	return table.String()
}

func (f *StorageFormatter) FormatBucketDetails(bucket storage.Bucket) string {
	var sb strings.Builder

	sb.WriteString(FormatHeaderSection("Bucket: " + bucket.Name))
	sb.WriteString("\n\n")

	sb.WriteString(f.formatBucketOverviewSection(bucket))
	sb.WriteString(f.formatAccessControlSection(bucket))
	sb.WriteString(f.formatDataProtectionSection(bucket))
	sb.WriteString(f.formatLifecycleSection(bucket))
	sb.WriteString(f.formatLabelsSection(bucket))

	return sb.String()
}

func (f *StorageFormatter) formatBucketOverviewSection(bucket storage.Bucket) string {
	var sb strings.Builder

	sb.WriteString(FormatSectionTitle("Overview"))
	sb.WriteString("\n")

	overviewTable := NewTable([]string{"Parameter", "Value"})
	overviewTable.AddRow([]string{"Provider", string(bucket.Provider)})
	overviewTable.AddRow([]string{"Location", bucket.Location})
	if bucket.Provider == common.GCP && bucket.LocationType != "" {
		overviewTable.AddRow([]string{"Location Type", bucket.LocationType})
	}
	overviewTable.AddRow([]string{"Default Storage Class", bucket.StorageClass})
	overviewTable.AddRow([]string{"Usage (Total Bytes)", storage.FormatBytes(bucket.UsageBytes)})

	if bucket.Provider == common.GCP && bucket.Autoclass != nil {
		autoclassStatus := "Disabled"
		if bucket.Autoclass.Enabled {
			autoclassStatus = "Enabled"
		}
		overviewTable.AddRow([]string{"Autoclass", autoclassStatus})
	}

	requesterPaysStatus := "Disabled"
	if bucket.RequesterPays {
		requesterPaysStatus = "Enabled"
	}
	if bucket.Provider == common.GCP {
		overviewTable.AddRow([]string{"Requester Pays", requesterPaysStatus})
	}

	overviewTable.AddRow([]string{"Created On", bucket.CreatedAt.Format(time.RFC1123)})
	overviewTable.AddRow([]string{"Updated On", bucket.UpdatedAt.Format(time.RFC1123)})

	sb.WriteString(overviewTable.String())
	sb.WriteString("\n\n")

	return sb.String()
}

func (f *StorageFormatter) formatAccessControlSection(bucket storage.Bucket) string {
	var sb strings.Builder

	sb.WriteString(FormatSectionTitle("Access Control & Logging"))
	sb.WriteString("\n")

	// --- Configuration Summary Table ---
	configTable := NewTable([]string{"Configuration", "Status"})

	isUBLAEnabled := bucket.UniformBucketLevelAccess != nil && bucket.UniformBucketLevelAccess.Enabled

	if bucket.UniformBucketLevelAccess != nil {
		status := "Disabled (Fine-grained via ACLs/IAM)"
		if isUBLAEnabled {
			status = "Enabled (Uniform via IAM)"
		}
		configTable.AddRow([]string{"Uniform Bucket-Level Access", status})
	}
	configTable.AddRow([]string{"Public Access Prevention", strings.ToTitle(bucket.PublicAccessPrevention)})

	if bucket.Logging != nil {
		// Display a recognizable URI based on the provider
		prefix := ""
		if bucket.Provider == common.GCP {
			prefix = "gs://"
		} else if bucket.Provider == common.AWS {
			prefix = "s3://"
		}
		configTable.AddRow([]string{"Usage Logging", fmt.Sprintf("%s%s/%s", prefix, bucket.Logging.LogBucket, bucket.Logging.LogObjectPrefix)})
	} else {
		configTable.AddRow([]string{"Usage Logging", "Not configured"})
	}
	sb.WriteString(configTable.String())
	sb.WriteString("\n\n")

	// --- IAM Policy Display ---
	sb.WriteString(f.formatIAMPolicy(bucket.IAMPolicy))

	// --- ACL Display Logic ---
	sb.WriteString(f.formatACLs(bucket, isUBLAEnabled))

	return sb.String()
}

func (f *StorageFormatter) formatIAMPolicy(policy *storage.IAMPolicy) string {
	var sb strings.Builder

	sb.WriteString("Identity and Access Management (IAM) Policy:\n")

	if policy == nil {
		sb.WriteString("  (Could not retrieve IAM policy - check permissions)\n\n")
		return sb.String()
	}

	if len(policy.Bindings) == 0 && !policy.HasConditions {
		sb.WriteString("  (No IAM bindings found)\n\n")
		return sb.String()
	}

	iamTable := NewTable([]string{"Role", "Principal(s)"})

	for _, binding := range policy.Bindings {
		if len(binding.Principals) == 0 {
			continue
		}

		// Add the first principal on the same row as the role
		iamTable.AddRow([]string{binding.Role, binding.Principals[0]})

		// Add subsequent principals on new rows, leaving the role column empty for visual grouping
		for i := 1; i < len(binding.Principals); i++ {
			iamTable.AddRow([]string{"", binding.Principals[i]})
		}
	}

	sb.WriteString(iamTable.String())
	sb.WriteString("\n")

	if policy.HasConditions {
		sb.WriteString("Note: This policy contains conditional bindings which are not displayed here.\n")
	}
	sb.WriteString("\n")

	return sb.String()
}

func (f *StorageFormatter) formatACLs(bucket storage.Bucket, isUBLAEnabled bool) string {
	var sb strings.Builder

	if isUBLAEnabled {
		// If UBLA is enabled, ACLs are inactive
		sb.WriteString("Access Control List (ACLs):\n")
		sb.WriteString("  (Inactive because Uniform Bucket-Level Access is enabled. IAM policies control access.)\n\n")
		return sb.String()
	}

	if len(bucket.ACLs) > 0 {
		// Display ACLs if UBLA is disabled (or not applicable) and ACLs exist
		sb.WriteString("Access Control List (ACLs) - (Fine-grained object control):\n")
		aclTable := NewTable([]string{"Entity", "Role"})
		for _, acl := range bucket.ACLs {
			aclTable.AddRow([]string{acl.Entity, acl.Role})
		}
		sb.WriteString(aclTable.String())
		sb.WriteString("\n\n")
	}

	return sb.String()
}

func (f *StorageFormatter) formatDataProtectionSection(bucket storage.Bucket) string {
	var sb strings.Builder

	sb.WriteString(FormatSectionTitle("Data Protection"))
	sb.WriteString("\n")

	protectionTable := NewTable([]string{"Feature", "Configuration"})

	if bucket.Encryption != nil && bucket.Encryption.KmsKeyName != "" {
		protectionTable.AddRow([]string{"Encryption (CMEK)", bucket.Encryption.KmsKeyName})
	} else {
		// Default to assuming Google-managed for GCP if not specified
		defaultEncryption := "Provider-managed"
		if bucket.Provider == common.GCP {
			defaultEncryption = "Google-managed"
		}
		protectionTable.AddRow([]string{"Encryption (CMEK)", defaultEncryption})
	}

	if bucket.Versioning != nil {
		status := "Suspended"
		if bucket.Versioning.Enabled {
			status = "Enabled"
		}
		protectionTable.AddRow([]string{"Object Versioning", status})
	}

	if bucket.SoftDeletePolicy != nil {
		protectionTable.AddRow([]string{"Soft Delete Policy", fmt.Sprintf("Enabled (Retention: %v)", bucket.SoftDeletePolicy.RetentionDuration)})
	} else {
		protectionTable.AddRow([]string{"Soft Delete Policy", "Disabled"})
	}

	if bucket.RetentionPolicy != nil {
		lockedStatus := ""
		if bucket.RetentionPolicy.IsLocked {
			lockedStatus = ", Locked"
		}
		protectionTable.AddRow([]string{"Retention Policy", fmt.Sprintf("Enabled (Period: %v%s)", bucket.RetentionPolicy.RetentionPeriod, lockedStatus)})
	} else {
		protectionTable.AddRow([]string{"Retention Policy", "Disabled"})
	}

	sb.WriteString(protectionTable.String())
	sb.WriteString("\n\n")

	return sb.String()
}

func (f *StorageFormatter) formatLifecycleSection(bucket storage.Bucket) string {
	if len(bucket.LifecycleRules) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(FormatSectionTitle("Lifecycle Rules"))
	sb.WriteString("\n")

	rulesTable := NewTable([]string{"Action", "Condition"})
	for _, rule := range bucket.LifecycleRules {
		var conditions []string
		if rule.Condition.Age > 0 {
			conditions = append(conditions, fmt.Sprintf("Age > %d days", rule.Condition.Age))
		}
		if rule.Condition.NumNewerVersions > 0 {
			conditions = append(conditions, fmt.Sprintf("NumNewerVersions = %d", rule.Condition.NumNewerVersions))
		}
		if len(rule.Condition.MatchesStorageClass) > 0 {
			conditions = append(conditions, fmt.Sprintf("StorageClass IN (%s)", strings.Join(rule.Condition.MatchesStorageClass, ", ")))
		}
		if !rule.Condition.CreatedBefore.IsZero() {
			conditions = append(conditions, fmt.Sprintf("CreatedBefore = %s", rule.Condition.CreatedBefore.Format("2006-01-02")))
		}
		rulesTable.AddRow([]string{rule.Action, strings.Join(conditions, " AND ")})
	}

	sb.WriteString(rulesTable.String())
	sb.WriteString("\n\n")

	return sb.String()
}

func (f *StorageFormatter) formatLabelsSection(bucket storage.Bucket) string {
	if len(bucket.Labels) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(FormatSectionTitle("Labels"))
	sb.WriteString("\n")
	labelsTable := NewTable([]string{"Key", "Value"})
	for k, v := range bucket.Labels {
		labelsTable.AddRow([]string{k, v})
	}
	sb.WriteString(labelsTable.String())
	sb.WriteString("\n\n")

	return sb.String()
}

// --- Object Formatters ---

func (f *StorageFormatter) FormatObjectList(list storage.ObjectList) string {
	var sb strings.Builder

	// Display context information
	sb.WriteString(fmt.Sprintf("Listing objects in bucket: %s\n", list.BucketName))
	if list.Prefix != "" {
		sb.WriteString(fmt.Sprintf("Prefix: %s\n", list.Prefix))
	}
	sb.WriteString("\n")

	if len(list.Objects) == 0 && len(list.CommonPrefixes) == 0 {
		sb.WriteString("No objects or directories found.\n")
		return sb.String()
	}

	table := NewTable([]string{"KEY", "SIZE", "STORAGE CLASS", "LAST MODIFIED"})

	// Add directories first
	for _, prefix := range list.CommonPrefixes {
		table.AddRow([]string{
			prefix,
			directoryMarker,
			"",
			"",
		})
	}

	// Add objects
	for _, obj := range list.Objects {
		table.AddRow([]string{
			obj.Key,
			storage.FormatBytes(obj.Size),
			obj.StorageClass,
			obj.LastModified.Format(time.RFC3339),
		})
	}

	sb.WriteString(table.String())
	return sb.String()
}

func (f *StorageFormatter) FormatObjectDetails(object storage.Object) string {
	var sb strings.Builder

	sb.WriteString(FormatHeaderSection("Object: " + object.Key))
	sb.WriteString("\n\n")

	sb.WriteString(f.formatObjectOverviewSection(object))
	sb.WriteString(f.formatObjectHttpHeadersSection(object))
	sb.WriteString(f.formatObjectMetadataSection(object))

	return sb.String()
}

func (f *StorageFormatter) formatObjectOverviewSection(object storage.Object) string {
	var sb strings.Builder

	sb.WriteString(FormatSectionTitle("Overview"))
	sb.WriteString("\n")

	overviewTable := NewTable([]string{"Parameter", "Value"})
	overviewTable.AddRow([]string{"Bucket", object.Bucket})
	overviewTable.AddRow([]string{"Provider", string(object.Provider)})
	overviewTable.AddRow([]string{"Size", storage.FormatBytes(object.Size)})
	overviewTable.AddRow([]string{"Storage Class", object.StorageClass})

	// Handle timestamps
	overviewTable.AddRow([]string{"Last Modified", object.LastModified.Format(time.RFC1123)})

	createdAtStr := timeNotAvailable
	if !object.CreatedAt.IsZero() {
		createdAtStr = object.CreatedAt.Format(time.RFC1123)
	}
	overviewTable.AddRow([]string{"Created At", createdAtStr})

	// Handle encryption
	if object.Encryption != nil {
		encryptionDetails := object.Encryption.KmsKeyName
		if object.Encryption.Algorithm != "" {
			encryptionDetails = fmt.Sprintf("%s (%s)", encryptionDetails, object.Encryption.Algorithm)
		}
		overviewTable.AddRow([]string{"Encryption", encryptionDetails})
	} else {
		overviewTable.AddRow([]string{"Encryption", "N/A"})
	}

	// Handle checksums/identifiers
	overviewTable.AddRow([]string{"ETag", object.ETag})
	if object.MD5Hash != "" {
		overviewTable.AddRow([]string{"MD5 Hash (Base64)", object.MD5Hash})
	}
	if object.Provider == common.GCP {
		if object.CRC32C != "" {
			overviewTable.AddRow([]string{"CRC32C (Base64)", object.CRC32C})
		}
		overviewTable.AddRow([]string{"Generation", fmt.Sprintf("%d", object.Generation)})
		overviewTable.AddRow([]string{"Metageneration", fmt.Sprintf("%d", object.Metageneration)})
	}

	sb.WriteString(overviewTable.String())
	sb.WriteString("\n\n")

	return sb.String()
}

func (f *StorageFormatter) formatObjectHttpHeadersSection(object storage.Object) string {
	// Check if any HTTP headers are set
	if object.ContentType == "" && object.ContentEncoding == "" && object.ContentLanguage == "" &&
		object.CacheControl == "" && object.ContentDisposition == "" {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(FormatSectionTitle("HTTP Headers"))
	sb.WriteString("\n")

	headersTable := NewTable([]string{"Header", "Value"})

	if object.ContentType != "" {
		headersTable.AddRow([]string{"Content-Type", object.ContentType})
	}
	if object.ContentEncoding != "" {
		headersTable.AddRow([]string{"Content-Encoding", object.ContentEncoding})
	}
	if object.ContentLanguage != "" {
		headersTable.AddRow([]string{"Content-Language", object.ContentLanguage})
	}
	if object.CacheControl != "" {
		headersTable.AddRow([]string{"Cache-Control", object.CacheControl})
	}
	if object.ContentDisposition != "" {
		headersTable.AddRow([]string{"Content-Disposition", object.ContentDisposition})
	}

	sb.WriteString(headersTable.String())
	sb.WriteString("\n\n")

	return sb.String()
}

func (f *StorageFormatter) formatObjectMetadataSection(object storage.Object) string {
	if len(object.Metadata) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(FormatSectionTitle("User-Defined Metadata"))
	sb.WriteString("\n")

	metadataTable := NewTable([]string{"Key", "Value"})
	for k, v := range object.Metadata {
		metadataTable.AddRow([]string{k, v})
	}

	sb.WriteString(metadataTable.String())
	sb.WriteString("\n\n")

	return sb.String()
}
