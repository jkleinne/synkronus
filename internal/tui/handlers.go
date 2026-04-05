// File: internal/tui/handlers.go
package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// --- View-level key dispatch ---

// handleViewKeys routes key messages based on the current viewState.
// Global keys (quit, help) are checked first.
func (m *Model) handleViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Global keys
	switch key {
	case keyCtrlC, "q":
		return m, tea.Quit
	case "h":
		m.overlay = OverlayHelp
		return m, nil
	}

	// View-specific dispatch to sub-models
	var vu ViewUpdate
	var cmd tea.Cmd
	switch m.viewState {
	case ViewStorageList:
		vu, cmd = m.storage.HandleListKeys(msg, m.deps.StorageService, &m.deps)
	case ViewStorageBucketDetail:
		vu, cmd = m.storage.HandleBucketDetailKeys(msg, m.deps.StorageService)
	case ViewStorageObjectList:
		vu, cmd = m.storage.HandleObjectListKeys(msg, m.deps.StorageService)
	case ViewStorageObjectDetail:
		vu, cmd = m.storage.HandleObjectDetailKeys(msg)
	case ViewSqlList:
		vu, cmd = m.sql.HandleListKeys(msg, m.deps.SqlService, &m.deps)
	case ViewSqlInstanceDetail:
		vu, cmd = m.sql.HandleInstanceDetailKeys(msg)
	case ViewConfigList:
		vu, cmd = m.config.HandleListKeys(msg, m.deps.ConfigManager, &m.storage.createFieldIndex)
	case ViewConfigEdit:
		vu, cmd = m.config.HandleEditKeys(msg)
		if vu.ForwardToTextInput {
			return m.handleConfigEditTextInput(msg)
		}
	}

	m.applyViewUpdate(vu)

	// Handle tab switching requested by a sub-model.
	if vu.SwitchTab != nil {
		return m.switchTab(*vu.SwitchTab)
	}

	return m, cmd
}

// handleConfigEditTextInput handles the config edit view's text-input forwarding.
func (m *Model) handleConfigEditTextInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case keyEsc:
		m.viewState = ViewConfigList
		m.textInput.Reset()
		m.err = nil
		return m, nil
	case keyEnter:
		value := strings.TrimSpace(m.textInput.Value())
		if value == "" {
			return m, nil
		}
		m.config.loading = true
		m.viewState = ViewConfigList
		m.textInput.Reset()
		m.err = nil
		return m, setConfigCmd(m.deps.ConfigManager, m.config.editKey, value)
	default:
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}
}

// --- Tab switching ---

// switchTab changes the active tab and triggers a lazy load if the target tab
// has not been loaded yet.
func (m *Model) switchTab(tab Tab) (tea.Model, tea.Cmd) {
	m.activeTab = tab
	m.err = nil

	switch tab {
	case TabStorage:
		m.viewState = ViewStorageList
		if !m.storage.loaded {
			m.storage.loading = true
			return m, fetchBucketsCmd(m.deps.StorageService, m.deps.Factory)
		}
	case TabSql:
		m.viewState = ViewSqlList
		if !m.sql.loaded {
			m.sql.loading = true
			return m, fetchInstancesCmd(m.deps.SqlService, m.deps.Factory)
		}
	case TabConfig:
		m.viewState = ViewConfigList
		if !m.config.loaded {
			m.config.loading = true
			return m, fetchConfigCmd(m.deps.ConfigManager)
		}
	}

	return m, nil
}

// --- Data message handlers (routing to sub-models) ---

func (m *Model) handleBucketsLoaded(msg BucketsLoadedMsg) (tea.Model, tea.Cmd) {
	m.err = m.storage.HandleBucketsLoaded(msg)
	return m, nil
}

func (m *Model) handleBucketDetailLoaded(msg BucketDetailMsg) (tea.Model, tea.Cmd) {
	m.err = m.storage.HandleBucketDetailLoaded(msg)
	return m, nil
}

func (m *Model) handleObjectsLoaded(msg ObjectsLoadedMsg) (tea.Model, tea.Cmd) {
	m.err = m.storage.HandleObjectsLoaded(msg)
	return m, nil
}

func (m *Model) handleObjectDetailLoaded(msg ObjectDetailMsg) (tea.Model, tea.Cmd) {
	m.err = m.storage.HandleObjectDetailLoaded(msg)
	return m, nil
}

func (m *Model) handleInstancesLoaded(msg InstancesLoadedMsg) (tea.Model, tea.Cmd) {
	m.err = m.sql.HandleInstancesLoaded(msg)
	return m, nil
}

func (m *Model) handleInstanceDetailLoaded(msg InstanceDetailMsg) (tea.Model, tea.Cmd) {
	m.err = m.sql.HandleInstanceDetailLoaded(msg)
	return m, nil
}

func (m *Model) handleConfigLoaded(msg ConfigLoadedMsg) (tea.Model, tea.Cmd) {
	m.err = m.config.HandleConfigLoaded(msg)
	return m, nil
}

// --- Mutation message handlers (routing to sub-models) ---

func (m *Model) handleBucketCreated(msg BucketCreatedMsg) (tea.Model, tea.Cmd) {
	if msg.Err != nil {
		m.storage.loading = false
		m.err = msg.Err
		return m, nil
	}
	statusMsg, cmd := m.storage.HandleBucketCreated(msg, m.deps.StorageService, &m.deps)
	m.statusMessage = statusMsg
	return m, cmd
}

func (m *Model) handleBucketDeleted(msg BucketDeletedMsg) (tea.Model, tea.Cmd) {
	if msg.Err != nil {
		m.storage.loading = false
		m.err = msg.Err
		return m, nil
	}
	statusMsg, viewState, cmd := m.storage.HandleBucketDeleted(msg, m.deps.StorageService, &m.deps)
	m.statusMessage = statusMsg
	m.viewState = viewState
	return m, cmd
}

func (m *Model) handleObjectDownloaded(msg ObjectDownloadedMsg) (tea.Model, tea.Cmd) {
	if msg.Err != nil {
		m.storage.loading = false
		m.storage.downloadingKey = ""
		m.err = msg.Err
		return m, nil
	}
	statusMsg, cmd := m.storage.HandleObjectDownloaded(msg)
	m.statusMessage = statusMsg
	return m, cmd
}

func (m *Model) handleObjectUploaded(msg ObjectUploadedMsg) (tea.Model, tea.Cmd) {
	if msg.Err != nil {
		m.storage.loading = false
		m.err = msg.Err
		return m, nil
	}
	statusMsg, cmd := m.storage.HandleObjectUploaded(msg, m.deps.StorageService)
	m.statusMessage = statusMsg
	return m, cmd
}

func (m *Model) handleObjectDeleted(msg ObjectDeletedMsg) (tea.Model, tea.Cmd) {
	if msg.Err != nil {
		m.storage.loading = false
		m.err = msg.Err
		return m, nil
	}
	statusMsg, cmd := m.storage.HandleObjectDeleted(msg, m.deps.StorageService)
	m.statusMessage = statusMsg
	m.viewState = ViewStorageObjectList
	return m, cmd
}

func (m *Model) handleConfigUpdated(msg ConfigUpdatedMsg) (tea.Model, tea.Cmd) {
	if msg.Err != nil {
		m.config.loading = false
		m.err = msg.Err
		return m, nil
	}
	m.refreshFactoryConfig()
	statusMsg, cmd := m.config.HandleConfigUpdated(msg, m.deps.ConfigManager)
	m.statusMessage = statusMsg
	return m, cmd
}

func (m *Model) handleConfigDeleted(msg ConfigDeletedMsg) (tea.Model, tea.Cmd) {
	if msg.Err != nil {
		m.config.loading = false
		m.err = msg.Err
		return m, nil
	}
	m.refreshFactoryConfig()
	statusMsg, cmd := m.config.HandleConfigDeleted(msg, m.deps.ConfigManager)
	m.statusMessage = statusMsg
	return m, cmd
}

func (m *Model) handleProviderRemoved(msg ProviderRemovedMsg) (tea.Model, tea.Cmd) {
	if msg.Err != nil {
		m.config.loading = false
		m.err = msg.Err
		return m, nil
	}
	m.refreshFactoryConfig()
	statusMsg, cmd := m.config.HandleProviderRemoved(msg, m.deps.ConfigManager)
	m.statusMessage = statusMsg
	// Invalidate cached data for tabs that depend on provider config
	m.storage.loaded = false
	m.sql.loaded = false
	return m, cmd
}

// --- Delegating wrappers for view-level key handlers ---
// These methods preserve the original Model-level API used by tests,
// delegating to the appropriate sub-model and applying the ViewUpdate.

func (m *Model) handleStorageListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	vu, cmd := m.storage.HandleListKeys(msg, m.deps.StorageService, &m.deps)
	m.applyViewUpdate(vu)
	if vu.SwitchTab != nil {
		return m.switchTab(*vu.SwitchTab)
	}
	return m, cmd
}

func (m *Model) handleObjectListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	vu, cmd := m.storage.HandleObjectListKeys(msg, m.deps.StorageService)
	m.applyViewUpdate(vu)
	return m, cmd
}

func (m *Model) handleObjectDetailKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	vu, cmd := m.storage.HandleObjectDetailKeys(msg)
	m.applyViewUpdate(vu)
	return m, cmd
}

// getCreateFieldOptions delegates to the storage sub-model.
func (m *Model) getCreateFieldOptions(field int) []string {
	return m.storage.getCreateFieldOptions(field)
}

// getCreateFieldValue delegates to the storage sub-model.
func (m *Model) getCreateFieldValue(field int) string {
	return m.storage.getCreateFieldValue(field)
}

// setCreateFieldValue delegates to the storage sub-model.
func (m *Model) setCreateFieldValue(field int, value string) {
	m.storage.setCreateFieldValue(field, value)
}

// --- Shared helpers ---

// cycleOption advances or retreats through a list of options based on key direction.
// Left/right cycle; other keys are ignored and the current value is returned.
func cycleOption(options []string, current, key string) string {
	idx := 0
	for i, o := range options {
		if o == current {
			idx = i
			break
		}
	}
	switch key {
	case keyLeft:
		idx = (idx - 1 + len(options)) % len(options)
	case keyRight:
		idx = (idx + 1) % len(options)
	}
	return options[idx]
}

// parseLabels parses a "key=value,key=value" string into a map.
// Pairs without an "=" separator are silently skipped. Returns nil for empty input.
func parseLabels(s string) map[string]string {
	labels := make(map[string]string)
	for _, pair := range strings.Split(s, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		k, v, ok := strings.Cut(pair, "=")
		if ok {
			labels[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}
	if len(labels) == 0 {
		return nil
	}
	return labels
}
