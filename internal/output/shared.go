package output

import "strings"

// renderLabelsSection renders a Labels table section.
// Returns an empty string if labels is empty.
func renderLabelsSection(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(FormatSectionTitle("Labels"))
	sb.WriteString("\n")
	table := NewTable([]string{"Key", "Value"})
	for k, val := range labels {
		table.AddRow([]string{k, val})
	}
	sb.WriteString(table.String())
	sb.WriteString("\n\n")

	return sb.String()
}
