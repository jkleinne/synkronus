// File: internal/tui/tui.go
package tui

import (
	"log/slog"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"synkronus/internal/config"
	"synkronus/internal/provider/factory"
	"synkronus/internal/service"
	"synkronus/internal/tui/ui"
)

// Deps holds the external dependencies injected into the TUI from the CLI layer.
type Deps struct {
	StorageService *service.StorageService
	SqlService     *service.SqlService
	ConfigManager  *config.ConfigManager
	Config         *config.Config
	Factory        *factory.Factory
	Logger         *slog.Logger
}

// Run creates the Bubble Tea model and starts the interactive TUI program.
func Run(deps Deps) error {
	m := NewModel(deps)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// Model is the single Bubble Tea model for the entire TUI.
// It acts as a thin router, delegating state and logic to focused sub-models.
type Model struct {
	deps Deps

	viewState ViewState
	overlay   OverlayState
	activeTab Tab

	width  int
	height int

	storage StorageModel
	sql     SqlModel
	config  ConfigModel

	spinner       spinner.Model
	textInput     textinput.Model
	err           error
	statusMessage string
}

// deleteTarget distinguishes which resource type a delete confirmation modal targets.
type deleteTarget int

const (
	deleteTargetBucket deleteTarget = iota
	deleteTargetObject
)

// NewModel constructs the TUI model with injected dependencies and sensible defaults.
func NewModel(deps Deps) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot

	ti := textinput.New()
	ti.CharLimit = textInputCharLimit

	return Model{
		deps: deps,

		viewState: ViewStorageList,
		overlay:   OverlayNone,
		activeTab: TabStorage,

		spinner:   s,
		textInput: ti,
	}
}

// Init implements tea.Model. It starts the spinner and triggers the initial bucket fetch.
func (m Model) Init() tea.Cmd {
	m.storage.loading = true
	return tea.Batch(m.spinner.Tick, fetchBucketsCmd(m.deps.StorageService, m.deps.Factory))
}

// Update implements tea.Model. It dispatches incoming messages to the appropriate handler.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return &m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return &m, cmd

	case StatusClearMsg:
		m.statusMessage = ""
		return &m, nil

	// --- Data messages ---
	case BucketsLoadedMsg:
		return m.handleBucketsLoaded(msg)
	case BucketDetailMsg:
		return m.handleBucketDetailLoaded(msg)
	case ObjectsLoadedMsg:
		return m.handleObjectsLoaded(msg)
	case ObjectDetailMsg:
		return m.handleObjectDetailLoaded(msg)
	case InstancesLoadedMsg:
		return m.handleInstancesLoaded(msg)
	case InstanceDetailMsg:
		return m.handleInstanceDetailLoaded(msg)
	case ConfigLoadedMsg:
		return m.handleConfigLoaded(msg)

	// --- Mutation messages ---
	case BucketCreatedMsg:
		return m.handleBucketCreated(msg)
	case BucketDeletedMsg:
		return m.handleBucketDeleted(msg)
	case ObjectDownloadedMsg:
		return m.handleObjectDownloaded(msg)
	case ObjectUploadedMsg:
		return m.handleObjectUploaded(msg)
	case ObjectDeletedMsg:
		return m.handleObjectDeleted(msg)
	case ConfigUpdatedMsg:
		return m.handleConfigUpdated(msg)
	case ConfigDeletedMsg:
		return m.handleConfigDeleted(msg)
	case ProviderRemovedMsg:
		return m.handleProviderRemoved(msg)

	case tea.KeyMsg:
		// Overlay intercepts keys first to prevent leaking to the base view.
		if m.overlay != OverlayNone {
			return m.handleOverlayKeys(msg)
		}
		return m.handleViewKeys(msg)
	}

	return &m, nil
}

// View implements tea.Model. It renders the full TUI frame.
func (m Model) View() string {
	if m.overlay != OverlayNone {
		return m.renderOverlay()
	}

	var b strings.Builder

	// Banner
	b.WriteString("\n")
	b.WriteString(ui.RenderBanner(m.width))
	b.WriteString("\n")

	// Tabs or breadcrumb
	if m.isListView() {
		b.WriteString(ui.RenderTabs(int(m.activeTab), m.width, m.providerStatuses()))
	} else {
		breadcrumb := ui.RenderBreadcrumb(m.breadcrumbParts())
		b.WriteString(ui.CenterContent(breadcrumb, m.width))
	}
	b.WriteString("\n")

	// Main content
	b.WriteString(m.renderContent())

	// Error or status line
	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(ui.RenderError(m.err.Error(), m.width))
	} else if m.statusMessage != "" {
		b.WriteString("\n")
		b.WriteString(ui.RenderStatusMessage(m.statusMessage, m.width))
	}

	// Key hints
	b.WriteString("\n")
	b.WriteString(ui.RenderKeyHints(m.currentBindingContext(), m.width))
	b.WriteString("\n")

	return b.String()
}

// isListView returns true when the current viewState is a top-level list.
func (m *Model) isListView() bool {
	switch m.viewState {
	case ViewStorageList, ViewSqlList, ViewConfigList:
		return true
	default:
		return false
	}
}

// providerStatuses builds the provider status dots for the tab bar.
func (m *Model) providerStatuses() []ui.ProviderStatus {
	storageProviders := m.deps.Factory.SupportedStorageProviders()
	statuses := make([]ui.ProviderStatus, 0, len(storageProviders))
	for _, name := range storageProviders {
		statuses = append(statuses, ui.ProviderStatus{
			Name:       name,
			Configured: m.deps.Factory.IsConfigured(name),
		})
	}
	return statuses
}

// breadcrumbParts returns the navigation trail for the current detail view.
func (m *Model) breadcrumbParts() []string {
	switch m.viewState {
	case ViewStorageBucketDetail:
		return []string{"Storage", m.storage.selectedBucket.Name}
	case ViewStorageObjectList:
		return []string{"Storage", m.storage.selectedBucket.Name, "Objects"}
	case ViewStorageObjectDetail:
		return []string{"Storage", m.storage.selectedBucket.Name, "Objects", m.storage.selectedObject.Key}
	case ViewSqlInstanceDetail:
		return []string{"SQL", m.sql.selectedInstance.Name}
	case ViewConfigEdit:
		return []string{"Config", m.config.editKey}
	default:
		return nil
	}
}

// currentBindingContext maps the current viewState to the keybinding context used for hints.
func (m *Model) currentBindingContext() ui.BindingContext {
	switch m.viewState {
	case ViewStorageList:
		return ui.ContextStorageList
	case ViewStorageBucketDetail:
		return ui.ContextBucketDetail
	case ViewStorageObjectList:
		return ui.ContextObjectList
	case ViewStorageObjectDetail:
		return ui.ContextObjectDetail
	case ViewSqlList:
		return ui.ContextSqlList
	case ViewSqlInstanceDetail:
		return ui.ContextInstanceDetail
	case ViewConfigList:
		return ui.ContextConfigList
	case ViewConfigEdit:
		return ui.ContextConfigEdit
	default:
		return ui.ContextStorageList
	}
}

// renderContent dispatches to the appropriate sub-model's renderer based on viewState.
func (m *Model) renderContent() string {
	switch m.viewState {
	case ViewStorageList, ViewStorageBucketDetail, ViewStorageObjectList, ViewStorageObjectDetail:
		return m.storage.RenderContent(m.viewState, m.spinner.View(), m.width)
	case ViewSqlList, ViewSqlInstanceDetail:
		return m.sql.RenderContent(m.viewState, m.spinner.View(), m.width)
	case ViewConfigList, ViewConfigEdit:
		return m.config.RenderContent(m.viewState, m.textInput.View(), m.spinner.View(), m.width, m.height)
	default:
		return ""
	}
}

// applyViewUpdate applies state-change requests from a sub-model's ViewUpdate to the root Model.
func (m *Model) applyViewUpdate(vu ViewUpdate) {
	if vu.ViewState != nil {
		m.viewState = *vu.ViewState
	}
	if vu.Overlay != nil {
		m.overlay = *vu.Overlay
	}
	if vu.ClearErr {
		m.err = nil
	}
	if vu.ResetTextInput {
		m.textInput.Reset()
	}
	if vu.TextInputValue != nil {
		m.textInput.SetValue(*vu.TextInputValue)
	}
	if vu.FocusTextInput {
		m.textInput.Focus()
	}
}

// refreshFactoryConfig reloads the config from disk and updates the factory
// so provider status checks and queries reflect the latest configuration.
func (m *Model) refreshFactoryConfig() {
	cfg, err := m.deps.ConfigManager.LoadConfig()
	if err != nil {
		return
	}
	m.deps.Config = cfg
	m.deps.Factory.UpdateConfig(cfg)
}
