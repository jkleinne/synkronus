package tui

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"synkronus/internal/config"
	"synkronus/internal/domain/sql"
	"synkronus/internal/domain/storage"
	"synkronus/internal/provider/factory"
	"synkronus/internal/service"
	"synkronus/internal/tui/ui"
)

const defaultTimeout = 30 * time.Second

// --- Message Types ---

// BucketsLoadedMsg is sent when the bucket list fetch completes.
type BucketsLoadedMsg struct {
	Buckets []storage.Bucket
	Err     error
}

// BucketDetailMsg is sent when a single bucket detail fetch completes.
type BucketDetailMsg struct {
	Bucket storage.Bucket
	Err    error
}

// ObjectsLoadedMsg is sent when an object list fetch completes.
type ObjectsLoadedMsg struct {
	Objects storage.ObjectList
	Err     error
}

// ObjectDetailMsg is sent when a single object detail fetch completes.
type ObjectDetailMsg struct {
	Object storage.Object
	Err    error
}

// InstancesLoadedMsg is sent when the SQL instance list fetch completes.
type InstancesLoadedMsg struct {
	Instances []sql.Instance
	Err       error
}

// InstanceDetailMsg is sent when a single SQL instance detail fetch completes.
type InstanceDetailMsg struct {
	Instance sql.Instance
	Err      error
}

// ConfigLoadedMsg is sent when the config key-value list has been read.
type ConfigLoadedMsg struct {
	Entries []ui.ConfigEntry
	Err     error
}

// BucketCreatedMsg is sent when a bucket creation operation completes.
type BucketCreatedMsg struct{ Err error }

// BucketDeletedMsg is sent when a bucket deletion operation completes.
type BucketDeletedMsg struct{ Err error }

// ObjectDownloadedMsg is sent when an object download operation completes.
type ObjectDownloadedMsg struct {
	FilePath string
	Err      error
}

// ConfigUpdatedMsg is sent when a config set operation completes.
type ConfigUpdatedMsg struct{ Err error }

// ConfigDeletedMsg is sent when a config delete operation completes.
type ConfigDeletedMsg struct{ Err error }
type ProviderRemovedMsg struct{ Err error }

// ObjectUploadedMsg is sent when an object upload operation completes.
type ObjectUploadedMsg struct{ Err error }

// ObjectDeletedMsg is sent when an object deletion operation completes.
type ObjectDeletedMsg struct{ Err error }

// StatusClearMsg is sent after a timer to clear transient status messages.
type StatusClearMsg struct{}

// --- Cmd Factories ---

// fetchBucketsCmd queries all configured storage providers for their buckets in parallel.
func fetchBucketsCmd(svc *service.StorageService, f *factory.Factory) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
		defer cancel()
		providers := f.GetConfiguredProviders()
		buckets, err := svc.ListAllBuckets(ctx, providers)
		return BucketsLoadedMsg{Buckets: buckets, Err: err}
	}
}

// fetchBucketDetailCmd fetches detailed information about a single bucket.
func fetchBucketDetailCmd(svc *service.StorageService, bucketName, provider string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
		defer cancel()
		bucket, err := svc.DescribeBucket(ctx, bucketName, provider)
		return BucketDetailMsg{Bucket: bucket, Err: err}
	}
}

// fetchObjectsCmd lists objects in a bucket under an optional key prefix.
func fetchObjectsCmd(svc *service.StorageService, bucketName, provider, prefix string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
		defer cancel()
		objects, err := svc.ListObjects(ctx, bucketName, provider, prefix, storage.DefaultMaxResults)
		return ObjectsLoadedMsg{Objects: objects, Err: err}
	}
}

// fetchObjectDetailCmd fetches detailed metadata for a single object.
func fetchObjectDetailCmd(svc *service.StorageService, bucketName, objectKey, provider string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
		defer cancel()
		object, err := svc.DescribeObject(ctx, bucketName, objectKey, provider)
		return ObjectDetailMsg{Object: object, Err: err}
	}
}

// fetchInstancesCmd queries all configured SQL providers for their instances in parallel.
func fetchInstancesCmd(svc *service.SqlService, f *factory.Factory) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
		defer cancel()
		providers := f.GetConfiguredSqlProviders()
		instances, err := svc.ListAllInstances(ctx, providers)
		return InstancesLoadedMsg{Instances: instances, Err: err}
	}
}

// fetchInstanceDetailCmd fetches detailed information about a single SQL instance.
func fetchInstanceDetailCmd(svc *service.SqlService, instanceName, provider string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
		defer cancel()
		instance, err := svc.DescribeInstance(ctx, instanceName, provider)
		return InstanceDetailMsg{Instance: instance, Err: err}
	}
}

// fetchConfigCmd reads all current config settings and flattens them into a key-value list.
func fetchConfigCmd(cm *config.ConfigManager) tea.Cmd {
	return func() tea.Msg {
		settings := cm.GetAllSettings()
		entries := flattenSettings(settings, "")
		return ConfigLoadedMsg{Entries: entries, Err: nil}
	}
}

// createBucketCmd submits a create-bucket request with the given options to the specified provider.
func createBucketCmd(svc *service.StorageService, opts storage.CreateBucketOptions, provider string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
		defer cancel()
		err := svc.CreateBucket(ctx, opts, provider)
		return BucketCreatedMsg{Err: err}
	}
}

// deleteBucketCmd deletes the named bucket from the specified provider.
func deleteBucketCmd(svc *service.StorageService, name, provider string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
		defer cancel()
		err := svc.DeleteBucket(ctx, name, provider)
		return BucketDeletedMsg{Err: err}
	}
}

// downloadObjectCmd downloads an object to the specified directory.
func downloadObjectCmd(svc *service.StorageService, bucketName, objectKey, provider, destDir string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		basename, err := objectBasename(objectKey)
		if err != nil {
			return ObjectDownloadedMsg{Err: err}
		}

		reader, err := svc.DownloadObject(ctx, bucketName, objectKey, provider)
		if err != nil {
			return ObjectDownloadedMsg{Err: fmt.Errorf("download failed: %w", err)}
		}
		defer reader.Close()

		expandedDir, expandErr := expandTilde(destDir)
		if expandErr != nil {
			return ObjectDownloadedMsg{Err: expandErr}
		}

		destPath := filepath.Join(expandedDir, basename)

		f, err := os.Create(destPath)
		if err != nil {
			return ObjectDownloadedMsg{Err: fmt.Errorf("error creating file '%s': %w", destPath, err)}
		}

		_, copyErr := io.Copy(f, reader)
		closeErr := f.Close()

		if copyErr != nil {
			os.Remove(destPath)
			return ObjectDownloadedMsg{Err: fmt.Errorf("error writing to '%s': %w", destPath, copyErr)}
		}
		if closeErr != nil {
			return ObjectDownloadedMsg{Err: fmt.Errorf("error closing '%s': %w", destPath, closeErr)}
		}

		return ObjectDownloadedMsg{FilePath: destPath}
	}
}

// uploadObjectCmd uploads a local file as a storage object.
// If objectKey is empty, the basename of filePath is used.
func uploadObjectCmd(svc *service.StorageService, bucketName, provider, filePath, objectKey string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		expandedPath, err := expandTilde(filePath)
		if err != nil {
			return ObjectUploadedMsg{Err: err}
		}

		info, err := os.Stat(expandedPath)
		if err != nil {
			return ObjectUploadedMsg{Err: fmt.Errorf("file not found: %s", expandedPath)}
		}
		if info.IsDir() {
			return ObjectUploadedMsg{Err: fmt.Errorf("path is a directory, not a file: %s", expandedPath)}
		}

		if objectKey == "" {
			objectKey = filepath.Base(expandedPath)
		}

		f, err := os.Open(expandedPath)
		if err != nil {
			return ObjectUploadedMsg{Err: fmt.Errorf("error opening file: %w", err)}
		}
		defer f.Close()

		opts := storage.UploadObjectOptions{
			BucketName: bucketName,
			ObjectKey:  objectKey,
		}

		err = svc.UploadObject(ctx, opts, provider, f)
		return ObjectUploadedMsg{Err: err}
	}
}

// deleteObjectCmd deletes an object from a bucket.
func deleteObjectCmd(svc *service.StorageService, bucketName, objectKey, provider string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
		defer cancel()
		err := svc.DeleteObject(ctx, bucketName, objectKey, provider)
		return ObjectDeletedMsg{Err: err}
	}
}

// objectBasename extracts a safe filename from an object key.
func objectBasename(objectKey string) (string, error) {
	if strings.HasSuffix(objectKey, "/") {
		return "", fmt.Errorf("cannot download directory marker object '%s'", objectKey)
	}
	base := filepath.Base(objectKey)
	if base == "." || base == "" {
		return "", fmt.Errorf("cannot derive filename from object key '%s'", objectKey)
	}
	return base, nil
}

// expandTilde replaces a leading ~ with the user's home directory.
func expandTilde(path string) (string, error) {
	if path == "~" {
		return os.UserHomeDir()
	}
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, path[2:]), nil
	}
	return path, nil
}

// setConfigCmd persists a single config key-value pair.
func setConfigCmd(cm *config.ConfigManager, key, value string) tea.Cmd {
	return func() tea.Msg {
		err := cm.SetValue(key, value)
		return ConfigUpdatedMsg{Err: err}
	}
}

// deleteConfigCmd removes a config key, returning an error if the key does not exist.
func deleteConfigCmd(cm *config.ConfigManager, key string) tea.Cmd {
	return func() tea.Msg {
		_, err := cm.DeleteValue(key)
		return ConfigDeletedMsg{Err: err}
	}
}

func removeProviderCmd(cm *config.ConfigManager, providerName string) tea.Cmd {
	return func() tea.Msg {
		_, err := cm.RemoveProvider(providerName)
		return ProviderRemovedMsg{Err: err}
	}
}

// clearStatusCmd returns a command that fires StatusClearMsg after 3 seconds,
// allowing transient status messages to be removed from the view.
func clearStatusCmd() tea.Cmd {
	return tea.Tick(3*time.Second, func(time.Time) tea.Msg {
		return StatusClearMsg{}
	})
}

// flattenSettings converts the nested config map into a flat key-value list.
// Nested keys are joined with "." (e.g., {"gcp": {"project": "x"}} → "gcp.project" = "x").
func flattenSettings(settings map[string]interface{}, prefix string) []ui.ConfigEntry {
	var entries []ui.ConfigEntry
	for key, val := range settings {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}
		switch v := val.(type) {
		case map[string]interface{}:
			entries = append(entries, flattenSettings(v, fullKey)...)
		default:
			entries = append(entries, ui.ConfigEntry{
				Key:   fullKey,
				Value: fmt.Sprintf("%v", v),
			})
		}
	}
	return entries
}
