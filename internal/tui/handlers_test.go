package tui

import (
	"errors"
	"log/slog"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"synkronus/internal/config"
	"synkronus/internal/domain"
	domainsql "synkronus/internal/domain/sql"
	"synkronus/internal/domain/storage"
	"synkronus/internal/provider/factory"
	"synkronus/internal/tui/ui"
)

// newTestModel returns a minimal Model suitable for handler unit tests.
// Callers set specific fields (buckets, overlay, viewState, etc.) after creation.
func newTestModel() Model {
	return NewModel(Deps{})
}

// newTestModelWithConfig returns a Model with a real ConfigManager and Factory,
// required for handlers that call refreshFactoryConfig (e.g., handleProviderRemoved).
func newTestModelWithConfig(t *testing.T) Model {
	t.Helper()
	cm, err := config.NewConfigManager()
	if err != nil {
		t.Fatalf("creating ConfigManager: %v", err)
	}
	f := factory.NewFactory(&config.Config{}, slog.Default())
	return NewModel(Deps{
		ConfigManager: cm,
		Factory:       f,
	})
}

func runeKey(r rune) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
}

// --- handleViewKeys ---

func TestHandleViewKeys_QuitKeys(t *testing.T) {
	tests := []struct {
		name string
		msg  tea.KeyMsg
	}{
		{"q_key", runeKey('q')},
		{"ctrl+c", tea.KeyMsg{Type: tea.KeyCtrlC}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestModel()
			_, cmd := m.handleViewKeys(tt.msg)
			if cmd == nil {
				t.Fatal("expected non-nil cmd for quit key")
			}
		})
	}
}

func TestHandleViewKeys_HelpOverlay(t *testing.T) {
	m := newTestModel()
	result, cmd := m.handleViewKeys(runeKey('h'))
	if cmd != nil {
		t.Errorf("expected nil cmd, got non-nil")
	}
	model := result.(*Model)
	if model.overlay != OverlayHelp {
		t.Errorf("overlay = %d, want OverlayHelp (%d)", model.overlay, OverlayHelp)
	}
}

// --- handleStorageListKeys ---

func TestHandleStorageListKeys_CursorNavigation(t *testing.T) {
	buckets := []storage.Bucket{
		{Name: "a"}, {Name: "b"}, {Name: "c"},
	}

	tests := []struct {
		name       string
		startCur   int
		msg        tea.KeyMsg
		wantCursor int
	}{
		{"j_from_0", 0, runeKey('j'), 1},
		{"j_from_1", 1, runeKey('j'), 2},
		{"j_at_last", 2, runeKey('j'), 2},
		{"k_from_2", 2, runeKey('k'), 1},
		{"k_from_1", 1, runeKey('k'), 0},
		{"k_at_first", 0, runeKey('k'), 0},
		{"down_from_0", 0, tea.KeyMsg{Type: tea.KeyDown}, 1},
		{"up_from_1", 1, tea.KeyMsg{Type: tea.KeyUp}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestModel()
			m.storage.buckets = buckets
			m.storage.cursor = tt.startCur
			m.viewState = ViewStorageList

			_, _ = m.handleStorageListKeys(tt.msg)
			if m.storage.cursor != tt.wantCursor {
				t.Errorf("cursor = %d, want %d", m.storage.cursor, tt.wantCursor)
			}
		})
	}
}

func TestHandleStorageListKeys_EnterSelectsBucket(t *testing.T) {
	m := newTestModel()
	m.viewState = ViewStorageList
	m.storage.buckets = []storage.Bucket{
		{Name: "bucket-0", Provider: domain.GCP},
		{Name: "bucket-1", Provider: domain.GCP},
	}
	m.storage.cursor = 1

	_, cmd := m.handleStorageListKeys(tea.KeyMsg{Type: tea.KeyEnter})

	if m.storage.selectedBucket.Name != "bucket-1" {
		t.Errorf("selectedBucket.Name = %q, want %q", m.storage.selectedBucket.Name, "bucket-1")
	}
	if m.viewState != ViewStorageBucketDetail {
		t.Errorf("viewState = %d, want ViewStorageBucketDetail (%d)", m.viewState, ViewStorageBucketDetail)
	}
	if cmd == nil {
		t.Error("expected non-nil cmd to fetch bucket detail")
	}
}

func TestHandleStorageListKeys_EmptyList(t *testing.T) {
	m := newTestModel()
	m.viewState = ViewStorageList
	m.storage.buckets = nil

	keys := []tea.KeyMsg{
		runeKey('j'),
		runeKey('k'),
		{Type: tea.KeyEnter},
	}
	for _, msg := range keys {
		_, cmd := m.handleStorageListKeys(msg)
		if cmd != nil {
			t.Errorf("key %q on empty list should produce nil cmd", msg.String())
		}
	}
	if m.storage.cursor != 0 {
		t.Errorf("cursor should remain 0 on empty list, got %d", m.storage.cursor)
	}
}

func TestHandleStorageListKeys_TabSwitchesTab(t *testing.T) {
	m := newTestModel()
	m.viewState = ViewStorageList
	m.activeTab = TabStorage

	_, _ = m.handleStorageListKeys(tea.KeyMsg{Type: tea.KeyTab})

	if m.activeTab != TabSql {
		t.Errorf("activeTab = %d, want TabSql (%d)", m.activeTab, TabSql)
	}
}

// --- handleOverlayKeys ---

func TestHandleOverlayKeys_EscClosesTextOverlay(t *testing.T) {
	textOverlays := []struct {
		name    string
		overlay OverlayState
	}{
		{"CreateBucket", OverlayCreateBucket},
		{"ConfigAdd", OverlayConfigAdd},
		{"DeleteConfirm", OverlayDeleteConfirm},
	}
	for _, tt := range textOverlays {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestModel()
			m.overlay = tt.overlay

			_, cmd := m.handleOverlayKeys(tea.KeyMsg{Type: tea.KeyEsc})

			if m.overlay != OverlayNone {
				t.Errorf("overlay = %d, want OverlayNone", m.overlay)
			}
			if cmd != nil {
				t.Errorf("expected nil cmd after Esc, got non-nil")
			}
		})
	}
}

func TestHandleOverlayKeys_EscClosesNonTextOverlay(t *testing.T) {
	nonTextOverlays := []struct {
		name    string
		overlay OverlayState
	}{
		{"Help", OverlayHelp},
		{"ConfigDelete", OverlayConfigDelete},
	}
	for _, tt := range nonTextOverlays {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestModel()
			m.overlay = tt.overlay

			_, cmd := m.handleOverlayKeys(tea.KeyMsg{Type: tea.KeyEsc})

			if m.overlay != OverlayNone {
				t.Errorf("overlay = %d, want OverlayNone", m.overlay)
			}
			if cmd != nil {
				t.Errorf("expected nil cmd, got non-nil")
			}
		})
	}
}

func TestHandleOverlayKeys_TabCyclesCreateFields(t *testing.T) {
	m := newTestModel()
	m.overlay = OverlayCreateBucket
	m.storage.createField = 0

	// Cycle through all 8 fields (0 -> 1 -> 2 -> 3 -> 4 -> 5 -> 6 -> 7 -> 0)
	expected := []int{1, 2, 3, 4, 5, 6, 7, 0}
	for i, want := range expected {
		_, _ = m.handleOverlayKeys(tea.KeyMsg{Type: tea.KeyTab})
		if m.storage.createField != want {
			t.Errorf("step %d: createField = %d, want %d", i, m.storage.createField, want)
		}
	}
}

func TestCycleOption(t *testing.T) {
	options := []string{"", "STANDARD", "NEARLINE", "COLDLINE"}

	// Right cycles forward
	if got := cycleOption(options, "", "right"); got != "STANDARD" {
		t.Errorf("cycleOption(right) from empty = %q, want STANDARD", got)
	}
	if got := cycleOption(options, "STANDARD", "right"); got != "NEARLINE" {
		t.Errorf("cycleOption(right) from STANDARD = %q, want NEARLINE", got)
	}

	// Right wraps: COLDLINE -> ""
	if got := cycleOption(options, "COLDLINE", "right"); got != "" {
		t.Errorf("cycleOption(right) from COLDLINE = %q, want empty (wrap)", got)
	}

	// Left cycles backward: "" -> COLDLINE (wraps)
	if got := cycleOption(options, "", "left"); got != "COLDLINE" {
		t.Errorf("cycleOption(left) from empty = %q, want COLDLINE (wrap)", got)
	}

	// Unknown key returns current value unchanged
	if got := cycleOption(options, "STANDARD", "x"); got != "STANDARD" {
		t.Errorf("cycleOption(x) = %q, want STANDARD (unchanged)", got)
	}
}

func TestGetCreateFieldOptions(t *testing.T) {
	m := newTestModel()
	m.storage.availableProviders = []string{"gcp", "aws"}

	// Provider (1) returns dynamic options
	if opts := m.getCreateFieldOptions(1); len(opts) != 2 || opts[0] != "gcp" {
		t.Errorf("field 1 options = %v, want [gcp aws]", opts)
	}

	// Storage Class (3) returns static options
	if opts := m.getCreateFieldOptions(3); len(opts) != 5 || opts[1] != "STANDARD" {
		t.Errorf("field 3 options = %v, want 5 items starting with empty+STANDARD", opts)
	}

	// Free-text fields (0, 2, 4) return nil
	for _, field := range []int{0, 2, 4} {
		if opts := m.getCreateFieldOptions(field); opts != nil {
			t.Errorf("field %d should be free-text (nil options), got %v", field, opts)
		}
	}

	// Selector fields (5, 6, 7) return options
	for _, field := range []int{5, 6, 7} {
		if opts := m.getCreateFieldOptions(field); opts == nil {
			t.Errorf("field %d should be a selector, got nil options", field)
		}
	}
}

func TestGetCreateFieldValue_ReturnsCorrectValues(t *testing.T) {
	m := newTestModel()
	m.storage.createName = "my-bucket"
	m.storage.createProvider = "gcp"
	m.storage.createLocation = "us-central1"
	m.storage.createStorageClass = "STANDARD"
	m.storage.createLabels = "env=prod"
	m.storage.createVersioning = "yes"
	m.storage.createUniformAccess = "no"
	m.storage.createPublicAccessPrevention = "enforced"

	expected := []struct {
		field int
		want  string
	}{
		{0, "my-bucket"},
		{1, "gcp"},
		{2, "us-central1"},
		{3, "STANDARD"},
		{4, "env=prod"},
		{5, "yes"},
		{6, "no"},
		{7, "enforced"},
	}
	for _, tt := range expected {
		got := m.getCreateFieldValue(tt.field)
		if got != tt.want {
			t.Errorf("getCreateFieldValue(%d) = %q, want %q", tt.field, got, tt.want)
		}
	}
}

func TestGetCreateFieldValue_OutOfRange_ReturnsEmpty(t *testing.T) {
	m := newTestModel()
	if got := m.getCreateFieldValue(-1); got != "" {
		t.Errorf("getCreateFieldValue(-1) = %q, want empty", got)
	}
	if got := m.getCreateFieldValue(99); got != "" {
		t.Errorf("getCreateFieldValue(99) = %q, want empty", got)
	}
}

func TestSetCreateFieldValue_SetsCorrectFields(t *testing.T) {
	m := newTestModel()

	m.setCreateFieldValue(0, "bucket-a")
	m.setCreateFieldValue(1, "aws")
	m.setCreateFieldValue(2, "eu-west-1")
	m.setCreateFieldValue(3, "NEARLINE")
	m.setCreateFieldValue(4, "team=data")
	m.setCreateFieldValue(5, "no")
	m.setCreateFieldValue(6, "yes")
	m.setCreateFieldValue(7, "inherited")

	if m.storage.createName != "bucket-a" {
		t.Errorf("createName = %q, want bucket-a", m.storage.createName)
	}
	if m.storage.createProvider != "aws" {
		t.Errorf("createProvider = %q, want aws", m.storage.createProvider)
	}
	if m.storage.createLocation != "eu-west-1" {
		t.Errorf("createLocation = %q, want eu-west-1", m.storage.createLocation)
	}
	if m.storage.createStorageClass != "NEARLINE" {
		t.Errorf("createStorageClass = %q, want NEARLINE", m.storage.createStorageClass)
	}
	if m.storage.createLabels != "team=data" {
		t.Errorf("createLabels = %q, want team=data", m.storage.createLabels)
	}
	if m.storage.createVersioning != "no" {
		t.Errorf("createVersioning = %q, want no", m.storage.createVersioning)
	}
	if m.storage.createUniformAccess != "yes" {
		t.Errorf("createUniformAccess = %q, want yes", m.storage.createUniformAccess)
	}
	if m.storage.createPublicAccessPrevention != "inherited" {
		t.Errorf("createPublicAccessPrevention = %q, want inherited", m.storage.createPublicAccessPrevention)
	}
}

func TestSyncTextInputToField_SkipsSelectorFields(t *testing.T) {
	m := newTestModel()
	m.overlay = OverlayCreateBucket
	m.storage.availableProviders = []string{"gcp", "aws"}
	m.storage.createProvider = "gcp"
	m.storage.createStorageClass = "STANDARD"

	// Selector field (1 = Provider): textinput value should NOT overwrite state.
	m.storage.createField = 1
	m.textInput.SetValue("something-typed")
	m.syncTextInputToField()
	if m.storage.createProvider != "gcp" {
		t.Errorf("selector field was overwritten: createProvider = %q, want gcp", m.storage.createProvider)
	}

	// Selector field (3 = StorageClass): same behavior.
	m.storage.createField = 3
	m.textInput.SetValue("INVALID")
	m.syncTextInputToField()
	if m.storage.createStorageClass != "STANDARD" {
		t.Errorf("selector field was overwritten: createStorageClass = %q, want STANDARD", m.storage.createStorageClass)
	}
}

func TestSyncTextInputToField_SyncsFreeTextFields(t *testing.T) {
	m := newTestModel()
	m.overlay = OverlayCreateBucket

	// Free-text field (0 = Name): textinput value should sync to state.
	m.storage.createField = 0
	m.textInput.SetValue("new-bucket")
	m.syncTextInputToField()
	if m.storage.createName != "new-bucket" {
		t.Errorf("free-text field not synced: createName = %q, want new-bucket", m.storage.createName)
	}

	// Free-text field (4 = Labels): same behavior.
	m.storage.createField = 4
	m.textInput.SetValue("env=staging")
	m.syncTextInputToField()
	if m.storage.createLabels != "env=staging" {
		t.Errorf("free-text field not synced: createLabels = %q, want env=staging", m.storage.createLabels)
	}
}

func TestHandleOverlayKeys_SelectorFieldBlocksFreeText(t *testing.T) {
	m := newTestModel()
	m.overlay = OverlayCreateBucket
	m.storage.availableProviders = []string{"gcp", "aws"}
	m.storage.createProvider = "gcp"
	m.storage.createField = 1 // Provider — selector field

	// Typing a character should not change the provider value.
	_, _ = m.handleOverlayKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if m.storage.createProvider != "gcp" {
		t.Errorf("typing on selector changed value: createProvider = %q, want gcp", m.storage.createProvider)
	}
}

func TestHandleOverlayKeys_SelectorFieldCyclesOnArrows(t *testing.T) {
	m := newTestModel()
	m.overlay = OverlayCreateBucket
	m.storage.createField = 3 // StorageClass — selector field
	m.storage.createStorageClass = ""

	// Right arrow: "" -> STANDARD
	_, _ = m.handleOverlayKeys(tea.KeyMsg{Type: tea.KeyRight})
	if m.storage.createStorageClass != "STANDARD" {
		t.Errorf("right on storage class = %q, want STANDARD", m.storage.createStorageClass)
	}

	// Right again: STANDARD -> NEARLINE
	_, _ = m.handleOverlayKeys(tea.KeyMsg{Type: tea.KeyRight})
	if m.storage.createStorageClass != "NEARLINE" {
		t.Errorf("right on storage class = %q, want NEARLINE", m.storage.createStorageClass)
	}

	// Left: NEARLINE -> STANDARD
	_, _ = m.handleOverlayKeys(tea.KeyMsg{Type: tea.KeyLeft})
	if m.storage.createStorageClass != "STANDARD" {
		t.Errorf("left on storage class = %q, want STANDARD", m.storage.createStorageClass)
	}
}

func TestHandleOverlayKeys_TabCyclesConfigAddFields(t *testing.T) {
	m := newTestModel()
	m.overlay = OverlayConfigAdd
	m.storage.createField = 0
	m.config.editKey = ""
	m.config.editValue = ""

	// Field 0 -> field 1
	_, _ = m.handleOverlayKeys(tea.KeyMsg{Type: tea.KeyTab})
	if m.storage.createField != 1 {
		t.Errorf("createField = %d, want 1 after first tab", m.storage.createField)
	}

	// Field 1 -> field 0
	_, _ = m.handleOverlayKeys(tea.KeyMsg{Type: tea.KeyTab})
	if m.storage.createField != 0 {
		t.Errorf("createField = %d, want 0 after second tab", m.storage.createField)
	}
}

// --- handleOverlaySubmit ---

func TestHandleOverlaySubmit_CreateBucket_EmptyFields(t *testing.T) {
	tests := []struct {
		name       string
		bucketName string
		provider   string
		location   string
	}{
		{"all_empty", "", "", ""},
		{"name_only", "my-bucket", "", ""},
		{"missing_location", "my-bucket", "gcp", ""},
		{"missing_provider", "my-bucket", "", "us-central1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestModel()
			m.overlay = OverlayCreateBucket
			m.storage.createName = tt.bucketName
			m.storage.createProvider = tt.provider
			m.storage.createLocation = tt.location

			_, cmd := m.handleOverlaySubmit()
			if cmd != nil {
				t.Error("expected nil cmd when required fields are empty")
			}
		})
	}
}

func TestHandleOverlaySubmit_CreateBucket_AllFieldsSet(t *testing.T) {
	m := newTestModel()
	m.overlay = OverlayCreateBucket
	m.storage.createName = "new-bucket"
	m.storage.createProvider = "gcp"
	m.storage.createLocation = "us-central1"
	// syncTextInputToField reads the textinput into the active field (field 0 = name),
	// so the textinput must hold the name value for submission to succeed.
	m.storage.createField = 0
	m.textInput.SetValue("new-bucket")

	_, cmd := m.handleOverlaySubmit()
	if cmd == nil {
		t.Error("expected non-nil cmd when all fields are provided")
	}
	if m.overlay != OverlayNone {
		t.Errorf("overlay = %d, want OverlayNone after submit", m.overlay)
	}
	if !m.storage.loading {
		t.Error("expected loading = true after submit")
	}
}

func TestHandleOverlaySubmit_DeleteConfirm_WrongInput(t *testing.T) {
	m := newTestModel()
	m.overlay = OverlayDeleteConfirm
	m.storage.selectedBucket = storage.Bucket{Name: "my-bucket"}
	m.textInput.SetValue("wrong-name")

	_, cmd := m.handleOverlaySubmit()
	if cmd != nil {
		t.Error("expected nil cmd when confirmation input does not match bucket name")
	}
}

func TestHandleOverlaySubmit_DeleteConfirm_CorrectInput(t *testing.T) {
	m := newTestModel()
	m.overlay = OverlayDeleteConfirm
	m.storage.selectedBucket = storage.Bucket{Name: "my-bucket", Provider: domain.GCP}
	m.textInput.SetValue("my-bucket")

	_, cmd := m.handleOverlaySubmit()
	if cmd == nil {
		t.Error("expected non-nil cmd when confirmation input matches bucket name")
	}
	if m.overlay != OverlayNone {
		t.Errorf("overlay = %d, want OverlayNone after confirmed delete", m.overlay)
	}
	if !m.storage.loading {
		t.Error("expected loading = true after confirmed delete")
	}
}

func TestHandleOverlaySubmit_ConfigAdd_EmptyFields(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
	}{
		{"both_empty", "", ""},
		{"key_only", "some.key", ""},
		{"value_only", "", "some-value"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestModel()
			m.overlay = OverlayConfigAdd
			m.config.editKey = tt.key
			m.config.editValue = tt.value

			_, cmd := m.handleOverlaySubmit()
			if cmd != nil {
				t.Error("expected nil cmd when config add fields are empty")
			}
		})
	}
}

func TestParseLabels(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  map[string]string
	}{
		{"empty", "", nil},
		{"single", "env=prod", map[string]string{"env": "prod"}},
		{"multiple", "env=prod,team=data", map[string]string{"env": "prod", "team": "data"}},
		{"spaces", " env = prod , team = data ", map[string]string{"env": "prod", "team": "data"}},
		{"no_equals", "invalid", nil},
		{"trailing_comma", "env=prod,", map[string]string{"env": "prod"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseLabels(tt.input)
			if tt.want == nil {
				if got != nil {
					t.Errorf("parseLabels(%q) = %v, want nil", tt.input, got)
				}
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("parseLabels(%q) has %d entries, want %d", tt.input, len(got), len(tt.want))
				return
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("parseLabels(%q)[%q] = %q, want %q", tt.input, k, got[k], v)
				}
			}
		})
	}
}

func TestHandleOverlaySubmit_CreateBucket_WithOptionalFields(t *testing.T) {
	m := newTestModel()
	m.overlay = OverlayCreateBucket
	m.storage.createName = "new-bucket"
	m.storage.createProvider = "gcp"
	m.storage.createLocation = "us-central1"
	m.storage.createStorageClass = "nearline"
	m.storage.createLabels = "env=prod,team=data"
	m.storage.createVersioning = "yes"
	m.storage.createUniformAccess = "yes"
	m.storage.createPublicAccessPrevention = "enforced"
	m.storage.createField = 0
	m.textInput.SetValue("new-bucket")

	_, cmd := m.handleOverlaySubmit()
	if cmd == nil {
		t.Error("expected non-nil cmd when all fields are provided")
	}
	if m.overlay != OverlayNone {
		t.Errorf("overlay = %d, want OverlayNone after submit", m.overlay)
	}
}

// --- Data message handlers ---

func TestHandleBucketsLoaded_Success(t *testing.T) {
	m := newTestModel()
	m.storage.loading = true
	m.storage.cursor = 5

	buckets := []storage.Bucket{{Name: "a"}, {Name: "b"}}
	_, _ = m.handleBucketsLoaded(BucketsLoadedMsg{Buckets: buckets})

	if m.storage.loading {
		t.Error("expected loading = false")
	}
	if !m.storage.loaded {
		t.Error("expected loaded = true")
	}
	if len(m.storage.buckets) != 2 {
		t.Errorf("buckets count = %d, want 2", len(m.storage.buckets))
	}
	if m.storage.cursor != 0 {
		t.Errorf("cursor = %d, want 0 (reset on load)", m.storage.cursor)
	}
	if m.storage.scrollOffset != 0 {
		t.Errorf("scrollOffset = %d, want 0", m.storage.scrollOffset)
	}
	if m.err != nil {
		t.Errorf("err = %v, want nil", m.err)
	}
}

func TestHandleBucketsLoaded_Error(t *testing.T) {
	m := newTestModel()
	m.storage.loading = true

	loadErr := errors.New("network timeout")
	_, _ = m.handleBucketsLoaded(BucketsLoadedMsg{Err: loadErr})

	if m.storage.loading {
		t.Error("expected loading = false")
	}
	if !m.storage.loaded {
		t.Error("expected loaded = true even on error")
	}
	if m.err == nil {
		t.Fatal("expected non-nil error")
	}
	if m.err.Error() != "network timeout" {
		t.Errorf("err = %q, want %q", m.err.Error(), "network timeout")
	}
}

func TestHandleBucketCreated_Success(t *testing.T) {
	m := newTestModel()
	m.storage.loading = true

	_, cmd := m.handleBucketCreated(BucketCreatedMsg{Err: nil})

	if m.statusMessage != "Bucket created successfully" {
		t.Errorf("statusMessage = %q, want %q", m.statusMessage, "Bucket created successfully")
	}
	if cmd == nil {
		t.Error("expected non-nil cmd (batch of refetch + clear status)")
	}
}

func TestHandleBucketCreated_Error(t *testing.T) {
	m := newTestModel()

	_, cmd := m.handleBucketCreated(BucketCreatedMsg{Err: errors.New("permission denied")})

	if m.err == nil || m.err.Error() != "permission denied" {
		t.Errorf("err = %v, want 'permission denied'", m.err)
	}
	if cmd != nil {
		t.Error("expected nil cmd on error")
	}
}

func TestHandleBucketDeleted_Success(t *testing.T) {
	m := newTestModel()
	m.viewState = ViewStorageBucketDetail

	_, cmd := m.handleBucketDeleted(BucketDeletedMsg{Err: nil})

	if m.viewState != ViewStorageList {
		t.Errorf("viewState = %d, want ViewStorageList (%d)", m.viewState, ViewStorageList)
	}
	if m.statusMessage != "Bucket deleted successfully" {
		t.Errorf("statusMessage = %q, want %q", m.statusMessage, "Bucket deleted successfully")
	}
	if cmd == nil {
		t.Error("expected non-nil cmd (batch of refetch + clear status)")
	}
}

func TestHandleBucketDeleted_Error(t *testing.T) {
	m := newTestModel()
	m.viewState = ViewStorageBucketDetail

	_, _ = m.handleBucketDeleted(BucketDeletedMsg{Err: errors.New("not found")})

	if m.err == nil || m.err.Error() != "not found" {
		t.Errorf("err = %v, want 'not found'", m.err)
	}
	// viewState should not change on error
	if m.viewState != ViewStorageBucketDetail {
		t.Errorf("viewState should remain unchanged on error, got %d", m.viewState)
	}
}

func TestHandleProviderRemoved_InvalidatesCaches(t *testing.T) {
	m := newTestModelWithConfig(t)
	m.storage.loaded = true
	m.sql.loaded = true

	_, cmd := m.handleProviderRemoved(ProviderRemovedMsg{Err: nil})

	if m.storage.loaded {
		t.Error("expected storage.loaded = false after provider removal")
	}
	if m.sql.loaded {
		t.Error("expected sql.loaded = false after provider removal")
	}
	if m.statusMessage != "Provider removed" {
		t.Errorf("statusMessage = %q, want %q", m.statusMessage, "Provider removed")
	}
	if m.config.cursor != 0 {
		t.Errorf("config.cursor = %d, want 0 (reset after removal)", m.config.cursor)
	}
	if cmd == nil {
		t.Error("expected non-nil cmd (batch of refetch + clear status)")
	}
}

func TestHandleProviderRemoved_Error(t *testing.T) {
	m := newTestModel()
	m.storage.loaded = true
	m.sql.loaded = true

	_, cmd := m.handleProviderRemoved(ProviderRemovedMsg{Err: errors.New("config write failed")})

	// Caches should NOT be invalidated on error
	if !m.storage.loaded {
		t.Error("storage.loaded should remain true on error")
	}
	if !m.sql.loaded {
		t.Error("sql.loaded should remain true on error")
	}
	if m.err == nil || m.err.Error() != "config write failed" {
		t.Errorf("err = %v, want 'config write failed'", m.err)
	}
	if cmd != nil {
		t.Error("expected nil cmd on error")
	}
}

// --- switchTab ---

func TestSwitchTab_LazyLoad(t *testing.T) {
	tests := []struct {
		name      string
		tab       Tab
		wantView  ViewState
		preloaded bool
	}{
		{"sql_first_visit", TabSql, ViewSqlList, false},
		{"sql_revisit", TabSql, ViewSqlList, true},
		{"config_first_visit", TabConfig, ViewConfigList, false},
		{"config_revisit", TabConfig, ViewConfigList, true},
		{"storage_first_visit", TabStorage, ViewStorageList, false},
		{"storage_revisit", TabStorage, ViewStorageList, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestModel()
			// Pre-load tabs as needed
			switch tt.tab {
			case TabSql:
				m.sql.loaded = tt.preloaded
			case TabConfig:
				m.config.loaded = tt.preloaded
			case TabStorage:
				m.storage.loaded = tt.preloaded
			}

			_, cmd := m.switchTab(tt.tab)

			if m.activeTab != tt.tab {
				t.Errorf("activeTab = %d, want %d", m.activeTab, tt.tab)
			}
			if m.viewState != tt.wantView {
				t.Errorf("viewState = %d, want %d", m.viewState, tt.wantView)
			}
			if tt.preloaded && cmd != nil {
				t.Error("expected nil cmd for already-loaded tab")
			}
			if !tt.preloaded && cmd == nil {
				t.Error("expected non-nil cmd for first visit to tab")
			}
		})
	}
}

func TestSwitchTab_ClearsError(t *testing.T) {
	m := newTestModel()
	m.err = errors.New("previous error")
	m.storage.loaded = true

	_, _ = m.switchTab(TabStorage)

	if m.err != nil {
		t.Errorf("err = %v, want nil (switchTab should clear errors)", m.err)
	}
}

// --- handleInstancesLoaded ---

func TestHandleInstancesLoaded_Success(t *testing.T) {
	m := newTestModel()
	m.sql.loading = true

	_, _ = m.handleInstancesLoaded(InstancesLoadedMsg{
		Instances: []domainsql.Instance{{Name: "db-1"}, {Name: "db-2"}},
	})

	if m.sql.loading {
		t.Error("expected loading = false")
	}
	if !m.sql.loaded {
		t.Error("expected loaded = true")
	}
	if len(m.sql.instances) != 2 {
		t.Errorf("instances count = %d, want 2", len(m.sql.instances))
	}
	if m.sql.cursor != 0 {
		t.Errorf("cursor = %d, want 0", m.sql.cursor)
	}
}

func TestHandleInstancesLoaded_Error(t *testing.T) {
	m := newTestModel()
	m.sql.loading = true

	_, _ = m.handleInstancesLoaded(InstancesLoadedMsg{Err: errors.New("sql error")})

	if m.sql.loading {
		t.Error("expected loading = false")
	}
	if m.err == nil || m.err.Error() != "sql error" {
		t.Errorf("err = %v, want 'sql error'", m.err)
	}
}

// --- handleConfigLoaded ---

func TestHandleConfigLoaded_Success(t *testing.T) {
	m := newTestModel()
	m.config.loading = true
	m.config.cursor = 3

	entries := []ui.ConfigEntry{{Key: "gcp.project", Value: "my-proj"}}
	_, _ = m.handleConfigLoaded(ConfigLoadedMsg{Entries: entries})

	if m.config.loading {
		t.Error("expected loading = false")
	}
	if !m.config.loaded {
		t.Error("expected loaded = true")
	}
	if len(m.config.entries) != 1 {
		t.Errorf("entries count = %d, want 1", len(m.config.entries))
	}
	if m.config.cursor != 0 {
		t.Errorf("cursor = %d, want 0 (reset on load)", m.config.cursor)
	}
}

func TestHandleConfigLoaded_Error(t *testing.T) {
	m := newTestModel()
	m.config.loading = true

	_, _ = m.handleConfigLoaded(ConfigLoadedMsg{Err: errors.New("read error")})

	if m.config.loading {
		t.Error("expected loading = false")
	}
	if m.err == nil || m.err.Error() != "read error" {
		t.Errorf("err = %v, want 'read error'", m.err)
	}
}
