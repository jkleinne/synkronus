package output

import (
	"fmt"
	"strings"
	"time"

	"synkronus/internal/domain"
	"synkronus/internal/domain/storage"
)

const (
	// directoryMarker indicates a directory/prefix in the object list.
	directoryMarker = "(DIR)"
	// timeNotAvailable is used when a timestamp is not available.
	timeNotAvailable = "N/A"
)

// BucketListView renders a slice of buckets as an ASCII table.
type BucketListView []storage.Bucket

// RenderTable returns the bucket list formatted as an ASCII table.
func (v BucketListView) RenderTable() string {
	table := NewTable([]string{"BUCKET NAME", "PROVIDER", "LOCATION", "USAGE", "STORAGE CLASS", "CREATED"})

	for _, bucket := range v {
		createdAt := timeNotAvailable
		if !bucket.CreatedAt.IsZero() {
			createdAt = bucket.CreatedAt.Format("2006-01-02")
		}
		table.AddRow([]string{
			bucket.Name,
			string(bucket.Provider),
			bucket.Location,
			storage.FormatBytes(bucket.UsageBytes),
			bucket.StorageClass,
			createdAt,
		})
	}

	return table.String()
}

// BucketDetailView renders a single bucket's full detail as an ASCII table.
type BucketDetailView struct{ storage.Bucket }

// RenderTable returns the bucket detail formatted as sectioned ASCII tables.
func (v BucketDetailView) RenderTable() string {
	var sb strings.Builder

	sb.WriteString(FormatHeaderSection("Bucket: " + v.Name))
	sb.WriteString("\n\n")

	sb.WriteString(v.renderOverview())
	sb.WriteString(v.renderAccessControl())
	sb.WriteString(v.renderDataProtection())
	sb.WriteString(v.renderLifecycle())
	sb.WriteString(v.renderLabels())

	return sb.String()
}

func (v BucketDetailView) renderOverview() string {
	var sb strings.Builder

	sb.WriteString(FormatSectionTitle("Overview"))
	sb.WriteString("\n")

	table := NewTable([]string{"Parameter", "Value"})
	table.AddRow([]string{"Provider", string(v.Provider)})
	table.AddRow([]string{"Location", v.Location})
	if v.Provider == domain.GCP && v.LocationType != "" {
		table.AddRow([]string{"Location Type", v.LocationType})
	}
	table.AddRow([]string{"Default Storage Class", v.StorageClass})
	table.AddRow([]string{"Usage (Total Bytes)", storage.FormatBytes(v.UsageBytes)})

	if v.Provider == domain.GCP && v.Autoclass != nil {
		autoclassStatus := "Disabled"
		if v.Autoclass.Enabled {
			autoclassStatus = "Enabled"
		}
		table.AddRow([]string{"Autoclass", autoclassStatus})
	}

	if v.Provider == domain.GCP {
		requesterPaysStatus := "Disabled"
		if v.RequesterPays {
			requesterPaysStatus = "Enabled"
		}
		table.AddRow([]string{"Requester Pays", requesterPaysStatus})
	}

	createdAtStr := timeNotAvailable
	if !v.CreatedAt.IsZero() {
		createdAtStr = v.CreatedAt.Format(time.RFC1123)
	}
	table.AddRow([]string{"Created On", createdAtStr})

	updatedAtStr := timeNotAvailable
	if !v.UpdatedAt.IsZero() {
		updatedAtStr = v.UpdatedAt.Format(time.RFC1123)
	}
	table.AddRow([]string{"Updated On", updatedAtStr})

	sb.WriteString(table.String())
	sb.WriteString("\n\n")

	return sb.String()
}

func (v BucketDetailView) renderAccessControl() string {
	var sb strings.Builder

	sb.WriteString(FormatSectionTitle("Access Control & Logging"))
	sb.WriteString("\n")

	configTable := NewTable([]string{"Configuration", "Status"})

	isUBLAEnabled := v.UniformBucketLevelAccess != nil && v.UniformBucketLevelAccess.Enabled

	if v.UniformBucketLevelAccess != nil {
		status := "Disabled (Fine-grained via ACLs/IAM)"
		if isUBLAEnabled {
			status = "Enabled (Uniform via IAM)"
		}
		configTable.AddRow([]string{"Uniform Bucket-Level Access", status})
	}
	configTable.AddRow([]string{"Public Access Prevention", strings.ToTitle(v.PublicAccessPrevention)})

	if v.Logging != nil {
		prefix := ""
		if v.Provider == domain.GCP {
			prefix = "gs://"
		} else if v.Provider == domain.AWS {
			prefix = "s3://"
		}
		configTable.AddRow([]string{"Usage Logging", fmt.Sprintf("%s%s/%s", prefix, v.Logging.LogBucket, v.Logging.LogObjectPrefix)})
	} else {
		configTable.AddRow([]string{"Usage Logging", "Not configured"})
	}
	sb.WriteString(configTable.String())
	sb.WriteString("\n\n")

	sb.WriteString(v.renderIAMPolicy())
	sb.WriteString(v.renderACLs(isUBLAEnabled))

	return sb.String()
}

func (v BucketDetailView) renderIAMPolicy() string {
	var sb strings.Builder

	sb.WriteString("Identity and Access Management (IAM) Policy:\n")

	if v.IAMPolicy == nil {
		sb.WriteString("  (Could not retrieve IAM policy - check permissions)\n\n")
		return sb.String()
	}

	if len(v.IAMPolicy.Bindings) == 0 && len(v.IAMPolicy.Statements) == 0 {
		sb.WriteString("  (No IAM bindings found)\n\n")
		return sb.String()
	}

	// GCP: role → principals bindings
	if len(v.IAMPolicy.Bindings) > 0 {
		iamTable := NewTable([]string{"Role", "Principal(s)"})

		for _, binding := range v.IAMPolicy.Bindings {
			if len(binding.Principals) == 0 {
				continue
			}

			iamTable.AddRow([]string{binding.Role, binding.Principals[0]})

			for i := 1; i < len(binding.Principals); i++ {
				iamTable.AddRow([]string{"", binding.Principals[i]})
			}
		}

		sb.WriteString(iamTable.String())
		sb.WriteString("\n")

		// Render condition annotations after the table
		hasConditions := false
		for _, binding := range v.IAMPolicy.Bindings {
			if binding.Condition != nil {
				if !hasConditions {
					sb.WriteString("Conditions:\n")
					hasConditions = true
				}
				firstPrincipal := ""
				if len(binding.Principals) > 0 {
					firstPrincipal = fmt.Sprintf(" (%s)", binding.Principals[0])
				}
				sb.WriteString(fmt.Sprintf("  %s%s — %s\n", binding.Role, firstPrincipal, binding.Condition.Title))
				if binding.Condition.Description != "" {
					sb.WriteString(fmt.Sprintf("    %s\n", binding.Condition.Description))
				}
				sb.WriteString(fmt.Sprintf("    %s\n", binding.Condition.Expression))
			}
		}
		if hasConditions {
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	// AWS: bucket policy statements
	if len(v.IAMPolicy.Statements) > 0 {
		sb.WriteString("Bucket Policy Statements:\n")
		stmtTable := NewTable([]string{"Effect", "Principal(s)", "Action(s)", "Resource(s)"})
		condCount := 0
		for _, stmt := range v.IAMPolicy.Statements {
			stmtTable.AddRow([]string{
				stmt.Effect,
				strings.Join(stmt.Principals, ", "),
				strings.Join(stmt.Actions, ", "),
				strings.Join(stmt.Resources, ", "),
			})
			condCount += len(stmt.Conditions)
		}
		sb.WriteString(stmtTable.String())
		sb.WriteString("\n")
		if condCount > 0 {
			sb.WriteString(fmt.Sprintf("Note: %d condition(s) present — use --output json for full details.\n", condCount))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func (v BucketDetailView) renderACLs(isUBLAEnabled bool) string {
	var sb strings.Builder

	if isUBLAEnabled {
		sb.WriteString("Access Control List (ACLs):\n")
		sb.WriteString("  (Inactive because Uniform Bucket-Level Access is enabled. IAM policies control access.)\n\n")
		return sb.String()
	}

	if len(v.ACLs) > 0 {
		sb.WriteString("Access Control List (ACLs) - (Fine-grained object control):\n")
		aclTable := NewTable([]string{"Entity", "Role"})
		for _, acl := range v.ACLs {
			aclTable.AddRow([]string{acl.Entity, acl.Role})
		}
		sb.WriteString(aclTable.String())
		sb.WriteString("\n\n")
	}

	return sb.String()
}

func (v BucketDetailView) renderDataProtection() string {
	var sb strings.Builder

	sb.WriteString(FormatSectionTitle("Data Protection"))
	sb.WriteString("\n")

	table := NewTable([]string{"Feature", "Configuration"})

	if v.Encryption != nil && v.Encryption.KmsKeyName != "" {
		table.AddRow([]string{"Encryption (CMEK)", v.Encryption.KmsKeyName})
	} else {
		defaultEncryption := "Provider-managed"
		if v.Provider == domain.GCP {
			defaultEncryption = "Google-managed"
		}
		table.AddRow([]string{"Encryption (CMEK)", defaultEncryption})
	}

	if v.Versioning != nil {
		status := "Suspended"
		if v.Versioning.Enabled {
			status = "Enabled"
		}
		table.AddRow([]string{"Object Versioning", status})
	}

	if v.SoftDeletePolicy != nil {
		table.AddRow([]string{"Soft Delete Policy", fmt.Sprintf("Enabled (Retention: %v)", v.SoftDeletePolicy.RetentionDuration)})
	} else {
		table.AddRow([]string{"Soft Delete Policy", "Disabled"})
	}

	if v.RetentionPolicy != nil {
		lockedStatus := ""
		if v.RetentionPolicy.IsLocked {
			lockedStatus = ", Locked"
		}
		table.AddRow([]string{"Retention Policy", fmt.Sprintf("Enabled (Period: %v%s)", v.RetentionPolicy.RetentionPeriod, lockedStatus)})
	} else {
		table.AddRow([]string{"Retention Policy", "Disabled"})
	}

	sb.WriteString(table.String())
	sb.WriteString("\n\n")

	return sb.String()
}

func (v BucketDetailView) renderLifecycle() string {
	if len(v.LifecycleRules) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(FormatSectionTitle("Lifecycle Rules"))
	sb.WriteString("\n")

	table := NewTable([]string{"Action", "Condition"})
	for _, rule := range v.LifecycleRules {
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
		if rule.Condition.Prefix != "" {
			conditions = append(conditions, fmt.Sprintf("Prefix = %s", rule.Condition.Prefix))
		}
		table.AddRow([]string{rule.Action, strings.Join(conditions, " AND ")})
	}

	sb.WriteString(table.String())
	sb.WriteString("\n\n")

	return sb.String()
}

func (v BucketDetailView) renderLabels() string {
	if len(v.Labels) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(FormatSectionTitle("Labels"))
	sb.WriteString("\n")
	table := NewTable([]string{"Key", "Value"})
	for k, val := range v.Labels {
		table.AddRow([]string{k, val})
	}
	sb.WriteString(table.String())
	sb.WriteString("\n\n")

	return sb.String()
}

// ObjectListView renders an object listing as an ASCII table.
type ObjectListView struct{ storage.ObjectList }

// RenderTable returns the object list formatted as an ASCII table.
func (v ObjectListView) RenderTable() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Listing objects in bucket: %s\n", v.BucketName))
	if v.Prefix != "" {
		sb.WriteString(fmt.Sprintf("Prefix: %s\n", v.Prefix))
	}
	sb.WriteString("\n")

	if len(v.Objects) == 0 && len(v.CommonPrefixes) == 0 {
		sb.WriteString("No objects or directories found.\n")
		return sb.String()
	}

	table := NewTable([]string{"KEY", "SIZE", "STORAGE CLASS", "LAST MODIFIED"})

	for _, prefix := range v.CommonPrefixes {
		table.AddRow([]string{prefix, directoryMarker, "", ""})
	}

	for _, obj := range v.Objects {
		lastMod := timeNotAvailable
		if !obj.LastModified.IsZero() {
			lastMod = obj.LastModified.Format(time.RFC3339)
		}
		table.AddRow([]string{
			obj.Key,
			storage.FormatBytes(obj.Size),
			obj.StorageClass,
			lastMod,
		})
	}

	sb.WriteString(table.String())
	return sb.String()
}

// ObjectDetailView renders a single object's full detail as an ASCII table.
type ObjectDetailView struct{ storage.Object }

// RenderTable returns the object detail formatted as sectioned ASCII tables.
func (v ObjectDetailView) RenderTable() string {
	var sb strings.Builder

	sb.WriteString(FormatHeaderSection("Object: " + v.Key))
	sb.WriteString("\n\n")

	sb.WriteString(v.renderOverview())
	sb.WriteString(v.renderHTTPHeaders())
	sb.WriteString(v.renderMetadata())

	return sb.String()
}

func (v ObjectDetailView) renderOverview() string {
	var sb strings.Builder

	sb.WriteString(FormatSectionTitle("Overview"))
	sb.WriteString("\n")

	table := NewTable([]string{"Parameter", "Value"})
	table.AddRow([]string{"Bucket", v.Bucket})
	table.AddRow([]string{"Provider", string(v.Provider)})
	table.AddRow([]string{"Size", storage.FormatBytes(v.Size)})
	table.AddRow([]string{"Storage Class", v.StorageClass})

	lastModifiedStr := timeNotAvailable
	if !v.LastModified.IsZero() {
		lastModifiedStr = v.LastModified.Format(time.RFC1123)
	}
	table.AddRow([]string{"Last Modified", lastModifiedStr})

	createdAtStr := timeNotAvailable
	if !v.CreatedAt.IsZero() {
		createdAtStr = v.CreatedAt.Format(time.RFC1123)
	}
	table.AddRow([]string{"Created At", createdAtStr})

	if v.Encryption != nil {
		encryptionDetails := v.Encryption.KmsKeyName
		if v.Encryption.Algorithm != "" {
			encryptionDetails = fmt.Sprintf("%s (%s)", encryptionDetails, v.Encryption.Algorithm)
		}
		table.AddRow([]string{"Encryption", encryptionDetails})
	} else {
		table.AddRow([]string{"Encryption", "N/A"})
	}

	table.AddRow([]string{"ETag", v.ETag})
	if v.MD5Hash != "" {
		table.AddRow([]string{"MD5 Hash (Base64)", v.MD5Hash})
	}
	if v.Provider == domain.GCP {
		if v.CRC32C != "" {
			table.AddRow([]string{"CRC32C (Base64)", v.CRC32C})
		}
		table.AddRow([]string{"Generation", fmt.Sprintf("%d", v.Generation)})
		table.AddRow([]string{"Metageneration", fmt.Sprintf("%d", v.Metageneration)})
	}
	if v.Provider == domain.AWS && v.VersionID != "" {
		table.AddRow([]string{"Version ID", v.VersionID})
	}

	sb.WriteString(table.String())
	sb.WriteString("\n\n")

	return sb.String()
}

func (v ObjectDetailView) renderHTTPHeaders() string {
	if v.ContentType == "" && v.ContentEncoding == "" && v.ContentLanguage == "" &&
		v.CacheControl == "" && v.ContentDisposition == "" {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(FormatSectionTitle("HTTP Headers"))
	sb.WriteString("\n")

	table := NewTable([]string{"Header", "Value"})

	if v.ContentType != "" {
		table.AddRow([]string{"Content-Type", v.ContentType})
	}
	if v.ContentEncoding != "" {
		table.AddRow([]string{"Content-Encoding", v.ContentEncoding})
	}
	if v.ContentLanguage != "" {
		table.AddRow([]string{"Content-Language", v.ContentLanguage})
	}
	if v.CacheControl != "" {
		table.AddRow([]string{"Cache-Control", v.CacheControl})
	}
	if v.ContentDisposition != "" {
		table.AddRow([]string{"Content-Disposition", v.ContentDisposition})
	}

	sb.WriteString(table.String())
	sb.WriteString("\n\n")

	return sb.String()
}

func (v ObjectDetailView) renderMetadata() string {
	if len(v.Metadata) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(FormatSectionTitle("User-Defined Metadata"))
	sb.WriteString("\n")

	table := NewTable([]string{"Key", "Value"})
	for k, val := range v.Metadata {
		table.AddRow([]string{k, val})
	}

	sb.WriteString(table.String())
	sb.WriteString("\n\n")

	return sb.String()
}
