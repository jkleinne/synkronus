// File: pkg/storage/gcp/objects.go
package gcp

import (
	"context"
	"fmt"
	"synkronus/pkg/common"
	"synkronus/pkg/storage"

	gcpstorage "cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

func (g *GCPStorage) ListObjects(ctx context.Context, bucketName string, prefix string) (storage.ObjectList, error) {
	g.logger.Debug("Starting GCP ListObjects operation (delimited)", "bucket", bucketName, "prefix", prefix)

	bucketHandle := g.client.Bucket(bucketName)

	query := &gcpstorage.Query{
		Prefix:    prefix,
		Delimiter: "/",
	}

	it := bucketHandle.Objects(ctx, query)

	result := storage.ObjectList{
		BucketName:     bucketName,
		Prefix:         prefix,
		Objects:        []storage.Object{},
		CommonPrefixes: []string{},
	}

	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return storage.ObjectList{}, fmt.Errorf("error iterating objects: %w", err)
		}

		// If attrs.Prefix is set, it's a common prefix (directory)
		if attrs.Prefix != "" {
			result.CommonPrefixes = append(result.CommonPrefixes, attrs.Prefix)
			continue
		}

		// Otherwise, it's an object (file)
		obj := mapObjectAttributes(attrs, nil)
		result.Objects = append(result.Objects, obj)
	}

	return result, nil
}

func (g *GCPStorage) DescribeObject(ctx context.Context, bucketName string, objectKey string) (storage.Object, error) {
	g.logger.Debug("Starting GCP DescribeObject operation", "bucket", bucketName, "object", objectKey)

	objectHandle := g.client.Bucket(bucketName).Object(objectKey)

	// Fetch the object attributes (metadata)
	attrs, err := objectHandle.Attrs(ctx)
	if err != nil {
		return storage.Object{}, fmt.Errorf("error getting object attributes: %w", err)
	}

	// Determine encryption details
	var encryption *storage.Encryption
	if attrs.KMSKeyName != "" {
		// Customer-Managed Encryption Key (CMEK)
		encryption = &storage.Encryption{
			KmsKeyName: attrs.KMSKeyName,
			Algorithm:  "AES256",
		}
	} else if attrs.CustomerKeySHA256 != "" {
		// Customer-Supplied Encryption Key (CSEK)
		encryption = &storage.Encryption{
			KmsKeyName: "(Customer-Supplied Key)",
			Algorithm:  "AES256",
		}
	}
	// If both are empty, it's Google-managed encryption, handled by the mapper if encryption is nil

	obj := mapObjectAttributes(attrs, encryption)

	return obj, nil
}

// Maps GCP SDK object attributes to the domain model
func mapObjectAttributes(attrs *gcpstorage.ObjectAttrs, encryption *storage.Encryption) storage.Object {
	if attrs == nil {
		return storage.Object{}
	}

	// Handle Google-managed encryption if no specific encryption was provided
	if encryption == nil {
		encryption = &storage.Encryption{
			KmsKeyName: "Google-managed",
			Algorithm:  "AES256",
		}
	}

	return storage.Object{
		Key:                attrs.Name,
		Bucket:             attrs.Bucket,
		Provider:           common.GCP,
		Size:               attrs.Size,
		StorageClass:       attrs.StorageClass,
		LastModified:       attrs.Updated,
		CreatedAt:          attrs.Created,
		UpdatedAt:          attrs.Updated,
		ETag:               attrs.Etag,
		ContentType:        attrs.ContentType,
		ContentEncoding:    attrs.ContentEncoding,
		ContentLanguage:    attrs.ContentLanguage,
		CacheControl:       attrs.CacheControl,
		ContentDisposition: attrs.ContentDisposition,
		MD5Hash:            formatMD5(attrs.MD5),
		CRC32C:             formatCRC32C(attrs.CRC32C),
		Generation:         attrs.Generation,
		Metageneration:     attrs.Metageneration,
		Metadata:           attrs.Metadata,
		Encryption:         encryption,
	}
}
