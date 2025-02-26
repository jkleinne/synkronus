package formatter

type StorageFormatter struct{}

func NewStorageFormatter() *StorageFormatter {
	return &StorageFormatter{}
}

func (f *StorageFormatter) FormatBucketList(buckets []string, provider string, details map[string]map[string]string) string {
	table := NewTable([]string{"BUCKET NAME", "PROVIDER", "LOCATION", "USAGE"})

	for _, bucketName := range buckets {
		var location, capacityUsage string
		provider := "GCP" // Default to GCP as we're focusing only on it

		if bucketDetails, exists := details[bucketName]; exists {
			if loc, ok := bucketDetails["Location / Region"]; ok {
				location = loc
			} else {
				location = "N/A"
			}

			if usage, ok := bucketDetails["Usage"]; ok {
				capacityUsage = usage
			} else {
				capacityUsage = "N/A"
			}
		} else {
			location = "N/A"
			capacityUsage = "N/A"
		}

		table.AddRow([]string{bucketName, provider, location, capacityUsage})
	}

	return table.String()
}

func (f *StorageFormatter) FormatBucketDetails(bucketName string, details map[string]string) string {
	var result string

	// Header Section
	result += FormatHeaderSection("Bucket: " + bucketName)
	result += "\n\n"

	// Overview Table
	result += FormatSectionTitle("Overview")
	result += "\n"

	overviewTable := NewTable([]string{"Parameter", "Value"})
	overviewFields := []string{
		"Provider",
		"Location / Region",
		"Capacity",
		"Usage",
		"Created On",
	}

	for _, field := range overviewFields {
		if value, ok := details[field]; ok {
			overviewTable.AddRow([]string{field, value})
		}
	}
	result += overviewTable.String()
	result += "\n\n"

	// Additional Notes (if any)
	if notes, ok := details["Notes"]; ok && notes != "" {
		result += FormatSectionTitle("Additional Notes")
		result += "\n"
		result += "• " + notes + "\n"
	}

	return result
}
