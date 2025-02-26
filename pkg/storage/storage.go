package storage

type Provider string

const (
	GCP Provider = "gcp"
	AWS Provider = "aws"
)

type Storage interface {
	// List returns a slice of bucket names
	List() ([]string, error)

	// DescribeBucket returns details about a specific bucket
	DescribeBucket(bucketName string) (map[string]interface{}, error)

	// Close releases any resources used by the storage client
	Close() error
}
