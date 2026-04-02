package output

import "testing"

func TestParseFormat_ValidFormats(t *testing.T) {
	tests := []struct {
		input    string
		expected Format
	}{
		{"table", FormatTable},
		{"json", FormatJSON},
		{"yaml", FormatYAML},
		{"TABLE", FormatTable},
		{"JSON", FormatJSON},
		{"YAML", FormatYAML},
		{"Json", FormatJSON},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseFormat(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestParseFormat_Invalid(t *testing.T) {
	_, err := ParseFormat("xml")
	if err == nil {
		t.Fatal("expected error for unsupported format")
	}
}
