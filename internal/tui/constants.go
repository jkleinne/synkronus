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

// UI option strings.
const (
	optionYes          = "yes"
	optionNo           = "no"
	defaultDownloadDir = "./"
)
