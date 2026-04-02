// File: internal/tui/ui/config.go
package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// ConfigEntry represents a flattened configuration key-value pair.
type ConfigEntry struct {
	Key   string
	Value string
}

// RenderConfigList renders the config key-value table with cursor and scroll.
// Returns a centered "No configuration entries" message when entries is nil or empty.
func RenderConfigList(entries []ConfigEntry, cursor, offset, termWidth int) string {
	if len(entries) == 0 {
		msg := TextDimStyle.Render("No configuration entries")
		return lipgloss.PlaceHorizontal(termWidth, lipgloss.Center, msg)
	}

	headers := []string{"Key", "Value"}
	rows := make([][]string, len(entries))
	for i, entry := range entries {
		rows[i] = []string{entry.Key, entry.Value}
	}

	return RenderTable(headers, rows, cursor, offset, termWidth)
}

// RenderConfigEdit renders the config edit form showing the key being edited and
// the current value. When isNew is true, the header reads "Add new entry".
// The value string is rendered as-is since the actual textinput view is composited
// by the caller.
func RenderConfigEdit(key, value string, isNew bool, termWidth, termHeight int) string {
	header := "Edit entry"
	if isNew {
		header = "Add new entry"
	}
	title := SectionHeaderStyle.Render(header)
	keyLabel := TextDimStyle.Render("Key: ") + TextPrimaryStyle.Render(key)
	valueLabel := TextDimStyle.Render("Value: ") + value

	content := lipgloss.JoinVertical(lipgloss.Left, title, "", keyLabel, "", valueLabel)
	return RenderModal(fmt.Sprintf("%s — %s", header, key), content, termWidth, termHeight)
}

// RenderConfigDeleteConfirm returns the confirmation modal content for deleting a
// config entry. The caller is expected to embed this in a modal via RenderModal.
func RenderConfigDeleteConfirm(key string) string {
	return fmt.Sprintf("Delete config entry '%s'? Press Enter to confirm or Esc to cancel.", key)
}
