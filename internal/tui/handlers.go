// File: internal/tui/handlers.go
package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"synkronus/internal/domain/storage"
)

// Key constants for tea.KeyMsg.String() comparisons.
const (
	keyCtrlC    = "ctrl+c"
	keyEsc      = "esc"
	keyUp       = "up"
	keyDown     = "down"
	keyEnter    = "enter"
	keyTab      = "tab"
	keyShiftTab = "shift+tab"
	keyLeft     = "left"
	keyRight    = "right"
)

// createFormFieldCount is the number of fields in the create-bucket overlay.
const createFormFieldCount = 8

// --- Overlay key handling ---

// handleOverlayKeys intercepts keystrokes when a modal overlay is active.
// Text-input overlays only intercept Esc/Enter/Tab; everything else goes to textinput.
// Non-text overlays (Help, ConfigDelete) close on Esc, h, or q.
func (m *Model) handleOverlayKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	if m.overlay.hasTextInput() {
		switch key {
		case keyEsc:
			m.overlay = OverlayNone
			m.textInput.Reset()
			return m, nil
		case keyEnter:
			return m.handleOverlaySubmit()
		case keyTab:
			// Tab cycles fields in multi-field overlays (CreateBucket, ConfigAdd).
			m.syncTextInputToField()
			switch m.overlay {
			case OverlayCreateBucket:
				m.storage.createField = (m.storage.createField + 1) % createFormFieldCount
				m.loadFieldIntoTextInput()
			case OverlayConfigAdd:
				// Toggle between key (field 0) and value (field 1).
				if m.storage.createField == 0 {
					m.config.editKey = m.textInput.Value()
					m.storage.createField = 1
					m.textInput.SetValue(m.config.editValue)
				} else {
					m.config.editValue = m.textInput.Value()
					m.storage.createField = 0
					m.textInput.SetValue(m.config.editKey)
				}
				m.textInput.Focus()
			}
			return m, nil
		default:
			// Provider field (1): cycle with left/right instead of free text.
			if m.overlay == OverlayCreateBucket && m.storage.createField == 1 && len(m.storage.availableProviders) > 0 {
				m.storage.createProvider = m.cycleProvider(key)
				m.textInput.SetValue(m.storage.createProvider)
				return m, nil
			}
			// Forward all other keys to the textinput bubble.
			var cmd tea.Cmd
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd
		}
	}

	// Non-text-input overlays (Help, ConfigDelete).
	switch key {
	case keyEsc, "h", "q":
		m.overlay = OverlayNone
		return m, nil
	case keyEnter:
		// ConfigDelete confirm.
		if m.overlay == OverlayConfigDelete {
			return m.handleConfigDeleteConfirm()
		}
		m.overlay = OverlayNone
		return m, nil
	}

	return m, nil
}

// handleOverlaySubmit processes Enter on a text-input overlay.
func (m *Model) handleOverlaySubmit() (tea.Model, tea.Cmd) {
	m.syncTextInputToField()

	switch m.overlay {
	case OverlayCreateBucket:
		name := strings.TrimSpace(m.storage.createName)
		provider := strings.TrimSpace(m.storage.createProvider)
		location := strings.TrimSpace(m.storage.createLocation)
		if name == "" || provider == "" || location == "" {
			return m, nil
		}
		opts := storage.CreateBucketOptions{
			Name:     name,
			Location: location,
		}
		if sc := strings.TrimSpace(m.storage.createStorageClass); sc != "" {
			opts.StorageClass = strings.ToUpper(sc)
		}
		if labelsStr := strings.TrimSpace(m.storage.createLabels); labelsStr != "" {
			opts.Labels = parseLabels(labelsStr)
		}
		if v := strings.TrimSpace(strings.ToLower(m.storage.createVersioning)); v == "yes" {
			t := true
			opts.Versioning = &t
		} else if v == "no" {
			f := false
			opts.Versioning = &f
		}
		if v := strings.TrimSpace(strings.ToLower(m.storage.createUniformAccess)); v == "yes" {
			t := true
			opts.UniformBucketLevelAccess = &t
		} else if v == "no" {
			f := false
			opts.UniformBucketLevelAccess = &f
		}
		if v := strings.TrimSpace(strings.ToLower(m.storage.createPublicAccessPrevention)); v == storage.PublicAccessPreventionEnforced || v == storage.PublicAccessPreventionInherited {
			opts.PublicAccessPrevention = &v
		}
		m.overlay = OverlayNone
		m.storage.loading = true
		m.textInput.Reset()
		return m, createBucketCmd(m.storageService, opts, provider)

	case OverlayDeleteConfirm:
		if m.textInput.Value() != m.storage.selectedBucket.Name {
			return m, nil
		}
		m.overlay = OverlayNone
		m.storage.loading = true
		m.textInput.Reset()
		return m, deleteBucketCmd(
			m.storageService,
			m.storage.selectedBucket.Name,
			strings.ToLower(string(m.storage.selectedBucket.Provider)),
		)

	case OverlayConfigAdd:
		key := strings.TrimSpace(m.config.editKey)
		value := strings.TrimSpace(m.config.editValue)
		if key == "" || value == "" {
			return m, nil
		}
		m.overlay = OverlayNone
		m.config.loading = true
		m.textInput.Reset()
		return m, setConfigCmd(m.configManager, key, value)
	}

	return m, nil
}

// syncTextInputToField persists the current textinput value back to the active field.
func (m *Model) syncTextInputToField() {
	switch m.overlay {
	case OverlayCreateBucket:
		switch m.storage.createField {
		case 0:
			m.storage.createName = m.textInput.Value()
		case 1:
			// Provider is a selector field — only sync from textinput if no providers are available (fallback).
			if len(m.storage.availableProviders) == 0 {
				m.storage.createProvider = m.textInput.Value()
			}
		case 2:
			m.storage.createLocation = m.textInput.Value()
		case 3:
			m.storage.createStorageClass = m.textInput.Value()
		case 4:
			m.storage.createLabels = m.textInput.Value()
		case 5:
			m.storage.createVersioning = m.textInput.Value()
		case 6:
			m.storage.createUniformAccess = m.textInput.Value()
		case 7:
			m.storage.createPublicAccessPrevention = m.textInput.Value()
		}
	case OverlayDeleteConfirm:
		m.storage.deleteInput = m.textInput.Value()
	case OverlayConfigAdd:
		if m.storage.createField == 0 {
			m.config.editKey = m.textInput.Value()
		} else {
			m.config.editValue = m.textInput.Value()
		}
	}
}

// loadFieldIntoTextInput sets the textinput value to the current create-bucket field.
func (m *Model) loadFieldIntoTextInput() {
	switch m.storage.createField {
	case 0:
		m.textInput.SetValue(m.storage.createName)
	case 1:
		m.textInput.SetValue(m.storage.createProvider)
	case 2:
		m.textInput.SetValue(m.storage.createLocation)
	case 3:
		m.textInput.SetValue(m.storage.createStorageClass)
	case 4:
		m.textInput.SetValue(m.storage.createLabels)
	case 5:
		m.textInput.SetValue(m.storage.createVersioning)
	case 6:
		m.textInput.SetValue(m.storage.createUniformAccess)
	case 7:
		m.textInput.SetValue(m.storage.createPublicAccessPrevention)
	}
	m.textInput.Focus()
}

// cycleProvider advances or retreats the provider selection based on the key.
// Left/right cycle through available providers; other keys are ignored.
func (m *Model) cycleProvider(key string) string {
	providers := m.storage.availableProviders
	current := m.storage.createProvider
	idx := 0
	for i, p := range providers {
		if p == current {
			idx = i
			break
		}
	}
	switch key {
	case keyLeft:
		idx = (idx - 1 + len(providers)) % len(providers)
	case keyRight:
		idx = (idx + 1) % len(providers)
	}
	return providers[idx]
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

	// View-specific dispatch
	switch m.viewState {
	case ViewStorageList:
		return m.handleStorageListKeys(msg)
	case ViewStorageBucketDetail:
		return m.handleBucketDetailKeys(msg)
	case ViewStorageObjectList:
		return m.handleObjectListKeys(msg)
	case ViewStorageObjectDetail:
		return m.handleObjectDetailKeys(msg)
	case ViewSqlList:
		return m.handleSqlListKeys(msg)
	case ViewSqlInstanceDetail:
		return m.handleInstanceDetailKeys(msg)
	case ViewConfigList:
		return m.handleConfigListKeys(msg)
	case ViewConfigEdit:
		return m.handleConfigEditKeys(msg)
	}

	return m, nil
}

// --- Storage list ---

func (m *Model) handleStorageListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "j", keyDown:
		if len(m.storage.buckets) > 0 {
			m.storage.cursor = min(m.storage.cursor+1, len(m.storage.buckets)-1)
		}
	case "k", keyUp:
		if m.storage.cursor > 0 {
			m.storage.cursor--
		}
	case keyEnter:
		if len(m.storage.buckets) > 0 {
			m.storage.selectedBucket = m.storage.buckets[m.storage.cursor]
			m.viewState = ViewStorageBucketDetail
			m.storage.loading = true
			m.err = nil
			return m, fetchBucketDetailCmd(
				m.storageService,
				m.storage.selectedBucket.Name,
				strings.ToLower(string(m.storage.selectedBucket.Provider)),
			)
		}
	case "o":
		if len(m.storage.buckets) > 0 {
			m.storage.selectedBucket = m.storage.buckets[m.storage.cursor]
			m.viewState = ViewStorageObjectList
			m.storage.loading = true
			m.storage.cursor = 0
			m.storage.scrollOffset = 0
			m.err = nil
			return m, fetchObjectsCmd(
				m.storageService,
				m.storage.selectedBucket.Name,
				strings.ToLower(string(m.storage.selectedBucket.Provider)),
				"",
			)
		}
	case "c":
		m.storage.createName = ""
		m.storage.createLocation = ""
		m.storage.createStorageClass = ""
		m.storage.createLabels = ""
		m.storage.createVersioning = ""
		m.storage.createUniformAccess = ""
		m.storage.createPublicAccessPrevention = ""
		m.storage.availableProviders = m.factory.GetConfiguredProviders()
		if len(m.storage.availableProviders) > 0 {
			m.storage.createProvider = m.storage.availableProviders[0]
		} else {
			m.storage.createProvider = ""
		}
		m.storage.createField = 0
		m.textInput.SetValue("")
		m.textInput.Focus()
		m.overlay = OverlayCreateBucket
	case "d":
		if len(m.storage.buckets) > 0 {
			m.storage.selectedBucket = m.storage.buckets[m.storage.cursor]
			m.storage.deleteInput = ""
			m.textInput.SetValue("")
			m.textInput.Focus()
			m.overlay = OverlayDeleteConfirm
		}
	case "r":
		m.storage.loading = true
		m.err = nil
		return m, fetchBucketsCmd(m.storageService, m.factory)
	case keyTab:
		return m.switchTab(m.activeTab.Next())
	case keyShiftTab:
		return m.switchTab(m.activeTab.Prev())
	}
	return m, nil
}

// --- Bucket detail ---

func (m *Model) handleBucketDetailKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case keyEsc:
		m.viewState = ViewStorageList
		m.err = nil
	case "o":
		m.viewState = ViewStorageObjectList
		m.storage.loading = true
		m.storage.cursor = 0
		m.storage.scrollOffset = 0
		m.err = nil
		return m, fetchObjectsCmd(
			m.storageService,
			m.storage.selectedBucket.Name,
			strings.ToLower(string(m.storage.selectedBucket.Provider)),
			"",
		)
	case "d":
		m.storage.deleteInput = ""
		m.textInput.SetValue("")
		m.textInput.Focus()
		m.overlay = OverlayDeleteConfirm
	case "j", keyDown:
		m.storage.scrollOffset++
	case "k", keyUp:
		if m.storage.scrollOffset > 0 {
			m.storage.scrollOffset--
		}
	}
	return m, nil
}

// --- Object list ---

func (m *Model) handleObjectListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	totalItems := len(m.storage.objects.Objects) + len(m.storage.objects.CommonPrefixes)

	switch key {
	case "j", keyDown:
		if totalItems > 0 {
			m.storage.cursor = min(m.storage.cursor+1, totalItems-1)
		}
	case "k", keyUp:
		if m.storage.cursor > 0 {
			m.storage.cursor--
		}
	case keyEnter:
		if totalItems > 0 {
			prefixCount := len(m.storage.objects.CommonPrefixes)
			if m.storage.cursor < prefixCount {
				// Navigate into a directory prefix.
				prefix := m.storage.objects.CommonPrefixes[m.storage.cursor]
				m.storage.loading = true
				m.storage.cursor = 0
				m.storage.scrollOffset = 0
				m.err = nil
				return m, fetchObjectsCmd(
					m.storageService,
					m.storage.selectedBucket.Name,
					strings.ToLower(string(m.storage.selectedBucket.Provider)),
					prefix,
				)
			}
			// Select an object.
			objIdx := m.storage.cursor - prefixCount
			if objIdx < len(m.storage.objects.Objects) {
				m.storage.selectedObject = m.storage.objects.Objects[objIdx]
				m.viewState = ViewStorageObjectDetail
				m.storage.loading = true
				m.err = nil
				return m, fetchObjectDetailCmd(
					m.storageService,
					m.storage.selectedBucket.Name,
					m.storage.selectedObject.Key,
					strings.ToLower(string(m.storage.selectedBucket.Provider)),
				)
			}
		}
	case keyEsc:
		m.viewState = ViewStorageBucketDetail
		m.storage.cursor = 0
		m.storage.scrollOffset = 0
		m.err = nil
	}
	return m, nil
}

// --- Object detail ---

func (m *Model) handleObjectDetailKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case keyEsc:
		m.viewState = ViewStorageObjectList
		m.err = nil
	case "j", keyDown:
		m.storage.scrollOffset++
	case "k", keyUp:
		if m.storage.scrollOffset > 0 {
			m.storage.scrollOffset--
		}
	}
	return m, nil
}

// --- SQL list ---

func (m *Model) handleSqlListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "j", keyDown:
		if len(m.sql.instances) > 0 {
			m.sql.cursor = min(m.sql.cursor+1, len(m.sql.instances)-1)
		}
	case "k", keyUp:
		if m.sql.cursor > 0 {
			m.sql.cursor--
		}
	case keyEnter:
		if len(m.sql.instances) > 0 {
			m.sql.selectedInstance = m.sql.instances[m.sql.cursor]
			m.viewState = ViewSqlInstanceDetail
			m.sql.loading = true
			m.err = nil
			return m, fetchInstanceDetailCmd(
				m.sqlService,
				m.sql.selectedInstance.Name,
				strings.ToLower(string(m.sql.selectedInstance.Provider)),
			)
		}
	case "r":
		m.sql.loading = true
		m.err = nil
		return m, fetchInstancesCmd(m.sqlService, m.factory)
	case keyTab:
		return m.switchTab(m.activeTab.Next())
	case keyShiftTab:
		return m.switchTab(m.activeTab.Prev())
	}
	return m, nil
}

// --- Instance detail ---

func (m *Model) handleInstanceDetailKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case keyEsc:
		m.viewState = ViewSqlList
		m.err = nil
	case "j", keyDown:
		m.sql.scrollOffset++
	case "k", keyUp:
		if m.sql.scrollOffset > 0 {
			m.sql.scrollOffset--
		}
	}
	return m, nil
}

// --- Config list ---

func (m *Model) handleConfigListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "j", keyDown:
		if len(m.config.entries) > 0 {
			m.config.cursor = min(m.config.cursor+1, len(m.config.entries)-1)
		}
	case "k", keyUp:
		if m.config.cursor > 0 {
			m.config.cursor--
		}
	case keyEnter:
		if len(m.config.entries) > 0 {
			entry := m.config.entries[m.config.cursor]
			m.config.editKey = entry.Key
			m.config.editValue = entry.Value
			m.config.isNewEntry = false
			m.viewState = ViewConfigEdit
			m.textInput.SetValue(entry.Value)
			m.textInput.Focus()
			m.err = nil
		}
	case "a":
		m.config.editKey = ""
		m.config.editValue = ""
		m.config.isNewEntry = true
		m.storage.createField = 0 // reuse for config add field index
		m.textInput.SetValue("")
		m.textInput.Focus()
		m.overlay = OverlayConfigAdd
	case "d":
		if len(m.config.entries) > 0 {
			m.config.editKey = m.config.entries[m.config.cursor].Key
			m.overlay = OverlayConfigDelete
		}
	case "r":
		m.config.loading = true
		m.err = nil
		return m, fetchConfigCmd(m.configManager)
	case keyTab:
		return m.switchTab(m.activeTab.Next())
	case keyShiftTab:
		return m.switchTab(m.activeTab.Prev())
	}
	return m, nil
}

// --- Config edit ---

func (m *Model) handleConfigEditKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case keyEsc:
		m.viewState = ViewConfigList
		m.textInput.Reset()
		m.err = nil
	case keyEnter:
		value := strings.TrimSpace(m.textInput.Value())
		if value == "" {
			return m, nil
		}
		m.config.loading = true
		m.viewState = ViewConfigList
		m.textInput.Reset()
		m.err = nil
		return m, setConfigCmd(m.configManager, m.config.editKey, value)
	default:
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}
	return m, nil
}

// handleConfigDeleteConfirm is invoked when Enter is pressed on the ConfigDelete overlay.
func (m *Model) handleConfigDeleteConfirm() (tea.Model, tea.Cmd) {
	m.overlay = OverlayNone
	m.config.loading = true

	// Extract the provider name (first segment of dot-notation key).
	// Use RemoveProvider to remove the entire provider block cleanly,
	// avoiding validation failures from deleting individual required fields.
	provider := m.config.editKey
	if idx := strings.IndexByte(provider, '.'); idx >= 0 {
		provider = provider[:idx]
	}
	return m, removeProviderCmd(m.configManager, provider)
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
			return m, fetchBucketsCmd(m.storageService, m.factory)
		}
	case TabSql:
		m.viewState = ViewSqlList
		if !m.sql.loaded {
			m.sql.loading = true
			return m, fetchInstancesCmd(m.sqlService, m.factory)
		}
	case TabConfig:
		m.viewState = ViewConfigList
		if !m.config.loaded {
			m.config.loading = true
			return m, fetchConfigCmd(m.configManager)
		}
	}

	return m, nil
}

// --- Data message handlers ---

func (m *Model) handleBucketsLoaded(msg BucketsLoadedMsg) (tea.Model, tea.Cmd) {
	m.storage.loading = false
	m.storage.loaded = true
	if msg.Err != nil {
		m.err = msg.Err
		return m, nil
	}
	m.storage.buckets = msg.Buckets
	m.storage.cursor = 0
	m.storage.scrollOffset = 0
	return m, nil
}

func (m *Model) handleBucketDetailLoaded(msg BucketDetailMsg) (tea.Model, tea.Cmd) {
	m.storage.loading = false
	if msg.Err != nil {
		m.err = msg.Err
		return m, nil
	}
	m.storage.selectedBucket = msg.Bucket
	return m, nil
}

func (m *Model) handleObjectsLoaded(msg ObjectsLoadedMsg) (tea.Model, tea.Cmd) {
	m.storage.loading = false
	if msg.Err != nil {
		m.err = msg.Err
		return m, nil
	}
	m.storage.objects = msg.Objects
	m.storage.cursor = 0
	m.storage.scrollOffset = 0
	return m, nil
}

func (m *Model) handleObjectDetailLoaded(msg ObjectDetailMsg) (tea.Model, tea.Cmd) {
	m.storage.loading = false
	if msg.Err != nil {
		m.err = msg.Err
		return m, nil
	}
	m.storage.selectedObject = msg.Object
	return m, nil
}

func (m *Model) handleInstancesLoaded(msg InstancesLoadedMsg) (tea.Model, tea.Cmd) {
	m.sql.loading = false
	m.sql.loaded = true
	if msg.Err != nil {
		m.err = msg.Err
		return m, nil
	}
	m.sql.instances = msg.Instances
	m.sql.cursor = 0
	m.sql.scrollOffset = 0
	return m, nil
}

func (m *Model) handleInstanceDetailLoaded(msg InstanceDetailMsg) (tea.Model, tea.Cmd) {
	m.sql.loading = false
	if msg.Err != nil {
		m.err = msg.Err
		return m, nil
	}
	m.sql.selectedInstance = msg.Instance
	return m, nil
}

func (m *Model) handleConfigLoaded(msg ConfigLoadedMsg) (tea.Model, tea.Cmd) {
	m.config.loading = false
	m.config.loaded = true
	if msg.Err != nil {
		m.err = msg.Err
		return m, nil
	}
	m.config.entries = msg.Entries
	m.config.cursor = 0
	m.config.scrollOffset = 0
	return m, nil
}

// --- Mutation message handlers ---

func (m *Model) handleBucketCreated(msg BucketCreatedMsg) (tea.Model, tea.Cmd) {
	m.storage.loading = false
	if msg.Err != nil {
		m.err = msg.Err
		return m, nil
	}
	m.statusMessage = "Bucket created successfully"
	m.storage.loading = true
	return m, tea.Batch(fetchBucketsCmd(m.storageService, m.factory), clearStatusCmd())
}

func (m *Model) handleBucketDeleted(msg BucketDeletedMsg) (tea.Model, tea.Cmd) {
	m.storage.loading = false
	if msg.Err != nil {
		m.err = msg.Err
		return m, nil
	}
	m.statusMessage = "Bucket deleted successfully"
	m.viewState = ViewStorageList
	m.storage.loading = true
	return m, tea.Batch(fetchBucketsCmd(m.storageService, m.factory), clearStatusCmd())
}

func (m *Model) handleConfigUpdated(msg ConfigUpdatedMsg) (tea.Model, tea.Cmd) {
	m.config.loading = false
	if msg.Err != nil {
		m.err = msg.Err
		return m, nil
	}
	m.refreshFactoryConfig()
	m.statusMessage = "Configuration updated"
	m.config.loading = true
	return m, tea.Batch(fetchConfigCmd(m.configManager), clearStatusCmd())
}

func (m *Model) handleConfigDeleted(msg ConfigDeletedMsg) (tea.Model, tea.Cmd) {
	m.config.loading = false
	if msg.Err != nil {
		m.err = msg.Err
		return m, nil
	}
	m.refreshFactoryConfig()
	m.statusMessage = "Configuration entry deleted"
	m.config.loading = true
	return m, tea.Batch(fetchConfigCmd(m.configManager), clearStatusCmd())
}

func (m *Model) handleProviderRemoved(msg ProviderRemovedMsg) (tea.Model, tea.Cmd) {
	m.config.loading = false
	if msg.Err != nil {
		m.err = msg.Err
		return m, nil
	}
	m.refreshFactoryConfig()
	m.statusMessage = "Provider removed"
	m.config.cursor = 0
	m.config.loading = true
	// Invalidate cached data for tabs that depend on provider config
	m.storage.loaded = false
	m.sql.loaded = false
	return m, tea.Batch(fetchConfigCmd(m.configManager), clearStatusCmd())
}

// refreshFactoryConfig reloads the config from disk and updates the factory
// so provider status checks and queries reflect the latest configuration.
func (m *Model) refreshFactoryConfig() {
	cfg, err := m.configManager.LoadConfig()
	if err != nil {
		return
	}
	m.cfg = cfg
	m.factory.UpdateConfig(cfg)
}
