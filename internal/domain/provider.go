package domain

// Provider identifies a cloud provider (e.g., GCP, AWS).
type Provider string

const (
	GCP Provider = "GCP"
	AWS Provider = "AWS"
)
