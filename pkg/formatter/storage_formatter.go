// File: pkg/formatter/storage_formatter.go
package formatter

import (
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
	var result string

	result += FormatHeaderSection("Bucket: " + bucket.Name)
	result += "\n\n"

	result += FormatSectionTitle("Overview")
	result += "\n"

	overviewTable := NewTable([]string{"Parameter", "Value"})

	details := []struct {
		Key   string
		Value string
	}{
		{"Provider", string(bucket.Provider)},
		{"Location / Region", bucket.Location},
		{"Storage Class", bucket.StorageClass},
		{"Usage", storage.FormatBytes(bucket.UsageBytes)},
		// Format time in a standard, detailed format (RFC1123)
		{"Created On", bucket.CreatedAt.Format(time.RFC1123)},
		{"Updated On", bucket.UpdatedAt.Format(time.RFC1123)},
	}

	for _, detail := range details {
		overviewTable.AddRow([]string{detail.Key, detail.Value})
	}

	result += overviewTable.String()
	result += "\n\n"

	if len(bucket.Labels) > 0 {
		result += FormatSectionTitle("Labels")
		result += "\n"
		labelsTable := NewTable([]string{"Key", "Value"})
		for k, v := range bucket.Labels {
			labelsTable.AddRow([]string{k, v})
		}
		result += labelsTable.String()
		result += "\n\n"
	}

	return result
}
