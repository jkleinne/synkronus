package aws

import (
	"context"
	"fmt"
	"io"
	"synkronus/internal/domain"
	"synkronus/internal/domain/storage"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func (s *AWSStorage) ListObjects(ctx context.Context, bucketName string, prefix string) (storage.ObjectList, error) {
	s.logger.Debug("Starting AWS ListObjects operation", "bucket", bucketName, "prefix", prefix)

	delimiter := "/"
	input := &s3.ListObjectsV2Input{
		Bucket:    &bucketName,
		Prefix:    &prefix,
		Delimiter: &delimiter,
	}

	output, err := s.client.ListObjectsV2(ctx, input)
	if err != nil {
		return storage.ObjectList{}, fmt.Errorf("failed to list S3 objects: %w", err)
	}

	result := storage.ObjectList{
		BucketName:     bucketName,
		Prefix:         prefix,
		Objects:        make([]storage.Object, 0, len(output.Contents)),
		CommonPrefixes: make([]string, 0, len(output.CommonPrefixes)),
	}

	for _, cp := range output.CommonPrefixes {
		result.CommonPrefixes = append(result.CommonPrefixes, derefString(cp.Prefix))
	}

	for _, obj := range output.Contents {
		o := storage.Object{
			Key:          derefString(obj.Key),
			Bucket:       bucketName,
			Provider:     domain.AWS,
			Size:         derefInt64(obj.Size),
			StorageClass: string(obj.StorageClass),
			ETag:         derefString(obj.ETag),
		}
		if obj.LastModified != nil {
			o.LastModified = *obj.LastModified
		}
		result.Objects = append(result.Objects, o)
	}

	return result, nil
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
		StorageClass:       string(out.StorageClass),
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

func derefInt64(p *int64) int64 {
	if p == nil {
		return 0
	}
	return *p
}
