// Package shared provides constants and utilities used by multiple storage
// provider implementations. These live at the provider layer — not domain —
// because they represent provider-facing display values and I/O helpers.
package shared

import (
	"mime"
	"path/filepath"
)

// Display-facing constants returned by provider mappers. Distinct from the
// lowercase domain constants (e.g. storage.PublicAccessPreventionEnforced)
// which are used for CLI input validation.
const (
	PublicAccessEnforced  = "Enforced"
	PublicAccessInherited = "Inherited"
	EncryptionAES256      = "AES256"
	StorageClassStandard  = "STANDARD"
)

// DetectContentType returns the MIME type based on the file extension of the
// object key. Returns empty string if the type cannot be determined (provider
// will use its own default).
func DetectContentType(objectKey string) string {
	ext := filepath.Ext(objectKey)
	if ext == "" {
		return ""
	}
	return mime.TypeByExtension(ext)
}
