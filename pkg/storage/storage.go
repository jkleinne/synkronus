package storage

type Provider string

const (
	GCP Provider = "gcp"
	AWS Provider = "aws"
)

type StorageService interface {
	List() ([]string, error)
	DescribeBucket(bucketName string) (map[string]string, error)
}

type StorageClient struct {
	services map[Provider]StorageService
}

func NewStorageClient(services map[Provider]StorageService) *StorageClient {
	return &StorageClient{
		services: services,
	}
}

func (c *StorageClient) GetService(provider Provider) (StorageService, bool) {
	service, exists := c.services[provider]
	return service, exists
}

func (c *StorageClient) ListProviders() []Provider {
	var providers []Provider
	for provider := range c.services {
		providers = append(providers, provider)
	}
	return providers
}
