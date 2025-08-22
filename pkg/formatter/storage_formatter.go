// File: pkg/formatter/storage_formatter.go
package formatter

import (
	"fmt"
	"strings"
	"synkronus/pkg/common"
	"synkronus/pkg/storage"
	"time"
)

type StorageFormatter struct{}

func NewStorageFormatter() *StorageFormatter {
	return &StorageFormatter{}
}

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

	sb.WriteString(f.formatOverviewSection(bucket))
	sb.WriteString(f.formatAccessControlSection(bucket))
	sb.WriteString(f.formatDataProtectionSection(bucket))
	sb.WriteString(f.formatLifecycleSection(bucket))
	sb.WriteString(f.formatLabelsSection(bucket))

	return sb.String()
}

func (f *StorageFormatter) formatOverviewSection(bucket storage.Bucket) string {
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
		protectionTable.AddRow([]string{"Encryption (CMEK)", "Google-managed"})
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
