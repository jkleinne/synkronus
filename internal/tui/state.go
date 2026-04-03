package tui

// ViewState represents the primary screen the TUI is showing.
type ViewState int

const (
	ViewStorageList         ViewState = iota // Bucket table (default/home)
	ViewStorageBucketDetail                  // Single bucket detail
	ViewStorageObjectList                    // Objects within a bucket
	ViewStorageObjectDetail                  // Single object detail
	ViewSqlList                              // SQL instances table
	ViewSqlInstanceDetail                    // Single instance detail
	ViewConfigList                           // Config key-value list
	ViewConfigEdit                           // Editing a config value
)

// OverlayState represents a modal overlay rendered on top of the current view.
type OverlayState int

const (
	OverlayNone         OverlayState = iota // No overlay
	OverlayHelp                             // Help screen
	OverlayCreateBucket                     // Create bucket form
	OverlayDeleteConfirm                    // Typed delete confirmation
	OverlayConfigAdd                        // Add new config key-value
	OverlayConfigDelete                     // Delete config key confirmation
	OverlayDownloadPath                     // Download directory input
)

// Tab identifies the active top-level tab.
type Tab int

const (
	TabStorage Tab = iota
	TabSql
	TabConfig
	tabCount // sentinel for wrapping arithmetic
)

// Next returns the next tab, wrapping around.
func (t Tab) Next() Tab {
	return (t + 1) % tabCount
}

// Prev returns the previous tab, wrapping around.
func (t Tab) Prev() Tab {
	return (t - 1 + tabCount) % tabCount
}

// hasTextInput returns true if the overlay captures keyboard input
// for text entry (suppressing single-key bindings).
func (o OverlayState) hasTextInput() bool {
	switch o {
	case OverlayCreateBucket, OverlayDeleteConfirm, OverlayConfigAdd, OverlayDownloadPath:
		return true
	default:
		return false
	}
}
