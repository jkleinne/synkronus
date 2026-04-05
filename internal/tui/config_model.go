package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"synkronus/internal/config"
	"synkronus/internal/tui/ui"
)

// ConfigModel owns the mutable state and key/message handling for the Config tab.
type ConfigModel struct {
	entries      []ui.ConfigEntry
	cursor       int
	scrollOffset int
	loading      bool
	loaded       bool
	editKey      string
	editValue    string
	isNewEntry   bool
}

// --- Key handlers ---

// HandleListKeys handles keystrokes on the config list view.
func (c *ConfigModel) HandleListKeys(msg tea.KeyMsg, cm *config.ConfigManager, createFieldIndex *int) (ViewUpdate, tea.Cmd) {
	key := msg.String()
	switch key {
	case "j", keyDown:
		if len(c.entries) > 0 {
			c.cursor = min(c.cursor+1, len(c.entries)-1)
		}
	case "k", keyUp:
		if c.cursor > 0 {
			c.cursor--
		}
	case keyEnter:
		if len(c.entries) > 0 {
			entry := c.entries[c.cursor]
			c.editKey = entry.Key
			c.editValue = entry.Value
			c.isNewEntry = false
			return ViewUpdate{
				ViewState:      ptrViewState(ViewConfigEdit),
				ClearErr:       true,
				FocusTextInput: true,
				TextInputValue: ptrString(entry.Value),
			}, nil
		}
	case "a":
		c.editKey = ""
		c.editValue = ""
		c.isNewEntry = true
		*createFieldIndex = 0
		return ViewUpdate{
			Overlay:        ptrOverlay(OverlayConfigAdd),
			FocusTextInput: true,
			TextInputValue: ptrString(""),
		}, nil
	case "d":
		if len(c.entries) > 0 {
			c.editKey = c.entries[c.cursor].Key
			return ViewUpdate{Overlay: ptrOverlay(OverlayConfigDelete)}, nil
		}
	case "r":
		c.loading = true
		return ViewUpdate{ClearErr: true}, fetchConfigCmd(cm)
	case keyTab:
		return ViewUpdate{SwitchTab: ptrTab(TabConfig.Next())}, nil
	case keyShiftTab:
		return ViewUpdate{SwitchTab: ptrTab(TabConfig.Prev())}, nil
	}
	return ViewUpdate{}, nil
}

// HandleEditKeys handles keystrokes on the config edit view.
// Non-Esc keys are forwarded to the root's textInput via ForwardToTextInput.
func (c *ConfigModel) HandleEditKeys(msg tea.KeyMsg) (ViewUpdate, tea.Cmd) {
	key := msg.String()
	if key == keyEsc {
		return ViewUpdate{ViewState: ptrViewState(ViewConfigList), ClearErr: true, ResetTextInput: true}, nil
	}
	return ViewUpdate{ForwardToTextInput: true}, nil
}

// HandleConfigDeleteConfirm is invoked when Enter is pressed on the ConfigDelete overlay.
func (c *ConfigModel) HandleConfigDeleteConfirm(cm *config.ConfigManager) (ViewUpdate, tea.Cmd) {
	c.loading = true
	provider := c.editKey
	if idx := strings.IndexByte(provider, '.'); idx >= 0 {
		provider = provider[:idx]
	}
	return ViewUpdate{Overlay: ptrOverlay(OverlayNone)}, removeProviderCmd(cm, provider)
}

// --- Message handlers ---

// HandleConfigLoaded processes a completed config fetch.
func (c *ConfigModel) HandleConfigLoaded(msg ConfigLoadedMsg) error {
	c.loading = false
	c.loaded = true
	if msg.Err != nil {
		return msg.Err
	}
	c.entries = msg.Entries
	c.cursor = 0
	c.scrollOffset = 0
	return nil
}

// HandleConfigUpdated processes a completed config update.
func (c *ConfigModel) HandleConfigUpdated(msg ConfigUpdatedMsg, cm *config.ConfigManager) (string, tea.Cmd) {
	c.loading = false
	if msg.Err != nil {
		return "", nil
	}
	c.loading = true
	return "Configuration updated", tea.Batch(fetchConfigCmd(cm), clearStatusCmd())
}

// HandleConfigDeleted processes a completed config deletion.
func (c *ConfigModel) HandleConfigDeleted(msg ConfigDeletedMsg, cm *config.ConfigManager) (string, tea.Cmd) {
	c.loading = false
	if msg.Err != nil {
		return "", nil
	}
	c.loading = true
	return "Configuration entry deleted", tea.Batch(fetchConfigCmd(cm), clearStatusCmd())
}

// HandleProviderRemoved processes a completed provider removal.
func (c *ConfigModel) HandleProviderRemoved(msg ProviderRemovedMsg, cm *config.ConfigManager) (string, tea.Cmd) {
	c.loading = false
	if msg.Err != nil {
		return "", nil
	}
	c.cursor = 0
	c.loading = true
	return "Provider removed", tea.Batch(fetchConfigCmd(cm), clearStatusCmd())
}

// --- View rendering ---

// RenderContent returns the rendered view for config-related viewStates.
func (c *ConfigModel) RenderContent(viewState ViewState, textInputView string, spinnerView string, width, height int) string {
	switch viewState {
	case ViewConfigList:
		if c.loading {
			return ui.CenterContent(ui.RenderSpinnerView(spinnerView, "Loading configuration..."), width)
		}
		return ui.RenderConfigList(c.entries, c.cursor, c.scrollOffset, width)

	case ViewConfigEdit:
		return ui.RenderConfigEdit(c.editKey, textInputView, c.isNewEntry, width, height)

	default:
		return ""
	}
}
