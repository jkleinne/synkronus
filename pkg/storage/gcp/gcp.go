// Package gcp provides Google Cloud Platform storage operations implementation
package gcp

import (
	"context"
	"fmt"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

type GCPStorage struct {
	client     *storage.Client
	projectID  string
	bucketName string
}

func NewGCPStorage(projectID, bucketName string) (*GCPStorage, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCP storage client: %w", err)
	}

	return &GCPStorage{
		client:     client,
		projectID:  projectID,
		bucketName: bucketName,
	}, nil
}

// List returns a slice of bucket names in the project
func (g *GCPStorage) List() ([]string, error) {
	ctx := context.Background()
	var buckets []string

	it := g.client.Buckets(ctx, g.projectID)
	for {
		bucketAttrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error listing buckets: %w", err)
		}
		buckets = append(buckets, bucketAttrs.Name)
	}

	return buckets, nil
}

// DescribeBucket returns details about a specific bucket
func (g *GCPStorage) DescribeBucket(bucketName string) (map[string]interface{}, error) {
	ctx := context.Background()

	bucket := g.client.Bucket(bucketName)
	attrs, err := bucket.Attrs(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting bucket attributes: %w", err)
	}

	details := map[string]interface{}{
		"name":         attrs.Name,
		"location":     attrs.Location,
		"storageClass": attrs.StorageClass,
		"created":      attrs.Created,
		"updated":      attrs.Updated,
	}

	return details, nil
}

// Close closes the GCP storage client
func (g *GCPStorage) Close() error {
	if g.client != nil {
		return g.client.Close()
	}
	return nil
}
