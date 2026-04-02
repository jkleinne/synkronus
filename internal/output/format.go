package output

import (
	"fmt"
	"strings"
)

// Format represents a supported output format for CLI rendering.
type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
	FormatYAML  Format = "yaml"
)

// ParseFormat parses a format string (case-insensitive) and returns the
// corresponding Format constant. Returns an error if the value is not one
// of: table, json, yaml.
func ParseFormat(s string) (Format, error) {
	switch strings.ToLower(s) {
	case "table":
		return FormatTable, nil
	case "json":
		return FormatJSON, nil
	case "yaml":
		return FormatYAML, nil
	default:
		return "", fmt.Errorf("unsupported output format %q: valid formats are table, json, yaml", s)
	}
}
