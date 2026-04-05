package output

import "strings"

// headerBorderPadding is the number of extra characters added on each side of the
// title text in FormatHeaderSection, so the border extends visibly beyond the text.
const headerBorderPadding = 30

// Table renders data as a bordered ASCII table.
type Table struct {
	Headers      []string
	Rows         [][]string
	columnWidths []int
}

// NewTable creates a new table with the given headers.
func NewTable(headers []string) *Table {
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	return &Table{
		Headers:      headers,
		Rows:         [][]string{},
		columnWidths: widths,
	}
}

// AddRow appends a row and incrementally updates column widths.
func (t *Table) AddRow(row []string) {
	t.Rows = append(t.Rows, row)
	for i, cell := range row {
		if i < len(t.columnWidths) && len(cell) > t.columnWidths[i] {
			t.columnWidths[i] = len(cell)
		}
	}
}

// String returns the rendered table as a string.
func (t *Table) String() string {
	if len(t.Headers) == 0 {
		return ""
	}

	var sb strings.Builder

	t.writeBorder(&sb)
	sb.WriteString("\n")

	sb.WriteString("| ")
	for i, h := range t.Headers {
		sb.WriteString(h)
		sb.WriteString(strings.Repeat(" ", t.columnWidths[i]-len(h)))
		sb.WriteString(" | ")
	}
	sb.WriteString("\n")

	t.writeBorder(&sb)
	sb.WriteString("\n")

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

	t.writeBorder(&sb)
	return sb.String()
}

func (t *Table) writeBorder(sb *strings.Builder) {
	sb.WriteString("+")
	for _, width := range t.columnWidths {
		sb.WriteString(strings.Repeat("-", width+2))
		sb.WriteString("+")
	}
}

// FormatHeaderSection formats a prominent section header with a title.
func FormatHeaderSection(title string) string {
	var sb strings.Builder
	borderLine := strings.Repeat("=", len(title)+headerBorderPadding)
	sb.WriteString(borderLine)
	sb.WriteString("\n")
	sb.WriteString("  " + title + "  ")
	sb.WriteString("\n")
	sb.WriteString(borderLine)
	return sb.String()
}

// FormatSectionTitle formats a simple section title.
func FormatSectionTitle(title string) string {
	return "-- " + title + " --"
}
