// File: internal/provider/factory.go
package provider

import (
	"context"
	"fmt"
	"synkronus/internal/config"
	"synkronus/pkg/storage"
	"synkronus/pkg/storage/aws"
	"synkronus/pkg/storage/gcp"
)

type Factory struct {
	cfg *config.Config
}

func NewFactory(cfg *config.Config) *Factory {
	return &Factory{cfg: cfg}
}

func (f *Factory) GetStorageProvider(ctx context.Context, providerName string) (storage.Storage, error) {
	switch providerName {
	case "gcp":
		if f.cfg.GCP == nil || f.cfg.GCP.Project == "" {
			return nil, fmt.Errorf("GCP project not configured. Use 'synkronus config set gcp.project <project-id>'")
		}
		return gcp.NewGCPStorage(ctx, f.cfg.GCP.Project)
	case "aws":
		if f.cfg.AWS == nil || f.cfg.AWS.Region == "" {
			return nil, fmt.Errorf("AWS region not configured. Use 'synkronus config set aws.region <region>'")
		}
		return aws.NewAWSStorage(f.cfg.AWS.Region), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", providerName)
	}
}

func (f *Factory) GetStorageProviders(useGCP, useAWS bool) []string {
	var providersToQuery []string

	onlyGCP := useGCP && !useAWS
	onlyAWS := useAWS && !useGCP
	noFlags := !useGCP && !useAWS

	if onlyGCP {
		providersToQuery = append(providersToQuery, "gcp")
	} else if onlyAWS {
		providersToQuery = append(providersToQuery, "aws")
	} else {
		gcpConfigured := f.cfg.GCP != nil && f.cfg.GCP.Project != ""
		awsConfigured := f.cfg.AWS != nil && f.cfg.AWS.Region != ""

		if (gcpConfigured && noFlags) || useGCP {
			providersToQuery = append(providersToQuery, "gcp")
		}
		if (awsConfigured && noFlags) || useAWS {
			providersToQuery = append(providersToQuery, "aws")
		}
	}

	return providersToQuery
}
