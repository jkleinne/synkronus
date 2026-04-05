// File: internal/provider/storage/gcp/objects.go
package gcp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"path/filepath"
	"synkronus/internal/domain"
	"synkronus/internal/domain/storage"

	gcpstorage "cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

// detectContentType returns the MIME type based on the file extension of the object key.
// Returns empty string if the type cannot be determined (provider will use its own default).
func detectContentType(objectKey string) string {
	ext := filepath.Ext(objectKey)
	if ext == "" {
		return ""
	}
	return mime.TypeByExtension(ext)
}

func (g *GCPStorage) ListObjects(ctx context.Context, bucketName string, prefix string, maxResults int) (storage.ObjectList, error) {
	g.logger.Debug("Starting GCP ListObjects operation (delimited)", "bucket", bucketName, "prefix", prefix, "maxResults", maxResults)

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

	count := 0
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return storage.ObjectList{}, fmt.Errorf("error iterating objects: %w", err)
		}

		if maxResults > 0 && count >= maxResults {
			result.IsTruncated = true
			break
		}

		if attrs.Prefix != "" {
			result.CommonPrefixes = append(result.CommonPrefixes, attrs.Prefix)
		} else {
			obj := mapObjectAttributes(attrs, nil)
			result.Objects = append(result.Objects, obj)
		}
		count++
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
		Provider:           domain.GCP,
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

func (g *GCPStorage) DownloadObject(ctx context.Context, bucketName string, objectKey string) (io.ReadCloser, error) {
	g.logger.Debug("Starting GCP DownloadObject operation", "bucket", bucketName, "object", objectKey)

	reader, err := g.client.Bucket(bucketName).Object(objectKey).NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to open object reader: %w", err)
	}
	return reader, nil
}

func (g *GCPStorage) UploadObject(ctx context.Context, opts storage.UploadObjectOptions, reader io.Reader) error {
	g.logger.Debug("Starting GCP UploadObject operation", "bucket", opts.BucketName, "key", opts.ObjectKey)

	w := g.client.Bucket(opts.BucketName).Object(opts.ObjectKey).NewWriter(ctx)

	contentType := opts.ContentType
	if contentType == "" {
		contentType = detectContentType(opts.ObjectKey)
	}
	if contentType != "" {
		w.ContentType = contentType
	}
	w.Metadata = opts.Metadata

	if _, err := io.Copy(w, reader); err != nil {
		w.Close()
		return fmt.Errorf("uploading object %s to bucket %s: %w", opts.ObjectKey, opts.BucketName, err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("uploading object %s to bucket %s: %w", opts.ObjectKey, opts.BucketName, err)
	}

	return nil
}

func (g *GCPStorage) DeleteObject(ctx context.Context, bucketName, objectKey string) error {
	g.logger.Debug("Starting GCP DeleteObject operation", "bucket", bucketName, "key", objectKey)

	err := g.client.Bucket(bucketName).Object(objectKey).Delete(ctx)
	if err != nil {
		if errors.Is(err, gcpstorage.ErrObjectNotExist) {
			return nil
		}
		return fmt.Errorf("deleting object %s from bucket %s: %w", objectKey, bucketName, err)
	}
	return nil
}

func (g *GCPStorage) CopyObject(ctx context.Context, srcBucket, srcKey, destBucket, destKey string) error {
	g.logger.Debug("Starting GCP CopyObject operation",
		"srcBucket", srcBucket, "srcKey", srcKey,
		"destBucket", destBucket, "destKey", destKey)

	src := g.client.Bucket(srcBucket).Object(srcKey)
	dst := g.client.Bucket(destBucket).Object(destKey)

	if _, err := dst.CopierFrom(src).Run(ctx); err != nil {
		return fmt.Errorf("copying object %s/%s to %s/%s: %w", srcBucket, srcKey, destBucket, destKey, err)
	}
	return nil
}
