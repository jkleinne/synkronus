// File: pkg/formatter/storage_formatter.go
package formatter

import (
	"fmt"
	"strings"
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
	overviewTable.AddRow([]string{"Location / Region", bucket.Location})
	overviewTable.AddRow([]string{"Default Storage Class", bucket.StorageClass})
	overviewTable.AddRow([]string{"Usage (Total Bytes)", storage.FormatBytes(bucket.UsageBytes)})
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

	configTable := NewTable([]string{"Configuration", "Status"})
	if bucket.UniformBucketLevelAccess != nil {
		status := "Disabled"
		if bucket.UniformBucketLevelAccess.Enabled {
			status = "Enabled"
		}
		configTable.AddRow([]string{"Uniform Bucket-Level Access", status})
	}
	configTable.AddRow([]string{"Public Access Prevention", strings.ToTitle(bucket.PublicAccessPrevention)})

	if bucket.Logging != nil {
		configTable.AddRow([]string{"Usage Logging", fmt.Sprintf("gs://%s/%s", bucket.Logging.LogBucket, bucket.Logging.LogObjectPrefix)})
	} else {
		configTable.AddRow([]string{"Usage Logging", "Not configured"})
	}
	sb.WriteString(configTable.String())
	sb.WriteString("\n\n")

	// Only show ACLs if Uniform Bucket-Level Access is disabled
	if bucket.UniformBucketLevelAccess != nil && !bucket.UniformBucketLevelAccess.Enabled && len(bucket.ACLs) > 0 {
		sb.WriteString("Access Control List (ACLs):\n")
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
