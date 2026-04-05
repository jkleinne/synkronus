package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"synkronus/internal/domain/storage"
	"synkronus/internal/tui/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// HandleOverlayKeys intercepts keystrokes when a modal overlay is active.
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
			m.syncTextInputToField()
			switch m.overlay {
			case OverlayCreateBucket:
				m.storage.createFieldIndex = m.storage.nextVisibleCreateField(m.storage.createFieldIndex)
				m.loadFieldIntoTextInput()
			case OverlayConfigAdd:
				if m.storage.createFieldIndex == 0 {
					m.config.editKey = m.textInput.Value()
					m.storage.createFieldIndex = 1
					m.textInput.SetValue(m.config.editValue)
				} else {
					m.config.editValue = m.textInput.Value()
					m.storage.createFieldIndex = 0
					m.textInput.SetValue(m.config.editKey)
				}
				m.textInput.Focus()
			case OverlayUploadObject:
				if m.storage.uploadField == 0 {
					m.storage.uploadFilePath = m.textInput.Value()
					m.storage.uploadField = 1
					m.textInput.SetValue(m.storage.uploadObjectKey)
				} else {
					m.storage.uploadObjectKey = m.textInput.Value()
					m.storage.uploadField = 0
					m.textInput.SetValue(m.storage.uploadFilePath)
				}
				m.textInput.Focus()
			}
			return m, nil
		default:
			if m.overlay == OverlayCreateBucket {
				if options := m.storage.getCreateFieldOptions(m.storage.createFieldIndex); len(options) > 0 {
					value := cycleOption(options, m.storage.getCreateFieldValue(m.storage.createFieldIndex), key)
					m.storage.setCreateFieldValue(m.storage.createFieldIndex, value)
					m.textInput.SetValue(value)
					if m.storage.createFieldIndex == createFieldProvider {
						m.storage.updateCreateHiddenFields()
					}
					return m, nil
				}
			}
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
		return m.submitCreateBucket()
	case OverlayDeleteConfirm:
		return m.submitDeleteConfirm()
	case OverlayConfigAdd:
		return m.submitConfigAdd()
	case OverlayDownloadPath:
		return m.submitDownloadPath()
	case OverlayUploadObject:
		return m.submitUploadObject()
	}

	return m, nil
}

func (m *Model) submitCreateBucket() (tea.Model, tea.Cmd) {
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
	if v := strings.TrimSpace(strings.ToLower(m.storage.createVersioning)); v == optionYes {
		t := true
		opts.Versioning = &t
	} else if v == optionNo {
		f := false
		opts.Versioning = &f
	}
	if v := strings.TrimSpace(strings.ToLower(m.storage.createUniformAccess)); v == optionYes {
		t := true
		opts.UniformBucketLevelAccess = &t
	} else if v == optionNo {
		f := false
		opts.UniformBucketLevelAccess = &f
	}
	if v := strings.TrimSpace(strings.ToLower(m.storage.createPublicAccessPrevention)); v == storage.PublicAccessPreventionEnforced || v == storage.PublicAccessPreventionInherited {
		opts.PublicAccessPrevention = &v
	}
	m.overlay = OverlayNone
	m.storage.loading = true
	m.textInput.Reset()
	return m, createBucketCmd(m.deps.StorageService, opts, provider)
}

func (m *Model) submitDeleteConfirm() (tea.Model, tea.Cmd) {
	switch m.storage.deleteKind {
	case deleteTargetBucket:
		if m.textInput.Value() != m.storage.selectedBucket.Name {
			return m, nil
		}
		m.overlay = OverlayNone
		m.storage.loading = true
		m.textInput.Reset()
		return m, deleteBucketCmd(
			m.deps.StorageService,
			m.storage.selectedBucket.Name,
			strings.ToLower(string(m.storage.selectedBucket.Provider)),
		)
	case deleteTargetObject:
		if m.textInput.Value() != m.storage.deleteObjectKey {
			return m, nil
		}
		m.overlay = OverlayNone
		m.storage.loading = true
		m.textInput.Reset()
		return m, deleteObjectCmd(
			m.deps.StorageService,
			m.storage.selectedBucket.Name,
			m.storage.deleteObjectKey,
			strings.ToLower(string(m.storage.selectedBucket.Provider)),
		)
	}
	return m, nil
}

func (m *Model) submitConfigAdd() (tea.Model, tea.Cmd) {
	key := strings.TrimSpace(m.config.editKey)
	value := strings.TrimSpace(m.config.editValue)
	if key == "" || value == "" {
		return m, nil
	}
	m.overlay = OverlayNone
	m.config.loading = true
	m.textInput.Reset()
	return m, setConfigCmd(m.deps.ConfigManager, key, value)
}

func (m *Model) submitDownloadPath() (tea.Model, tea.Cmd) {
	dir := strings.TrimSpace(m.storage.downloadDir)
	if dir == "" {
		return m, nil
	}
	m.overlay = OverlayNone
	m.storage.loading = true
	m.err = nil
	m.textInput.Reset()
	return m, downloadObjectCmd(
		m.deps.StorageService,
		m.storage.selectedBucket.Name,
		m.storage.downloadingKey,
		strings.ToLower(string(m.storage.selectedBucket.Provider)),
		dir,
	)
}

func (m *Model) submitUploadObject() (tea.Model, tea.Cmd) {
	filePath := strings.TrimSpace(m.storage.uploadFilePath)
	if filePath == "" {
		return m, nil
	}
	objectKey := strings.TrimSpace(m.storage.uploadObjectKey)
	m.overlay = OverlayNone
	m.storage.loading = true
	m.err = nil
	m.textInput.Reset()
	if objectKey != "" {
		m.storage.uploadObjectKey = objectKey
	} else {
		m.storage.uploadObjectKey = filepath.Base(filePath)
	}
	return m, uploadObjectCmd(
		m.deps.StorageService,
		m.storage.selectedBucket.Name,
		strings.ToLower(string(m.storage.selectedBucket.Provider)),
		filePath,
		objectKey,
	)
}

// handleConfigDeleteConfirm is invoked when Enter is pressed on the ConfigDelete overlay.
func (m *Model) handleConfigDeleteConfirm() (tea.Model, tea.Cmd) {
	m.overlay = OverlayNone
	m.config.loading = true

	provider := m.config.editKey
	if idx := strings.IndexByte(provider, '.'); idx >= 0 {
		provider = provider[:idx]
	}
	return m, removeProviderCmd(m.deps.ConfigManager, provider)
}

// syncTextInputToField persists the current textinput value back to the active field.
func (m *Model) syncTextInputToField() {
	switch m.overlay {
	case OverlayCreateBucket:
		if m.storage.getCreateFieldOptions(m.storage.createFieldIndex) == nil {
			m.storage.setCreateFieldValue(m.storage.createFieldIndex, m.textInput.Value())
		}
	case OverlayDeleteConfirm:
		m.storage.deleteInput = m.textInput.Value()
	case OverlayConfigAdd:
		if m.storage.createFieldIndex == 0 {
			m.config.editKey = m.textInput.Value()
		} else {
			m.config.editValue = m.textInput.Value()
		}
	case OverlayDownloadPath:
		m.storage.downloadDir = m.textInput.Value()
	case OverlayUploadObject:
		if m.storage.uploadField == 0 {
			m.storage.uploadFilePath = m.textInput.Value()
		} else {
			m.storage.uploadObjectKey = m.textInput.Value()
		}
	}
}

// loadFieldIntoTextInput sets the textinput value to the current create-bucket field.
func (m *Model) loadFieldIntoTextInput() {
	m.textInput.SetValue(m.storage.getCreateFieldValue(m.storage.createFieldIndex))
	m.textInput.Focus()
}

// --- Overlay rendering ---

// renderOverlay renders the active modal overlay, replacing the base view entirely.
func (m *Model) renderOverlay() string {
	switch m.overlay {
	case OverlayHelp:
		content := ui.RenderHelpContent()
		return ui.RenderModal("Help", content, m.width, m.height)

	case OverlayCreateBucket:
		selectorFields := make(map[int]bool)
		for i := range createFormFieldCount {
			if m.storage.getCreateFieldOptions(i) != nil {
				selectorFields[i] = true
			}
		}
		content := ui.RenderCreateBucketForm(
			ui.CreateBucketFormFields{
				Name:                   m.storage.createName,
				Provider:               m.storage.createProvider,
				Location:               m.storage.createLocation,
				StorageClass:           m.storage.createStorageClass,
				Labels:                 m.storage.createLabels,
				Versioning:             m.storage.createVersioning,
				UniformAccess:          m.storage.createUniformAccess,
				PublicAccessPrevention: m.storage.createPublicAccessPrevention,
				SelectorFields:         selectorFields,
				HiddenFields:           m.storage.createHiddenFields,
			},
			m.storage.createFieldIndex,
			m.textInput.View(),
		)
		return ui.RenderModal("Create Bucket", content, m.width, m.height)

	case OverlayDeleteConfirm:
		var title, targetName string
		switch m.storage.deleteKind {
		case deleteTargetBucket:
			title = "Delete Bucket"
			targetName = m.storage.selectedBucket.Name
		case deleteTargetObject:
			title = "Delete Object"
			targetName = m.storage.deleteObjectKey
		}
		content := ui.RenderDeleteConfirm(targetName, m.textInput.View())
		return ui.RenderModal(title, content, m.width, m.height)

	case OverlayConfigAdd:
		content := fmt.Sprintf(
			"%s %s\n%s %s",
			ui.TextDimStyle.Render("Key:"),
			m.textInput.View(),
			ui.TextDimStyle.Render("Value:"),
			ui.TextSecondaryStyle.Render(m.config.editValue),
		)
		if m.config.isNewEntry && m.storage.createFieldIndex == 1 {
			content = fmt.Sprintf(
				"%s %s\n%s %s",
				ui.TextDimStyle.Render("Key:"),
				ui.TextSecondaryStyle.Render(m.config.editKey),
				ui.TextDimStyle.Render("Value:"),
				m.textInput.View(),
			)
		}
		return ui.RenderModal("Add Config Entry", content, m.width, m.height)

	case OverlayConfigDelete:
		content := ui.RenderConfigDeleteConfirm(m.config.editKey)
		return ui.RenderModal("Remove Provider", content, m.width, m.height)

	case OverlayDownloadPath:
		content := fmt.Sprintf(
			"%s %s\n\n%s %s",
			ui.TextDimStyle.Render("Object:"),
			ui.TextSecondaryStyle.Render(m.storage.downloadingKey),
			ui.TextDimStyle.Render("Save to:"),
			m.textInput.View(),
		)
		return ui.RenderModal("Download Object", content, m.width, m.height)

	case OverlayUploadObject:
		fieldLabels := []string{"File path:", "Object key (optional):"}
		values := []string{m.storage.uploadFilePath, m.storage.uploadObjectKey}
		var lines []string
		for i, label := range fieldLabels {
			if i == m.storage.uploadField {
				lines = append(lines, fmt.Sprintf("%s %s", ui.TextDimStyle.Render(label), m.textInput.View()))
			} else {
				lines = append(lines, fmt.Sprintf("%s %s", ui.TextDimStyle.Render(label), ui.TextSecondaryStyle.Render(values[i])))
			}
		}
		content := strings.Join(lines, "\n")
		return ui.RenderModal("Upload Object", content, m.width, m.height)

	default:
		return ""
	}
}
