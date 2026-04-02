package storage

import "testing"

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected string
	}{
		{"Negative", -1, "N/A"},
		{"Zero", 0, "0 B"},
		{"OneByte", 1, "1 B"},
		{"SmallBytes", 500, "500 B"},
		{"JustBelowKB", 1023, "1023 B"},
		{"ExactlyOneKB", 1024, "1.0 KB"},
		{"OneAndHalfKB", 1536, "1.5 KB"},
		{"ExactlyOneMB", 1048576, "1.0 MB"},
		{"OneAndHalfMB", 1572864, "1.5 MB"},
		{"ExactlyOneGB", 1073741824, "1.0 GB"},
		{"ExactlyOneTB", 1099511627776, "1.0 TB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatBytes(tt.input)
			if result != tt.expected {
				t.Errorf("FormatBytes(%d) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
