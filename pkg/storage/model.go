package storage

import (
	"fmt"
	"time"
)

type Provider string

const (
	GCP Provider = "GCP"
	AWS Provider = "AWS"
)

type Bucket struct {
	Name         string
	Provider     Provider
	Location     string
	StorageClass string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	// A value of -1 indicates that the usage is unknown or could not be retrieved
	UsageBytes int64
	Labels     map[string]string
}

func FormatBytes(bytes int64) string {
	if bytes < 0 {
		return "N/A"
	}
	if bytes == 0 {
		return "0 B"
	}

	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	sizes := []string{"KB", "MB", "GB", "TB", "PB", "EB"}
	if exp >= len(sizes) {
		return fmt.Sprintf("%d B", bytes) // Fallback if extremely large
	}
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), sizes[exp])
}
