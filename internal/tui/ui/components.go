// File: internal/tui/ui/components.go
package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Viewport holds terminal dimensions and scroll state for render functions.
type Viewport struct {
	Width  int
	Height int
	Offset int
	Cursor int
}

// ProviderStatus represents a cloud provider's connectivity for the tab bar.
type ProviderStatus struct {
	Name       string
	Configured bool
}

// RenderBanner returns the gradient "SYNKRONUS" title in a rounded border box, centered.
func RenderBanner(termWidth int) string {
	title := GradientText("SYNKRONUS", gradientColors)
	subtitle := TextDimStyle.Render("multi-cloud infrastructure")
	content := lipgloss.JoinVertical(lipgloss.Center, title, subtitle)
	box := BannerBoxStyle.Render(content)
	return lipgloss.PlaceHorizontal(termWidth, lipgloss.Center, box)
}

// RenderTabs renders the tab bar with the active tab underlined.
func RenderTabs(activeTab int, termWidth int, providers []ProviderStatus) string {
	tabNames := []string{"Storage", "SQL", "Config"}
	width := ContentWidth(termWidth)

	var tabs []string
	for i, name := range tabNames {
		if i == activeTab {
			styled := ActiveTabStyle.Render(name)
			underline := TabUnderlineStyle.Render(strings.Repeat("─", lipgloss.Width(name)))
			tabs = append(tabs, lipgloss.JoinVertical(lipgloss.Left, styled, underline))
		} else {
			styled := InactiveTabStyle.Render(name)
			spacer := strings.Repeat(" ", lipgloss.Width(name))
			tabs = append(tabs, lipgloss.JoinVertical(lipgloss.Left, styled, spacer))
		}
	}

	tabBar := lipgloss.JoinHorizontal(lipgloss.Bottom, intersperse(tabs, "  ")...)

	if len(providers) > 0 {
		var dots []string
		for _, p := range providers {
			dot := "○"
			style := TextDimStyle
			if p.Configured {
				dot = "●"
				style = StatusGreenStyle
			}
			dots = append(dots, style.Render(dot)+" "+TextDimStyle.Render(strings.ToUpper(p.Name)))
		}
		providerInfo := strings.Join(dots, "  ")
		gap := width - lipgloss.Width(tabBar) - lipgloss.Width(providerInfo)
		if gap > 0 {
			tabBar = tabBar + strings.Repeat(" ", gap) + providerInfo
		}
	}

	separator := TextDimStyle.Render(strings.Repeat("─", width))
	result := lipgloss.JoinVertical(lipgloss.Left, tabBar, separator)
	return lipgloss.PlaceHorizontal(termWidth, lipgloss.Center, result)
}

// RenderTable renders a scrollable table with headers, rows, and cursor highlight.
func RenderTable(headers []string, rows [][]string, cursor, offset, termWidth int) string {
	width := ContentWidth(termWidth)
	if len(headers) == 0 {
		return ""
	}

	colWidths := make([]int, len(headers))
	for i, h := range headers {
		colWidths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(colWidths) && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	var headerCells []string
	for i, h := range headers {
		headerCells = append(headerCells, TableHeaderStyle.Width(colWidths[i]+2).Render(strings.ToUpper(h)))
	}
	headerLine := lipgloss.JoinHorizontal(lipgloss.Top, headerCells...)

	var renderedRows []string
	for i, row := range rows {
		if i < offset {
			continue
		}
		var cells []string
		isSelected := i == cursor
		for j, cell := range row {
			if j >= len(colWidths) {
				break
			}
			var style lipgloss.Style
			if isSelected {
				style = SelectedRowStyle
			} else if i%2 == 0 {
				style = EvenRowStyle
			} else {
				style = OddRowStyle
			}
			cellContent := cell
			if isSelected && j == 0 {
				cellContent = "▸ " + cellContent
			} else if j == 0 {
				cellContent = "  " + cellContent
			}
			cells = append(cells, style.Width(colWidths[j]+2).Render(cellContent))
		}
		renderedRows = append(renderedRows, lipgloss.JoinHorizontal(lipgloss.Top, cells...))
	}

	all := append([]string{headerLine}, renderedRows...)
	table := lipgloss.JoinVertical(lipgloss.Left, all...)
	box := ContentBoxStyle.Width(width).Render(table)
	return lipgloss.PlaceHorizontal(termWidth, lipgloss.Center, box)
}

// KeyValueSection groups key-value pairs under a section header.
type KeyValueSection struct {
	Title   string
	Entries []KeyValue
}

// KeyValue is a single key-value pair with optional style hint.
type KeyValue struct {
	Key   string
	Value string
	Style ValueStyle
}

// ValueStyle controls how a value is rendered.
type ValueStyle int

const (
	ValueDefault  ValueStyle = iota
	ValueEnabled             // rendered in green
	ValueDisabled            // rendered in red
	ValueProvider            // rendered in purple (provider accent)
)

func formatValue(value string, style ValueStyle) string {
	switch style {
	case ValueEnabled:
		return StatusGreenStyle.Render(value)
	case ValueDisabled:
		return StatusRedStyle.Render(value)
	case ValueProvider:
		return ProviderStyle.Render(value)
	default:
		return TextSecondaryStyle.Render(value)
	}
}

// RenderKeyValueGrid renders sections of key-value pairs in a two-column layout.
func RenderKeyValueGrid(sections []KeyValueSection, termWidth int) string {
	width := ContentWidth(termWidth)
	var parts []string

	for _, section := range sections {
		header := SectionHeaderStyle.Render(section.Title)
		var rows []string
		for _, kv := range section.Entries {
			label := TextDimStyle.Width(20).Render(kv.Key)
			value := formatValue(kv.Value, kv.Style)
			rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, label, value))
		}
		sectionContent := lipgloss.JoinVertical(lipgloss.Left, append([]string{header}, rows...)...)
		parts = append(parts, sectionContent)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, parts...)
	box := ContentBoxStyle.Width(width).Render(content)
	return lipgloss.PlaceHorizontal(termWidth, lipgloss.Center, box)
}

// RenderModal renders a centered modal box with title and content.
func RenderModal(title, content string, termWidth, termHeight int) string {
	width := min(ContentWidth(termWidth), 60)
	titleStr := SectionHeaderStyle.Render(title)
	inner := lipgloss.JoinVertical(lipgloss.Left, titleStr, "", content)
	box := ModalBoxStyle.Width(width).Render(inner)
	return lipgloss.Place(termWidth, termHeight, lipgloss.Center, lipgloss.Center, box)
}

// RenderSpinnerView renders a spinner frame with a label.
func RenderSpinnerView(spinnerFrame, label string) string {
	return SpinnerStyle.Render(spinnerFrame) + " " + TextSecondaryStyle.Render(label)
}

// RenderKeyHints renders the bottom keybinding hint bar for a given context.
func RenderKeyHints(ctx BindingContext, termWidth int) string {
	hints := FormatHints(ctx)
	return lipgloss.PlaceHorizontal(termWidth, lipgloss.Center, hints)
}

// RenderBreadcrumb renders a navigation breadcrumb trail.
func RenderBreadcrumb(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	var rendered []string
	for i, part := range parts {
		if i == len(parts)-1 {
			rendered = append(rendered, BreadcrumbActiveStyle.Render(part))
		} else {
			rendered = append(rendered, TextDimStyle.Render(part))
			rendered = append(rendered, BreadcrumbSepStyle.Render(" › "))
		}
	}
	return strings.Join(rendered, "")
}

// RenderError renders an inline error message.
func RenderError(msg string, termWidth int) string {
	content := ErrorStyle.Render("Error: " + msg)
	retry := TextDimStyle.Render("  Press r to retry")
	return lipgloss.PlaceHorizontal(termWidth, lipgloss.Center, content+retry)
}

// RenderStatusMessage renders a transient status message.
func RenderStatusMessage(msg string, termWidth int) string {
	return lipgloss.PlaceHorizontal(termWidth, lipgloss.Center, StatusGreenStyle.Render(msg))
}

// RenderHelpContent renders a full help screen listing keybindings grouped by context.
func RenderHelpContent(viewState int, activeTab int) string {
	contexts := []BindingContext{
		ContextStorageList, ContextBucketDetail, ContextObjectList,
		ContextSqlList, ContextInstanceDetail,
		ContextConfigList, ContextConfigEdit,
		ContextModal, ContextCreateForm,
	}
	var sections []string
	for _, ctx := range contexts {
		ctxBindings := BindingsForContext(ctx)
		if len(ctxBindings) == 0 {
			continue
		}
		header := SectionHeaderStyle.Render(string(ctx))
		var lines []string
		for _, b := range ctxBindings {
			lines = append(lines, fmt.Sprintf("  %s  %s", HintKeyStyle.Width(12).Render(b.Key), HintDescStyle.Render(b.Description)))
		}
		sections = append(sections, header+"\n"+strings.Join(lines, "\n"))
	}
	return strings.Join(sections, "\n\n")
}

// CenterContent centers a rendered string horizontally within the terminal width.
func CenterContent(content string, termWidth int) string {
	return lipgloss.PlaceHorizontal(termWidth, lipgloss.Center, content)
}

// FormatBool renders a boolean as "enabled"/"disabled" with appropriate color.
func FormatBool(b bool) string {
	if b {
		return StatusGreenStyle.Render("enabled")
	}
	return StatusRedStyle.Render("disabled")
}

// FormatBoolValue converts a boolean to a KeyValue-compatible string + style.
func FormatBoolValue(b bool) (string, ValueStyle) {
	if b {
		return "enabled", ValueEnabled
	}
	return "disabled", ValueDisabled
}

// FormatOptionalString returns the value or a dim "none" placeholder.
func FormatOptionalString(s string) string {
	if s == "" {
		return TextDimStyle.Render("none")
	}
	return s
}

// RenderCount formats a count summary like "5 buckets · 1 provider".
func RenderCount(count int, noun string, extra string) string {
	s := fmt.Sprintf("%d %s", count, noun)
	if extra != "" {
		s += " · " + extra
	}
	return TextDimStyle.Render(s)
}

func intersperse(items []string, sep string) []string {
	if len(items) <= 1 {
		return items
	}
	result := make([]string, 0, len(items)*2-1)
	for i, item := range items {
		result = append(result, item)
		if i < len(items)-1 {
			result = append(result, sep)
		}
	}
	return result
}
