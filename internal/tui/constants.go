package tui

import "time"

// Create-bucket form field indices.
const (
	createFieldName                   = 0
	createFieldProvider               = 1
	createFieldLocation               = 2
	createFieldStorageClass           = 3
	createFieldLabels                 = 4
	createFieldVersioning             = 5
	createFieldUniformAccess          = 6
	createFieldPublicAccessPrevention = 7
)

// Timeouts and durations.
const (
	transferTimeout       = 5 * time.Minute
	statusDisplayDuration = 3 * time.Second
)

// textInputCharLimit is the maximum character limit for text input fields.
const textInputCharLimit = 256

// createFormFieldCount is the total number of fields in the create-bucket form.
const createFormFieldCount = 8

// Static options for selector fields in the create-bucket form.
// Empty string at index 0 represents "unset" (use provider default).
var (
	storageClassOptions           = []string{"", "STANDARD", "NEARLINE", "COLDLINE", "ARCHIVE"}
	versioningOptions             = []string{"", "yes", "no"}
	uniformAccessOptions          = []string{"", "yes", "no"}
	publicAccessPreventionOptions = []string{"", "enforced", "inherited"}
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

// UI option strings.
const (
	optionYes          = "yes"
	optionNo           = "no"
	defaultDownloadDir = "./"
)
