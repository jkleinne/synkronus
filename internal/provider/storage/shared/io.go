package shared

import (
	"fmt"
	"io"
	"os"
)

// WriteToFile creates the destination file, copies the reader content into it,
// and removes the file if the copy fails to avoid leaving partial data on disk.
func WriteToFile(path string, src io.Reader) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("error creating file '%s': %w", path, err)
	}

	_, copyErr := io.Copy(f, src)
	closeErr := f.Close()

	if copyErr != nil {
		os.Remove(path)
		return fmt.Errorf("error writing to '%s': %w", path, copyErr)
	}
	if closeErr != nil {
		return fmt.Errorf("error closing '%s': %w", path, closeErr)
	}
	return nil
}
