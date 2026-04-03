// File: internal/tui/tui.go
package tui

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"synkronus/internal/config"
	domainsql "synkronus/internal/domain/sql"
	"synkronus/internal/domain/storage"
	"synkronus/internal/provider/factory"
	"synkronus/internal/provider/registry"
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
type Model struct {
	storageService *service.StorageService
	sqlService     *service.SqlService
	configManager  *config.ConfigManager
	cfg            *config.Config
	factory        *factory.Factory
	logger         *slog.Logger

	viewState ViewState
	overlay   OverlayState
	activeTab Tab

	width  int
	height int

	storage storageState
	sql     sqlState
	config  configState

	spinner       spinner.Model
	textInput     textinput.Model
	err           error
	statusMessage string
}

// storageState holds the mutable state for the Storage tab.
type storageState struct {
	buckets        []storage.Bucket
	objects        storage.ObjectList
	selectedBucket storage.Bucket
	selectedObject storage.Object
	cursor         int
	scrollOffset   int
	loading        bool
	loaded         bool
	createName                   string
	createProvider               string
	createLocation               string
	availableProviders           []string
	createStorageClass           string
	createLabels                 string
	createVersioning             string // "yes"/"no"/""
	createUniformAccess          string // "yes"/"no"/""
	createPublicAccessPrevention string // "enforced"/"inherited"/""
	createField                  int
	deleteInput                  string
}

// sqlState holds the mutable state for the SQL tab.
type sqlState struct {
	instances        []domainsql.Instance
	selectedInstance domainsql.Instance
	cursor           int
	scrollOffset     int
	loading          bool
	loaded           bool
}

// configState holds the mutable state for the Config tab.
type configState struct {
	entries      []ui.ConfigEntry
	cursor       int
	scrollOffset int
	loading      bool
	loaded       bool
	editKey      string
	editValue    string
	isNewEntry   bool
}

// NewModel constructs the TUI model with injected dependencies and sensible defaults.
func NewModel(deps Deps) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot

	ti := textinput.New()
	ti.CharLimit = 256

	return Model{
		storageService: deps.StorageService,
		sqlService:     deps.SqlService,
		configManager:  deps.ConfigManager,
		cfg:            deps.Config,
		factory:        deps.Factory,
		logger:         deps.Logger,

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
	return tea.Batch(m.spinner.Tick, fetchBucketsCmd(m.storageService, m.factory))
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
	storageProviders := registry.GetSupportedProviders()
	statuses := make([]ui.ProviderStatus, 0, len(storageProviders))
	for _, name := range storageProviders {
		statuses = append(statuses, ui.ProviderStatus{
			Name:       name,
			Configured: m.factory.IsConfigured(name),
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

// renderContent dispatches to the appropriate ui.Render* function based on viewState.
func (m *Model) renderContent() string {
	switch m.viewState {
	case ViewStorageList:
		if m.storage.loading {
			return ui.CenterContent(ui.RenderSpinnerView(m.spinner.View(), "Loading buckets..."), m.width)
		}
		return ui.RenderBucketList(m.storage.buckets, m.storage.cursor, m.storage.scrollOffset, m.width)

	case ViewStorageBucketDetail:
		if m.storage.loading {
			return ui.CenterContent(ui.RenderSpinnerView(m.spinner.View(), "Loading bucket details..."), m.width)
		}
		return ui.RenderBucketDetail(m.storage.selectedBucket, m.width)

	case ViewStorageObjectList:
		if m.storage.loading {
			return ui.CenterContent(ui.RenderSpinnerView(m.spinner.View(), "Loading objects..."), m.width)
		}
		return ui.RenderObjectList(m.storage.objects, m.storage.cursor, m.storage.scrollOffset, m.width)

	case ViewStorageObjectDetail:
		if m.storage.loading {
			return ui.CenterContent(ui.RenderSpinnerView(m.spinner.View(), "Loading object details..."), m.width)
		}
		return ui.RenderObjectDetail(m.storage.selectedObject, m.width)

	case ViewSqlList:
		if m.sql.loading {
			return ui.CenterContent(ui.RenderSpinnerView(m.spinner.View(), "Loading instances..."), m.width)
		}
		return ui.RenderInstanceList(m.sql.instances, m.sql.cursor, m.sql.scrollOffset, m.width)

	case ViewSqlInstanceDetail:
		if m.sql.loading {
			return ui.CenterContent(ui.RenderSpinnerView(m.spinner.View(), "Loading instance details..."), m.width)
		}
		return ui.RenderInstanceDetail(m.sql.selectedInstance, m.width)

	case ViewConfigList:
		if m.config.loading {
			return ui.CenterContent(ui.RenderSpinnerView(m.spinner.View(), "Loading configuration..."), m.width)
		}
		return ui.RenderConfigList(m.config.entries, m.config.cursor, m.config.scrollOffset, m.width)

	case ViewConfigEdit:
		return ui.RenderConfigEdit(m.config.editKey, m.textInput.View(), m.config.isNewEntry, m.width, m.height)

	default:
		return ""
	}
}

// renderOverlay renders the active modal overlay, replacing the base view entirely.
func (m *Model) renderOverlay() string {
	switch m.overlay {
	case OverlayHelp:
		content := ui.RenderHelpContent(int(m.viewState), int(m.activeTab))
		return ui.RenderModal("Help", content, m.width, m.height)

	case OverlayCreateBucket:
		selectorFields := make(map[int]bool)
		for i := 0; i < createFormFieldCount; i++ {
			if m.getCreateFieldOptions(i) != nil {
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
			},
			m.storage.createField,
			m.textInput.View(),
		)
		return ui.RenderModal("Create Bucket", content, m.width, m.height)

	case OverlayDeleteConfirm:
		content := ui.RenderDeleteConfirm(
			m.storage.selectedBucket.Name,
			m.storage.deleteInput,
			m.textInput.View(),
		)
		return ui.RenderModal("Delete Bucket", content, m.width, m.height)

	case OverlayConfigAdd:
		content := fmt.Sprintf(
			"%s %s\n%s %s",
			ui.TextDimStyle.Render("Key:"),
			m.textInput.View(),
			ui.TextDimStyle.Render("Value:"),
			ui.TextSecondaryStyle.Render(m.config.editValue),
		)
		if m.config.isNewEntry && m.storage.createField == 1 {
			// Second field focused: show key as static, value as input
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

	default:
		return ""
	}
}
