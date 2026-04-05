package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	domainsql "synkronus/internal/domain/sql"
	"synkronus/internal/service"
	"synkronus/internal/tui/ui"
)

// SqlModel owns the mutable state and key/message handling for the SQL tab.
type SqlModel struct {
	instances        []domainsql.Instance
	selectedInstance domainsql.Instance
	cursor           int
	scrollOffset     int
	loading          bool
	loaded           bool
}

// --- Key handlers ---

// HandleListKeys handles keystrokes on the SQL instance list view.
func (s *SqlModel) HandleListKeys(msg tea.KeyMsg, svc *service.SqlService, deps *Deps) (ViewUpdate, tea.Cmd) {
	key := msg.String()
	switch key {
	case "j", keyDown:
		if len(s.instances) > 0 {
			s.cursor = min(s.cursor+1, len(s.instances)-1)
		}
	case "k", keyUp:
		if s.cursor > 0 {
			s.cursor--
		}
	case keyEnter:
		if len(s.instances) > 0 {
			s.selectedInstance = s.instances[s.cursor]
			s.loading = true
			return ViewUpdate{ViewState: ptrViewState(ViewSqlInstanceDetail), ClearErr: true}, fetchInstanceDetailCmd(
				svc,
				s.selectedInstance.Name,
				strings.ToLower(string(s.selectedInstance.Provider)),
			)
		}
	case "r":
		s.loading = true
		return ViewUpdate{ClearErr: true}, fetchInstancesCmd(svc, deps.Factory)
	case keyTab:
		return ViewUpdate{SwitchTab: ptrTab(TabSql.Next())}, nil
	case keyShiftTab:
		return ViewUpdate{SwitchTab: ptrTab(TabSql.Prev())}, nil
	}
	return ViewUpdate{}, nil
}

// HandleInstanceDetailKeys handles keystrokes on the SQL instance detail view.
func (s *SqlModel) HandleInstanceDetailKeys(msg tea.KeyMsg) (ViewUpdate, tea.Cmd) {
	key := msg.String()
	switch key {
	case keyEsc:
		return ViewUpdate{ViewState: ptrViewState(ViewSqlList), ClearErr: true}, nil
	case "j", keyDown:
		s.scrollOffset++
	case "k", keyUp:
		if s.scrollOffset > 0 {
			s.scrollOffset--
		}
	}
	return ViewUpdate{}, nil
}

// --- Message handlers ---

// HandleInstancesLoaded processes a completed SQL instance list fetch.
func (s *SqlModel) HandleInstancesLoaded(msg InstancesLoadedMsg) error {
	s.loading = false
	s.loaded = true
	if msg.Err != nil {
		return msg.Err
	}
	s.instances = msg.Instances
	s.cursor = 0
	s.scrollOffset = 0
	return nil
}

// HandleInstanceDetailLoaded processes a completed SQL instance detail fetch.
func (s *SqlModel) HandleInstanceDetailLoaded(msg InstanceDetailMsg) error {
	s.loading = false
	if msg.Err != nil {
		return msg.Err
	}
	s.selectedInstance = msg.Instance
	return nil
}

// --- View rendering ---

// RenderContent returns the rendered view for SQL-related viewStates.
func (s *SqlModel) RenderContent(viewState ViewState, spinnerView string, width int) string {
	switch viewState {
	case ViewSqlList:
		if s.loading {
			return ui.CenterContent(ui.RenderSpinnerView(spinnerView, "Loading instances..."), width)
		}
		return ui.RenderInstanceList(s.instances, s.cursor, s.scrollOffset, width)

	case ViewSqlInstanceDetail:
		if s.loading {
			return ui.CenterContent(ui.RenderSpinnerView(spinnerView, "Loading instance details..."), width)
		}
		return ui.RenderInstanceDetail(s.selectedInstance, width)

	default:
		return ""
	}
}
