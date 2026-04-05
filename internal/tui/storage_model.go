package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"synkronus/internal/domain/storage"
	"synkronus/internal/provider/storage/shared"
	"synkronus/internal/service"
	"synkronus/internal/tui/ui"
)

// StorageModel owns the mutable state and key/message handling for the Storage tab.
type StorageModel struct {
	buckets        []storage.Bucket
	objects        storage.ObjectList
	selectedBucket storage.Bucket
	selectedObject storage.Object
	cursor         int
	scrollOffset   int
	loading        bool
	loaded         bool

	// Create-bucket form fields
	createName                   string
	createProvider               string
	createLocation               string
	availableProviders           []string
	createStorageClass           string
	createLabels                 string
	createVersioning             string // "yes"/"no"/""
	createUniformAccess          string // "yes"/"no"/""
	createPublicAccessPrevention string // "enforced"/"inherited"/""
	createFieldIndex             int
	createHiddenFields           map[int]bool

	// Delete state
	deleteInput     string
	deleteKind      deleteTarget
	deleteObjectKey string

	// Download/upload state
	downloadingKey string
	downloadDir    string
	uploadFilePath string
	uploadObjectKey string
	uploadField     int
}

// --- Key handlers ---

// HandleListKeys handles keystrokes on the bucket list view.
// It returns updated root-level fields (viewState, overlay, err) via the returned ViewUpdate.
func (s *StorageModel) HandleListKeys(msg tea.KeyMsg, svc *service.StorageService, deps *Deps) (ViewUpdate, tea.Cmd) {
	key := msg.String()
	switch key {
	case "j", keyDown:
		if len(s.buckets) > 0 {
			s.cursor = min(s.cursor+1, len(s.buckets)-1)
		}
	case "k", keyUp:
		if s.cursor > 0 {
			s.cursor--
		}
	case keyEnter:
		if len(s.buckets) > 0 {
			s.selectedBucket = s.buckets[s.cursor]
			return ViewUpdate{ViewState: ptrViewState(ViewStorageBucketDetail)}, fetchBucketDetailCmd(
				svc,
				s.selectedBucket.Name,
				strings.ToLower(string(s.selectedBucket.Provider)),
			)
		}
	case "o":
		if len(s.buckets) > 0 {
			s.selectedBucket = s.buckets[s.cursor]
			s.cursor = 0
			s.scrollOffset = 0
			s.loading = true
			return ViewUpdate{ViewState: ptrViewState(ViewStorageObjectList), ClearErr: true}, fetchObjectsCmd(
				svc,
				s.selectedBucket.Name,
				strings.ToLower(string(s.selectedBucket.Provider)),
				"",
			)
		}
	case "c":
		s.resetCreateForm(deps)
		return ViewUpdate{Overlay: ptrOverlay(OverlayCreateBucket), FocusTextInput: true, TextInputValue: ptrString("")}, nil
	case "d":
		if len(s.buckets) > 0 {
			s.selectedBucket = s.buckets[s.cursor]
			s.deleteKind = deleteTargetBucket
			s.deleteInput = ""
			return ViewUpdate{Overlay: ptrOverlay(OverlayDeleteConfirm), FocusTextInput: true, TextInputValue: ptrString("")}, nil
		}
	case "r":
		s.loading = true
		return ViewUpdate{ClearErr: true}, fetchBucketsCmd(svc, deps.Factory)
	case keyTab:
		return ViewUpdate{SwitchTab: ptrTab(TabStorage.Next())}, nil
	case keyShiftTab:
		return ViewUpdate{SwitchTab: ptrTab(TabStorage.Prev())}, nil
	}
	return ViewUpdate{}, nil
}

// HandleBucketDetailKeys handles keystrokes on the bucket detail view.
func (s *StorageModel) HandleBucketDetailKeys(msg tea.KeyMsg, svc *service.StorageService) (ViewUpdate, tea.Cmd) {
	key := msg.String()
	switch key {
	case keyEsc:
		return ViewUpdate{ViewState: ptrViewState(ViewStorageList), ClearErr: true}, nil
	case "o":
		s.loading = true
		s.cursor = 0
		s.scrollOffset = 0
		return ViewUpdate{ViewState: ptrViewState(ViewStorageObjectList), ClearErr: true}, fetchObjectsCmd(
			svc,
			s.selectedBucket.Name,
			strings.ToLower(string(s.selectedBucket.Provider)),
			"",
		)
	case "d":
		s.deleteKind = deleteTargetBucket
		s.deleteInput = ""
		return ViewUpdate{Overlay: ptrOverlay(OverlayDeleteConfirm), FocusTextInput: true, TextInputValue: ptrString("")}, nil
	case "j", keyDown:
		s.scrollOffset++
	case "k", keyUp:
		if s.scrollOffset > 0 {
			s.scrollOffset--
		}
	}
	return ViewUpdate{}, nil
}

// HandleObjectListKeys handles keystrokes on the object list view.
func (s *StorageModel) HandleObjectListKeys(msg tea.KeyMsg, svc *service.StorageService) (ViewUpdate, tea.Cmd) {
	key := msg.String()
	totalItems := len(s.objects.Objects) + len(s.objects.CommonPrefixes)

	switch key {
	case "j", keyDown:
		if totalItems > 0 {
			s.cursor = min(s.cursor+1, totalItems-1)
		}
	case "k", keyUp:
		if s.cursor > 0 {
			s.cursor--
		}
	case keyEnter:
		if totalItems > 0 {
			prefixCount := len(s.objects.CommonPrefixes)
			if s.cursor < prefixCount {
				prefix := s.objects.CommonPrefixes[s.cursor]
				s.loading = true
				s.cursor = 0
				s.scrollOffset = 0
				return ViewUpdate{ClearErr: true}, fetchObjectsCmd(
					svc,
					s.selectedBucket.Name,
					strings.ToLower(string(s.selectedBucket.Provider)),
					prefix,
				)
			}
			objIdx := s.cursor - prefixCount
			if objIdx < len(s.objects.Objects) {
				s.selectedObject = s.objects.Objects[objIdx]
				s.loading = true
				return ViewUpdate{ViewState: ptrViewState(ViewStorageObjectDetail), ClearErr: true}, fetchObjectDetailCmd(
					svc,
					s.selectedBucket.Name,
					s.selectedObject.Key,
					strings.ToLower(string(s.selectedBucket.Provider)),
				)
			}
		}
	case "w":
		if totalItems > 0 {
			prefixCount := len(s.objects.CommonPrefixes)
			if s.cursor >= prefixCount {
				objIdx := s.cursor - prefixCount
				if objIdx < len(s.objects.Objects) {
					obj := s.objects.Objects[objIdx]
					s.downloadingKey = obj.Key
					s.downloadDir = defaultDownloadDir
					return ViewUpdate{
						Overlay:        ptrOverlay(OverlayDownloadPath),
						FocusTextInput: true,
						TextInputValue: ptrString(defaultDownloadDir),
					}, nil
				}
			}
		}
	case "u":
		s.uploadFilePath = ""
		s.uploadObjectKey = ""
		s.uploadField = 0
		return ViewUpdate{Overlay: ptrOverlay(OverlayUploadObject), FocusTextInput: true, TextInputValue: ptrString("")}, nil
	case "d":
		if totalItems > 0 {
			prefixCount := len(s.objects.CommonPrefixes)
			if s.cursor >= prefixCount {
				objIdx := s.cursor - prefixCount
				if objIdx < len(s.objects.Objects) {
					obj := s.objects.Objects[objIdx]
					s.deleteObjectKey = obj.Key
					s.deleteKind = deleteTargetObject
					s.deleteInput = ""
					return ViewUpdate{Overlay: ptrOverlay(OverlayDeleteConfirm), FocusTextInput: true, TextInputValue: ptrString("")}, nil
				}
			}
		}
	case keyEsc:
		s.cursor = 0
		s.scrollOffset = 0
		return ViewUpdate{ViewState: ptrViewState(ViewStorageBucketDetail), ClearErr: true}, nil
	}
	return ViewUpdate{}, nil
}

// HandleObjectDetailKeys handles keystrokes on the object detail view.
func (s *StorageModel) HandleObjectDetailKeys(msg tea.KeyMsg) (ViewUpdate, tea.Cmd) {
	key := msg.String()
	switch key {
	case keyEsc:
		return ViewUpdate{ViewState: ptrViewState(ViewStorageObjectList), ClearErr: true}, nil
	case "w":
		s.downloadingKey = s.selectedObject.Key
		s.downloadDir = defaultDownloadDir
		return ViewUpdate{
			Overlay:        ptrOverlay(OverlayDownloadPath),
			FocusTextInput: true,
			TextInputValue: ptrString(defaultDownloadDir),
		}, nil
	case "d":
		s.deleteObjectKey = s.selectedObject.Key
		s.deleteKind = deleteTargetObject
		s.deleteInput = ""
		return ViewUpdate{Overlay: ptrOverlay(OverlayDeleteConfirm), FocusTextInput: true, TextInputValue: ptrString("")}, nil
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

// HandleBucketsLoaded processes a completed bucket list fetch.
func (s *StorageModel) HandleBucketsLoaded(msg BucketsLoadedMsg) error {
	s.loading = false
	s.loaded = true
	if msg.Err != nil {
		return msg.Err
	}
	s.buckets = msg.Buckets
	s.cursor = 0
	s.scrollOffset = 0
	return nil
}

// HandleBucketDetailLoaded processes a completed bucket detail fetch.
func (s *StorageModel) HandleBucketDetailLoaded(msg BucketDetailMsg) error {
	s.loading = false
	if msg.Err != nil {
		return msg.Err
	}
	s.selectedBucket = msg.Bucket
	return nil
}

// HandleObjectsLoaded processes a completed object list fetch.
func (s *StorageModel) HandleObjectsLoaded(msg ObjectsLoadedMsg) error {
	s.loading = false
	if msg.Err != nil {
		return msg.Err
	}
	s.objects = msg.Objects
	s.cursor = 0
	s.scrollOffset = 0
	return nil
}

// HandleObjectDetailLoaded processes a completed object detail fetch.
func (s *StorageModel) HandleObjectDetailLoaded(msg ObjectDetailMsg) error {
	s.loading = false
	if msg.Err != nil {
		return msg.Err
	}
	s.selectedObject = msg.Object
	return nil
}

// HandleBucketCreated processes a completed bucket creation.
func (s *StorageModel) HandleBucketCreated(msg BucketCreatedMsg, svc *service.StorageService, deps *Deps) (string, tea.Cmd) {
	s.loading = false
	if msg.Err != nil {
		return "", nil
	}
	statusMsg := "Bucket created successfully"
	if len(msg.Warnings) > 0 {
		statusMsg += fmt.Sprintf(" (%d warning(s): %s)", len(msg.Warnings), strings.Join(msg.Warnings, "; "))
	}
	s.loading = true
	return statusMsg, tea.Batch(fetchBucketsCmd(svc, deps.Factory), clearStatusCmd())
}

// HandleBucketDeleted processes a completed bucket deletion.
func (s *StorageModel) HandleBucketDeleted(msg BucketDeletedMsg, svc *service.StorageService, deps *Deps) (string, ViewState, tea.Cmd) {
	s.loading = false
	if msg.Err != nil {
		return "", 0, nil
	}
	s.loading = true
	return "Bucket deleted successfully", ViewStorageList, tea.Batch(fetchBucketsCmd(svc, deps.Factory), clearStatusCmd())
}

// HandleObjectDownloaded processes a completed object download.
func (s *StorageModel) HandleObjectDownloaded(msg ObjectDownloadedMsg) (string, tea.Cmd) {
	s.loading = false
	s.downloadingKey = ""
	if msg.Err != nil {
		return "", nil
	}
	return fmt.Sprintf("Downloaded to %s", msg.FilePath), clearStatusCmd()
}

// HandleObjectUploaded processes a completed object upload.
func (s *StorageModel) HandleObjectUploaded(msg ObjectUploadedMsg, svc *service.StorageService) (string, tea.Cmd) {
	s.loading = false
	if msg.Err != nil {
		return "", nil
	}
	statusMsg := fmt.Sprintf("Uploaded %s", s.uploadObjectKey)
	s.loading = true
	return statusMsg, tea.Batch(
		fetchObjectsCmd(svc, s.selectedBucket.Name,
			strings.ToLower(string(s.selectedBucket.Provider)),
			s.objects.Prefix),
		clearStatusCmd(),
	)
}

// HandleObjectDeleted processes a completed object deletion.
func (s *StorageModel) HandleObjectDeleted(msg ObjectDeletedMsg, svc *service.StorageService) (string, tea.Cmd) {
	s.loading = false
	if msg.Err != nil {
		return "", nil
	}
	statusMsg := fmt.Sprintf("Object '%s' deleted", s.deleteObjectKey)
	s.loading = true
	return statusMsg, tea.Batch(
		fetchObjectsCmd(svc, s.selectedBucket.Name,
			strings.ToLower(string(s.selectedBucket.Provider)),
			s.objects.Prefix),
		clearStatusCmd(),
	)
}

// --- View rendering ---

// RenderContent returns the rendered view for storage-related viewStates.
func (s *StorageModel) RenderContent(viewState ViewState, spinnerView string, width int) string {
	switch viewState {
	case ViewStorageList:
		if s.loading {
			return ui.CenterContent(ui.RenderSpinnerView(spinnerView, "Loading buckets..."), width)
		}
		return ui.RenderBucketList(s.buckets, s.cursor, s.scrollOffset, width)

	case ViewStorageBucketDetail:
		if s.loading {
			return ui.CenterContent(ui.RenderSpinnerView(spinnerView, "Loading bucket details..."), width)
		}
		return ui.RenderBucketDetail(s.selectedBucket, width)

	case ViewStorageObjectList:
		if s.loading {
			spinnerMsg := "Loading objects..."
			if s.downloadingKey != "" {
				spinnerMsg = fmt.Sprintf("Downloading %s...", s.downloadingKey)
			}
			return ui.CenterContent(ui.RenderSpinnerView(spinnerView, spinnerMsg), width)
		}
		return ui.RenderObjectList(s.objects, s.cursor, s.scrollOffset, width)

	case ViewStorageObjectDetail:
		if s.loading {
			spinnerMsg := "Loading object details..."
			if s.downloadingKey != "" {
				spinnerMsg = fmt.Sprintf("Downloading %s...", s.downloadingKey)
			}
			return ui.CenterContent(ui.RenderSpinnerView(spinnerView, spinnerMsg), width)
		}
		return ui.RenderObjectDetail(s.selectedObject, width)

	default:
		return ""
	}
}

// --- Create-bucket form helpers ---

// resetCreateForm clears all create-bucket form fields and sets provider defaults.
func (s *StorageModel) resetCreateForm(deps *Deps) {
	s.createName = ""
	s.createLocation = ""
	s.createStorageClass = ""
	s.createLabels = ""
	s.createVersioning = ""
	s.createUniformAccess = ""
	s.createPublicAccessPrevention = ""
	s.availableProviders = deps.Factory.GetConfiguredProviders()
	if len(s.availableProviders) > 0 {
		s.createProvider = s.availableProviders[0]
	} else {
		s.createProvider = ""
	}
	s.createFieldIndex = 0
	s.updateCreateHiddenFields()
}

// getCreateFieldOptions returns the valid options for a selector field, or nil for free-text fields.
func (s *StorageModel) getCreateFieldOptions(field int) []string {
	switch field {
	case createFieldProvider:
		return s.availableProviders
	case createFieldStorageClass:
		return storageClassOptions
	case createFieldVersioning:
		return versioningOptions
	case createFieldUniformAccess:
		return uniformAccessOptions
	case createFieldPublicAccessPrevention:
		return publicAccessPreventionOptions
	default:
		return nil
	}
}

// getCreateFieldValue returns the current value of a create-bucket form field.
func (s *StorageModel) getCreateFieldValue(field int) string {
	switch field {
	case createFieldName:
		return s.createName
	case createFieldProvider:
		return s.createProvider
	case createFieldLocation:
		return s.createLocation
	case createFieldStorageClass:
		return s.createStorageClass
	case createFieldLabels:
		return s.createLabels
	case createFieldVersioning:
		return s.createVersioning
	case createFieldUniformAccess:
		return s.createUniformAccess
	case createFieldPublicAccessPrevention:
		return s.createPublicAccessPrevention
	default:
		return ""
	}
}

// setCreateFieldValue sets the value of a create-bucket form field.
func (s *StorageModel) setCreateFieldValue(field int, value string) {
	switch field {
	case createFieldName:
		s.createName = value
	case createFieldProvider:
		s.createProvider = value
	case createFieldLocation:
		s.createLocation = value
	case createFieldStorageClass:
		s.createStorageClass = value
	case createFieldLabels:
		s.createLabels = value
	case createFieldVersioning:
		s.createVersioning = value
	case createFieldUniformAccess:
		s.createUniformAccess = value
	case createFieldPublicAccessPrevention:
		s.createPublicAccessPrevention = value
	}
}

// nextVisibleCreateField returns the next visible field index, skipping hidden fields.
func (s *StorageModel) nextVisibleCreateField(current int) int {
	for i := 1; i <= createFormFieldCount; i++ {
		next := (current + i) % createFormFieldCount
		if !s.createHiddenFields[next] {
			return next
		}
	}
	return current
}

// updateCreateHiddenFields shows/hides provider-specific fields based on the selected provider.
func (s *StorageModel) updateCreateHiddenFields() {
	if s.createHiddenFields == nil {
		s.createHiddenFields = make(map[int]bool)
	}
	s.createHiddenFields[createFieldUniformAccess] = !shared.SupportsOption(s.createProvider, "uniform-access")
}
