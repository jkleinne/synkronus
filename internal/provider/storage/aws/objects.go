package aws

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/url"
	"path/filepath"
	"synkronus/internal/domain"
	"synkronus/internal/domain/storage"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func (s *AWSStorage) ListObjects(ctx context.Context, bucketName string, prefix string, maxResults int) (storage.ObjectList, error) {
	s.logger.Debug("Starting AWS ListObjects operation", "bucket", bucketName, "prefix", prefix, "maxResults", maxResults)

	delimiter := "/"
	result := storage.ObjectList{
		BucketName:     bucketName,
		Prefix:         prefix,
		Objects:        []storage.Object{},
		CommonPrefixes: []string{},
	}

	input := &s3.ListObjectsV2Input{
		Bucket:    &bucketName,
		Prefix:    &prefix,
		Delimiter: &delimiter,
	}

	if maxResults > 0 {
		maxKeys := int32(maxResults)
		input.MaxKeys = &maxKeys

		page, err := s.client.ListObjectsV2(ctx, input)
		if err != nil {
			return storage.ObjectList{}, fmt.Errorf("failed to list S3 objects: %w", err)
		}

		for _, cp := range page.CommonPrefixes {
			result.CommonPrefixes = append(result.CommonPrefixes, derefString(cp.Prefix))
		}
		for _, obj := range page.Contents {
			result.Objects = append(result.Objects, mapListObject(obj, bucketName))
		}
		if page.IsTruncated != nil && *page.IsTruncated {
			result.IsTruncated = true
		}

		return result, nil
	}

	// No limit — paginate through all results
	paginator := s3.NewListObjectsV2Paginator(s.client, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return storage.ObjectList{}, fmt.Errorf("failed to list S3 objects: %w", err)
		}

		for _, cp := range page.CommonPrefixes {
			result.CommonPrefixes = append(result.CommonPrefixes, derefString(cp.Prefix))
		}
		for _, obj := range page.Contents {
			result.Objects = append(result.Objects, mapListObject(obj, bucketName))
		}
	}

	return result, nil
}

// mapListObject maps an S3 ObjectIdentifier from a list response to the domain model.
func mapListObject(obj types.Object, bucketName string) storage.Object {
	o := storage.Object{
		Key:          derefString(obj.Key),
		Bucket:       bucketName,
		Provider:     domain.AWS,
		Size:         derefInt64(obj.Size),
		StorageClass: storageClassOrDefault(string(obj.StorageClass)),
		ETag:         derefString(obj.ETag),
	}
	if obj.LastModified != nil {
		o.LastModified = *obj.LastModified
	}
	return o
}

func (s *AWSStorage) DescribeObject(ctx context.Context, bucketName string, objectKey string) (storage.Object, error) {
	s.logger.Debug("Starting AWS DescribeObject operation", "bucket", bucketName, "object", objectKey)

	out, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: &bucketName,
		Key:    &objectKey,
	})
	if err != nil {
		return storage.Object{}, fmt.Errorf("failed to describe S3 object: %w", err)
	}

	obj := storage.Object{
		Key:                objectKey,
		Bucket:             bucketName,
		Provider:           domain.AWS,
		Size:               derefInt64(out.ContentLength),
		StorageClass:       storageClassOrDefault(string(out.StorageClass)),
		ETag:               derefString(out.ETag),
		ContentType:        derefString(out.ContentType),
		ContentEncoding:    derefString(out.ContentEncoding),
		ContentLanguage:    derefString(out.ContentLanguage),
		CacheControl:       derefString(out.CacheControl),
		ContentDisposition: derefString(out.ContentDisposition),
		VersionID:          derefString(out.VersionId),
		Metadata:           out.Metadata,
	}

	if out.LastModified != nil {
		obj.LastModified = *out.LastModified
	}

	// Map encryption
	if out.ServerSideEncryption != "" {
		obj.Encryption = &storage.Encryption{
			Algorithm: string(out.ServerSideEncryption),
		}
		if kmsKey := derefString(out.SSEKMSKeyId); kmsKey != "" {
			obj.Encryption.KmsKeyName = kmsKey
		}
	}

	return obj, nil
}

func (s *AWSStorage) DownloadObject(ctx context.Context, bucketName string, objectKey string) (io.ReadCloser, error) {
	s.logger.Debug("Starting AWS DownloadObject operation", "bucket", bucketName, "object", objectKey)

	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucketName,
		Key:    &objectKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download S3 object: %w", err)
	}

	return out.Body, nil
}

func (s *AWSStorage) UploadObject(ctx context.Context, opts storage.UploadObjectOptions, reader io.Reader) error {
	s.logger.Debug("Starting AWS UploadObject operation", "bucket", opts.BucketName, "key", opts.ObjectKey)

	input := &s3.PutObjectInput{
		Bucket: &opts.BucketName,
		Key:    &opts.ObjectKey,
		Body:   reader,
	}

	contentType := opts.ContentType
	if contentType == "" {
		contentType = detectContentType(opts.ObjectKey)
	}
	if contentType != "" {
		input.ContentType = &contentType
	}
	if len(opts.Metadata) > 0 {
		input.Metadata = opts.Metadata
	}

	if _, err := s.client.PutObject(ctx, input); err != nil {
		return fmt.Errorf("uploading object %s to bucket %s: %w", opts.ObjectKey, opts.BucketName, err)
	}
	return nil
}

func (s *AWSStorage) DeleteObject(ctx context.Context, bucketName, objectKey string) error {
	s.logger.Debug("Starting AWS DeleteObject operation", "bucket", bucketName, "key", objectKey)

	if _, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &bucketName,
		Key:    &objectKey,
	}); err != nil {
		return fmt.Errorf("deleting object %s from bucket %s: %w", objectKey, bucketName, err)
	}
	return nil
}

func (s *AWSStorage) CopyObject(ctx context.Context, srcBucket, srcKey, destBucket, destKey string) error {
	s.logger.Debug("Starting AWS CopyObject operation",
		"srcBucket", srcBucket, "srcKey", srcKey,
		"destBucket", destBucket, "destKey", destKey)

	copySource := srcBucket + "/" + url.PathEscape(srcKey)

	if _, err := s.client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     &destBucket,
		Key:        &destKey,
		CopySource: &copySource,
	}); err != nil {
		return fmt.Errorf("copying object %s/%s to %s/%s: %w", srcBucket, srcKey, destBucket, destKey, err)
	}
	return nil
}

// storageClassOrDefault returns "STANDARD" when S3 omits the storage class
// (which it does for STANDARD-class objects).
func storageClassOrDefault(sc string) string {
	if sc == "" {
		return "STANDARD"
	}
	return sc
}

func detectContentType(objectKey string) string {
	ext := filepath.Ext(objectKey)
	if ext == "" {
		return ""
	}
	return mime.TypeByExtension(ext)
}

func derefInt64(p *int64) int64 {
	if p == nil {
		return 0
	}
	return *p
}
