// File: internal/tui/ui/storage.go
package ui

import (
	"fmt"
	"strings"
	"time"

	"synkronus/internal/domain/storage"
)

// RenderBucketList renders the bucket table with cursor and scroll.
// Returns a centered "No buckets found" message when buckets is nil or empty.
func RenderBucketList(buckets []storage.Bucket, cursor, offset, termWidth int) string {
	if len(buckets) == 0 {
		msg := TextDimStyle.Render("No buckets found")
		return CenterContent(msg, termWidth)
	}

	headers := []string{"Name", "Provider", "Location", "Usage", "Class", "Created"}
	rows := make([][]string, len(buckets))
	for i, b := range buckets {
		rows[i] = []string{
			b.Name,
			string(b.Provider),
			b.Location,
			storage.FormatBytes(b.UsageBytes),
			b.StorageClass,
			b.CreatedAt.Format("2006-01-02"),
		}
	}

	return RenderTable(headers, rows, cursor, offset, termWidth)
}

// RenderBucketDetail renders bucket metadata in key-value sections.
func RenderBucketDetail(bucket storage.Bucket, termWidth int) string {
	sections := []KeyValueSection{
		buildBucketOverviewSection(bucket),
		buildBucketAccessControlSection(bucket),
		buildBucketDataProtectionSection(bucket),
	}
	return RenderKeyValueGrid(sections, termWidth)
}

func buildBucketOverviewSection(b storage.Bucket) KeyValueSection {
	entries := []KeyValue{
		{Key: "Name", Value: b.Name},
		{Key: "Provider", Value: string(b.Provider), Style: ValueProvider},
		{Key: "Location", Value: b.Location},
		{Key: "Storage Class", Value: b.StorageClass},
		{Key: "Usage", Value: storage.FormatBytes(b.UsageBytes)},
		{Key: "Created", Value: formatTimestamp(b.CreatedAt)},
		{Key: "Updated", Value: formatTimestamp(b.UpdatedAt)},
	}

	if b.Autoclass != nil {
		val, style := FormatBoolValue(b.Autoclass.Enabled)
		entries = append(entries, KeyValue{Key: "Autoclass", Value: val, Style: style})
	}

	requesterPaysVal, requesterPaysStyle := FormatBoolValue(b.RequesterPays)
	entries = append(entries, KeyValue{Key: "Requester Pays", Value: requesterPaysVal, Style: requesterPaysStyle})

	return KeyValueSection{Title: "Overview", Entries: entries}
}

func buildBucketAccessControlSection(b storage.Bucket) KeyValueSection {
	var entries []KeyValue

	if b.UniformBucketLevelAccess != nil {
		val, style := FormatBoolValue(b.UniformBucketLevelAccess.Enabled)
		entries = append(entries, KeyValue{Key: "Uniform Bucket-Level Access", Value: val, Style: style})
	}

	papValue := b.PublicAccessPrevention
	if papValue == "" {
		papValue = "unspecified"
	}
	entries = append(entries, KeyValue{Key: "Public Access Prevention", Value: papValue})

	return KeyValueSection{Title: "Access Control", Entries: entries}
}

func buildBucketDataProtectionSection(b storage.Bucket) KeyValueSection {
	var entries []KeyValue

	encryptionValue := "provider-managed"
	if b.Encryption != nil && b.Encryption.KmsKeyName != "" {
		encryptionValue = b.Encryption.KmsKeyName
		if b.Encryption.Algorithm != "" {
			encryptionValue = fmt.Sprintf("%s (%s)", encryptionValue, b.Encryption.Algorithm)
		}
	}
	entries = append(entries, KeyValue{Key: "Encryption", Value: encryptionValue})

	if b.Versioning != nil {
		val, style := FormatBoolValue(b.Versioning.Enabled)
		entries = append(entries, KeyValue{Key: "Versioning", Value: val, Style: style})
	}

	if b.SoftDeletePolicy != nil {
		entries = append(entries, KeyValue{
			Key:   "Soft Delete",
			Value: fmt.Sprintf("enabled (retention: %v)", b.SoftDeletePolicy.RetentionDuration),
			Style: ValueEnabled,
		})
	} else {
		entries = append(entries, KeyValue{Key: "Soft Delete", Value: "disabled", Style: ValueDisabled})
	}

	if b.RetentionPolicy != nil {
		locked := ""
		if b.RetentionPolicy.IsLocked {
			locked = ", locked"
		}
		entries = append(entries, KeyValue{
			Key:   "Retention Policy",
			Value: fmt.Sprintf("enabled (period: %v%s)", b.RetentionPolicy.RetentionPeriod, locked),
			Style: ValueEnabled,
		})
	} else {
		entries = append(entries, KeyValue{Key: "Retention Policy", Value: "disabled", Style: ValueDisabled})
	}

	return KeyValueSection{Title: "Data Protection", Entries: entries}
}

// RenderObjectList renders the object table within a bucket.
// Common prefixes (directories) appear as "(DIR)" entries at the top.
// Returns a centered "No objects found" message when the list is empty.
func RenderObjectList(objects storage.ObjectList, cursor, offset, termWidth int) string {
	if len(objects.Objects) == 0 && len(objects.CommonPrefixes) == 0 {
		msg := TextDimStyle.Render("No objects found")
		return CenterContent(msg, termWidth)
	}

	headers := []string{"Key", "Size", "Class", "Modified"}
	var rows [][]string

	// Directories first
	for _, prefix := range objects.CommonPrefixes {
		rows = append(rows, []string{prefix, directoryEntry, "", ""})
	}

	for _, obj := range objects.Objects {
		rows = append(rows, []string{
			obj.Key,
			storage.FormatBytes(obj.Size),
			obj.StorageClass,
			formatTimestamp(obj.LastModified),
		})
	}

	return RenderTable(headers, rows, cursor, offset, termWidth)
}

// RenderObjectDetail renders object metadata in key-value sections.
func RenderObjectDetail(object storage.Object, termWidth int) string {
	sections := []KeyValueSection{
		buildObjectOverviewSection(object),
		buildObjectHTTPHeadersSection(object),
	}

	if metaSection, hasMetadata := buildObjectMetadataSection(object); hasMetadata {
		sections = append(sections, metaSection)
	}

	return RenderKeyValueGrid(sections, termWidth)
}

func buildObjectOverviewSection(obj storage.Object) KeyValueSection {
	createdAt := "N/A"
	if !obj.CreatedAt.IsZero() {
		createdAt = formatTimestamp(obj.CreatedAt)
	}

	encryptionValue := "N/A"
	if obj.Encryption != nil {
		encryptionValue = obj.Encryption.KmsKeyName
		if obj.Encryption.Algorithm != "" {
			encryptionValue = fmt.Sprintf("%s (%s)", encryptionValue, obj.Encryption.Algorithm)
		}
	}

	entries := []KeyValue{
		{Key: "Key", Value: obj.Key},
		{Key: "Bucket", Value: obj.Bucket},
		{Key: "Provider", Value: string(obj.Provider), Style: ValueProvider},
		{Key: "Size", Value: storage.FormatBytes(obj.Size)},
		{Key: "Storage Class", Value: obj.StorageClass},
		{Key: "Last Modified", Value: formatTimestamp(obj.LastModified)},
		{Key: "Created", Value: createdAt},
		{Key: "ETag", Value: obj.ETag},
		{Key: "Encryption", Value: encryptionValue},
	}

	return KeyValueSection{Title: "Overview", Entries: entries}
}

func buildObjectHTTPHeadersSection(obj storage.Object) KeyValueSection {
	var entries []KeyValue

	if obj.ContentType != "" {
		entries = append(entries, KeyValue{Key: "Content-Type", Value: obj.ContentType})
	}
	if obj.ContentEncoding != "" {
		entries = append(entries, KeyValue{Key: "Content-Encoding", Value: obj.ContentEncoding})
	}
	if obj.ContentLanguage != "" {
		entries = append(entries, KeyValue{Key: "Content-Language", Value: obj.ContentLanguage})
	}
	if obj.CacheControl != "" {
		entries = append(entries, KeyValue{Key: "Cache-Control", Value: obj.CacheControl})
	}
	if obj.ContentDisposition != "" {
		entries = append(entries, KeyValue{Key: "Content-Disposition", Value: obj.ContentDisposition})
	}

	return KeyValueSection{Title: "HTTP Headers", Entries: entries}
}

func buildObjectMetadataSection(obj storage.Object) (KeyValueSection, bool) {
	if len(obj.Metadata) == 0 {
		return KeyValueSection{}, false
	}

	entries := make([]KeyValue, 0, len(obj.Metadata))
	for k, v := range obj.Metadata {
		entries = append(entries, KeyValue{Key: k, Value: v})
	}

	return KeyValueSection{Title: "Metadata", Entries: entries}, true
}

// CreateBucketFormFields holds the current values for the create bucket form.
type CreateBucketFormFields struct {
	Name                   string
	Provider               string
	Location               string
	StorageClass           string
	Labels                 string
	Versioning             string
	UniformAccess          string
	PublicAccessPrevention string
	SelectorFields         map[int]bool // field indices that use left/right cycling instead of free text
	HiddenFields           map[int]bool // field indices to hide (e.g., UBLA for non-GCP providers)
}

// RenderCreateBucketForm renders the create bucket form fields.
// The active field is highlighted with SectionHeaderStyle; textInputView is shown beside it.
func RenderCreateBucketForm(fields CreateBucketFormFields, activeField int, textInputView string) string {
	type entry struct {
		label string
		value string
		hint  string
	}
	entries := []entry{
		{"Name", fields.Name, ""},
		{"Provider", fields.Provider, "(◀/▶ to select)"},
		{"Location", fields.Location, ""},
		{"Storage Class", fields.StorageClass, "(◀/▶ to select)"},
		{"Labels", fields.Labels, "(key=value,key=value)"},
		{"Versioning", fields.Versioning, "(◀/▶ to select)"},
		{"Uniform Access", fields.UniformAccess, "(◀/▶ to select)"},
		{"Public Access Prevention", fields.PublicAccessPrevention, "(◀/▶ to select)"},
	}

	var lines []string
	for i, e := range entries {
		if fields.HiddenFields[i] {
			continue
		}
		if i == activeField {
			labelStr := SectionHeaderStyle.Render(e.label + ":")
			if fields.SelectorFields[i] {
				display := e.value
				if display == "" {
					display = "(unset)"
				}
				selector := SectionHeaderStyle.Render("◀ " + display + " ▶")
				lines = append(lines, labelStr+" "+selector)
			} else {
				lines = append(lines, labelStr+" "+textInputView)
			}
		} else {
			labelStr := TextDimStyle.Render(e.label + ":")
			val := e.value
			if val == "" && e.hint != "" {
				val = e.hint
			}
			valueStr := TextSecondaryStyle.Render(val)
			lines = append(lines, labelStr+" "+valueStr)
		}
	}

	return strings.Join(lines, "\n")
}

// RenderDeleteConfirm renders the typed-confirmation delete modal content.
// resourceName is the bucket name or object key the user must type to confirm.
func RenderDeleteConfirm(resourceName, currentInput, textInputView string) string {
	warning := TextSecondaryStyle.Render("Type ") +
		SectionHeaderStyle.Render("'"+resourceName+"'") +
		TextSecondaryStyle.Render(" to confirm deletion:")
	return warning + "\n" + textInputView
}

// directoryEntry marks a common prefix as a virtual directory row in the object table.
const directoryEntry = "(DIR)"

// formatTimestamp renders a time.Time as a short date-time string.
// Zero times are returned as "N/A".
func formatTimestamp(t time.Time) string {
	if t.IsZero() {
		return "N/A"
	}
	return t.Format("2006-01-02 15:04:05")
}
