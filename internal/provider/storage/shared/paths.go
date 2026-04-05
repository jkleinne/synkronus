package shared

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ObjectBasename extracts a safe filename from an object key.
// Returns an error for directory markers or keys that cannot produce a filename.
func ObjectBasename(objectKey string) (string, error) {
	if strings.HasSuffix(objectKey, "/") {
		return "", fmt.Errorf("cannot download directory marker object '%s'", objectKey)
	}
	base := filepath.Base(objectKey)
	if base == "." || base == "" {
		return "", fmt.Errorf("cannot derive filename from object key '%s'", objectKey)
	}
	return base, nil
}
