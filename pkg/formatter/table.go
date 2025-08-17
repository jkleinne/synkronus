// File: pkg/formatter/table.go
package formatter

import (
	"strings"
)

type Table struct {
	Headers      []string
	Rows         [][]string
	columnWidths []int
}

// Creates a new table with the given headers
func NewTable(headers []string) *Table {
	t := &Table{
		Headers: headers,
		Rows:    [][]string{},
	}
	t.calculateColumnWidths()
	return t
}

func (t *Table) AddRow(row []string) {
	t.Rows = append(t.Rows, row)
	t.calculateColumnWidths()
}

func (t *Table) calculateColumnWidths() {
	// Initialize column widths from headers
	t.columnWidths = make([]int, len(t.Headers))
	for i, h := range t.Headers {
		t.columnWidths[i] = len(h)
	}

	// Update column widths from rows
	for _, row := range t.Rows {
		for i, cell := range row {
			if i < len(t.columnWidths) && len(cell) > t.columnWidths[i] {
				t.columnWidths[i] = len(cell)
			}
		}
	}
}

// Returns the string representation of the table
func (t *Table) String() string {
	if len(t.Headers) == 0 {
		return ""
	}

	t.calculateColumnWidths()

	var sb strings.Builder

	// Build the top border
	t.writeBorder(&sb)
	sb.WriteString("\n")

	// Write headers
	sb.WriteString("| ")
	for i, h := range t.Headers {
		sb.WriteString(h)
		sb.WriteString(strings.Repeat(" ", t.columnWidths[i]-len(h)))
		sb.WriteString(" | ")
	}
	sb.WriteString("\n")

	// Build the header-row separator
	t.writeBorder(&sb)
	sb.WriteString("\n")

	// Write rows
	for _, row := range t.Rows {
		sb.WriteString("| ")
		for i, cell := range row {
			if i < len(t.columnWidths) {
				sb.WriteString(cell)
				sb.WriteString(strings.Repeat(" ", t.columnWidths[i]-len(cell)))
				sb.WriteString(" | ")
			}
		}
		sb.WriteString("\n")
	}

	// Build the bottom border
	t.writeBorder(&sb)

	return sb.String()
}

// writeBorder writes a horizontal border to the string builder
func (t *Table) writeBorder(sb *strings.Builder) {
	sb.WriteString("+")
	for _, width := range t.columnWidths {
		sb.WriteString(strings.Repeat("-", width+2))
		sb.WriteString("+")
	}
}

// Formats a section header with a title
func FormatHeaderSection(title string) string {
	var sb strings.Builder

	borderLine := strings.Repeat("=", len(title)+30) // Add extra padding for aesthetics

	sb.WriteString(borderLine)
	sb.WriteString("\n")
	sb.WriteString("  " + title + "  ")
	sb.WriteString("\n")
	sb.WriteString(borderLine)

	return sb.String()
}

// Formats a simple section title
func FormatSectionTitle(title string) string {
	return "-- " + title + " --"
}
