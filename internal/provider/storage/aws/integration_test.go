//go:build integration

package aws

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"synkronus/internal/config"
	"synkronus/internal/domain/storage"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func newLocalStackStorage(t *testing.T) *AWSStorage {
	t.Helper()
	cfg := &config.AWSConfig{
		Region:   "us-east-1",
		Endpoint: "http://localhost:4566",
	}
	s, err := NewAWSStorage(context.Background(), cfg, slog.Default())
	if err != nil {
		t.Fatalf("failed to create LocalStack storage: %v", err)
	}
	return s
}

func uniqueBucketName(t *testing.T) string {
	t.Helper()
	name := strings.ToLower(t.Name())
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, "_", "-")
	// S3 bucket names must be 3-63 characters
	suffix := fmt.Sprintf("%d", time.Now().UnixNano()%1000000)
	full := fmt.Sprintf("t-%s-%s", name, suffix)
	if len(full) > 63 {
		full = full[:63]
	}
	return full
}

func TestIntegration_BucketLifecycle(t *testing.T) {
	s := newLocalStackStorage(t)
	ctx := context.Background()
	bucketName := uniqueBucketName(t)

	// Create
	err := s.CreateBucket(ctx, storage.CreateBucketOptions{
		Name:     bucketName,
		Location: "us-east-1",
	})
	if err != nil {
		t.Fatalf("CreateBucket failed: %v", err)
	}
	t.Cleanup(func() {
		_ = s.DeleteBucket(ctx, bucketName)
	})

	// List — verify present
	buckets, err := s.ListBuckets(ctx)
	if err != nil {
		t.Fatalf("ListBuckets failed: %v", err)
	}
	found := false
	for _, b := range buckets {
		if b.Name == bucketName {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected bucket %q in list, not found", bucketName)
	}

	// Describe
	bucket, err := s.DescribeBucket(ctx, bucketName)
	if err != nil {
		t.Fatalf("DescribeBucket failed: %v", err)
	}
	if bucket.Name != bucketName {
		t.Errorf("expected name %q, got %q", bucketName, bucket.Name)
	}

	// Delete
	err = s.DeleteBucket(ctx, bucketName)
	if err != nil {
		t.Fatalf("DeleteBucket failed: %v", err)
	}

	// List — verify absent
	buckets, err = s.ListBuckets(ctx)
	if err != nil {
		t.Fatalf("ListBuckets after delete failed: %v", err)
	}
	for _, b := range buckets {
		if b.Name == bucketName {
			t.Errorf("expected bucket %q to be deleted, still found", bucketName)
		}
	}
}

func TestIntegration_CreateBucketWithOptions(t *testing.T) {
	s := newLocalStackStorage(t)
	ctx := context.Background()
	bucketName := uniqueBucketName(t)

	versioning := true
	pap := storage.PublicAccessPreventionEnforced
	err := s.CreateBucket(ctx, storage.CreateBucketOptions{
		Name:                  bucketName,
		Location:              "us-east-1",
		Versioning:            &versioning,
		Labels:                map[string]string{"env": "test", "team": "integration"},
		PublicAccessPrevention: &pap,
	})
	if err != nil {
		t.Fatalf("CreateBucket with options failed: %v", err)
	}
	t.Cleanup(func() {
		_ = s.DeleteBucket(ctx, bucketName)
	})

	bucket, err := s.DescribeBucket(ctx, bucketName)
	if err != nil {
		t.Fatalf("DescribeBucket failed: %v", err)
	}

	if bucket.Versioning == nil || !bucket.Versioning.Enabled {
		t.Error("expected versioning to be enabled")
	}
	if len(bucket.Labels) != 2 {
		t.Errorf("expected 2 labels, got %d", len(bucket.Labels))
	}
	if bucket.Labels["env"] != "test" {
		t.Errorf("expected label env=test, got %q", bucket.Labels["env"])
	}
	if bucket.PublicAccessPrevention != "Enforced" {
		t.Errorf("expected public access 'Enforced', got %q", bucket.PublicAccessPrevention)
	}
}

func TestIntegration_ObjectOperations(t *testing.T) {
	s := newLocalStackStorage(t)
	ctx := context.Background()
	bucketName := uniqueBucketName(t)

	err := s.CreateBucket(ctx, storage.CreateBucketOptions{
		Name:     bucketName,
		Location: "us-east-1",
	})
	if err != nil {
		t.Fatalf("CreateBucket failed: %v", err)
	}
	t.Cleanup(func() {
		_, _ = s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: &bucketName,
			Key:    strPtr("test/hello.txt"),
		})
		_ = s.DeleteBucket(ctx, bucketName)
	})

	// Upload a test object directly via SDK
	content := "hello world"
	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &bucketName,
		Key:         strPtr("test/hello.txt"),
		Body:        strings.NewReader(content),
		ContentType: strPtr("text/plain"),
	})
	if err != nil {
		t.Fatalf("PutObject failed: %v", err)
	}

	// ListObjects — root level shows prefix
	objList, err := s.ListObjects(ctx, bucketName, "")
	if err != nil {
		t.Fatalf("ListObjects failed: %v", err)
	}
	if len(objList.CommonPrefixes) == 0 {
		t.Error("expected at least one common prefix (test/)")
	}

	// ListObjects with prefix — shows objects
	objList, err = s.ListObjects(ctx, bucketName, "test/")
	if err != nil {
		t.Fatalf("ListObjects with prefix failed: %v", err)
	}
	if len(objList.Objects) != 1 {
		t.Fatalf("expected 1 object, got %d", len(objList.Objects))
	}
	if objList.Objects[0].Key != "test/hello.txt" {
		t.Errorf("expected key 'test/hello.txt', got %q", objList.Objects[0].Key)
	}

	// DescribeObject
	obj, err := s.DescribeObject(ctx, bucketName, "test/hello.txt")
	if err != nil {
		t.Fatalf("DescribeObject failed: %v", err)
	}
	if obj.ContentType != "text/plain" {
		t.Errorf("expected content-type 'text/plain', got %q", obj.ContentType)
	}
	if obj.Size != int64(len(content)) {
		t.Errorf("expected size %d, got %d", len(content), obj.Size)
	}

	// DownloadObject
	reader, err := s.DownloadObject(ctx, bucketName, "test/hello.txt")
	if err != nil {
		t.Fatalf("DownloadObject failed: %v", err)
	}
	defer reader.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, reader); err != nil {
		t.Fatalf("reading download failed: %v", err)
	}
	if buf.String() != content {
		t.Errorf("expected content %q, got %q", content, buf.String())
	}
}

func TestIntegration_ListObjects_Empty(t *testing.T) {
	s := newLocalStackStorage(t)
	ctx := context.Background()
	bucketName := uniqueBucketName(t)

	err := s.CreateBucket(ctx, storage.CreateBucketOptions{
		Name:     bucketName,
		Location: "us-east-1",
	})
	if err != nil {
		t.Fatalf("CreateBucket failed: %v", err)
	}
	t.Cleanup(func() {
		_ = s.DeleteBucket(ctx, bucketName)
	})

	objList, err := s.ListObjects(ctx, bucketName, "")
	if err != nil {
		t.Fatalf("ListObjects failed: %v", err)
	}
	if len(objList.Objects) != 0 {
		t.Errorf("expected 0 objects, got %d", len(objList.Objects))
	}
}

func TestIntegration_UploadObject(t *testing.T) {
	s := newLocalStackStorage(t)
	ctx := context.Background()
	bucketName := uniqueBucketName(t)

	err := s.CreateBucket(ctx, storage.CreateBucketOptions{
		Name:     bucketName,
		Location: "us-east-1",
	})
	if err != nil {
		t.Fatalf("CreateBucket failed: %v", err)
	}
	t.Cleanup(func() {
		_, _ = s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: &bucketName,
			Key:    strPtr("upload-test.json"),
		})
		_ = s.DeleteBucket(ctx, bucketName)
	})

	content := `{"key": "value"}`
	opts := storage.UploadObjectOptions{
		BucketName:  bucketName,
		ObjectKey:   "upload-test.json",
		ContentType: "application/json",
		Metadata:    map[string]string{"env": "test"},
	}

	err = s.UploadObject(ctx, opts, strings.NewReader(content))
	if err != nil {
		t.Fatalf("UploadObject failed: %v", err)
	}

	obj, err := s.DescribeObject(ctx, bucketName, "upload-test.json")
	if err != nil {
		t.Fatalf("DescribeObject after upload failed: %v", err)
	}
	if obj.ContentType != "application/json" {
		t.Errorf("expected content-type 'application/json', got %q", obj.ContentType)
	}
	if obj.Size != int64(len(content)) {
		t.Errorf("expected size %d, got %d", len(content), obj.Size)
	}
	if obj.Metadata["env"] != "test" {
		t.Errorf("expected metadata env=test, got %q", obj.Metadata["env"])
	}

	reader, err := s.DownloadObject(ctx, bucketName, "upload-test.json")
	if err != nil {
		t.Fatalf("DownloadObject after upload failed: %v", err)
	}
	defer reader.Close()
	var buf bytes.Buffer
	io.Copy(&buf, reader)
	if buf.String() != content {
		t.Errorf("expected content %q, got %q", content, buf.String())
	}
}

func TestIntegration_DeleteObject(t *testing.T) {
	s := newLocalStackStorage(t)
	ctx := context.Background()
	bucketName := uniqueBucketName(t)

	err := s.CreateBucket(ctx, storage.CreateBucketOptions{
		Name:     bucketName,
		Location: "us-east-1",
	})
	if err != nil {
		t.Fatalf("CreateBucket failed: %v", err)
	}
	t.Cleanup(func() {
		_ = s.DeleteBucket(ctx, bucketName)
	})

	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &bucketName,
		Key:    strPtr("to-delete.txt"),
		Body:   strings.NewReader("delete me"),
	})
	if err != nil {
		t.Fatalf("PutObject failed: %v", err)
	}

	err = s.DeleteObject(ctx, bucketName, "to-delete.txt")
	if err != nil {
		t.Fatalf("DeleteObject failed: %v", err)
	}

	_, err = s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: &bucketName,
		Key:    strPtr("to-delete.txt"),
	})
	if err == nil {
		t.Error("expected HeadObject to fail after deletion")
	}

	err = s.DeleteObject(ctx, bucketName, "to-delete.txt")
	if err != nil {
		t.Errorf("expected idempotent delete, got error: %v", err)
	}
}

func TestIntegration_CopyObject(t *testing.T) {
	s := newLocalStackStorage(t)
	ctx := context.Background()
	srcBucket := uniqueBucketName(t)
	dstBucket := uniqueBucketName(t)

	for _, name := range []string{srcBucket, dstBucket} {
		err := s.CreateBucket(ctx, storage.CreateBucketOptions{
			Name:     name,
			Location: "us-east-1",
		})
		if err != nil {
			t.Fatalf("CreateBucket(%s) failed: %v", name, err)
		}
	}
	t.Cleanup(func() {
		_, _ = s.client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: &srcBucket, Key: strPtr("original.txt")})
		_, _ = s.client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: &dstBucket, Key: strPtr("copied.txt")})
		_ = s.DeleteBucket(ctx, srcBucket)
		_ = s.DeleteBucket(ctx, dstBucket)
	})

	content := "copy me"
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &srcBucket,
		Key:         strPtr("original.txt"),
		Body:        strings.NewReader(content),
		ContentType: strPtr("text/plain"),
	})
	if err != nil {
		t.Fatalf("PutObject failed: %v", err)
	}

	err = s.CopyObject(ctx, srcBucket, "original.txt", dstBucket, "copied.txt")
	if err != nil {
		t.Fatalf("CopyObject failed: %v", err)
	}

	obj, err := s.DescribeObject(ctx, dstBucket, "copied.txt")
	if err != nil {
		t.Fatalf("DescribeObject on copy target failed: %v", err)
	}
	if obj.Size != int64(len(content)) {
		t.Errorf("expected size %d, got %d", len(content), obj.Size)
	}

	reader, err := s.DownloadObject(ctx, dstBucket, "copied.txt")
	if err != nil {
		t.Fatalf("DownloadObject on copy target failed: %v", err)
	}
	defer reader.Close()
	var buf bytes.Buffer
	io.Copy(&buf, reader)
	if buf.String() != content {
		t.Errorf("expected content %q, got %q", content, buf.String())
	}
}
