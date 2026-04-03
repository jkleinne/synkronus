package tui

import (
	"context"
	"fmt"
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

// ConfigUpdatedMsg is sent when a config set operation completes.
type ConfigUpdatedMsg struct{ Err error }

// ConfigDeletedMsg is sent when a config delete operation completes.
type ConfigDeletedMsg struct{ Err error }
type ProviderRemovedMsg struct{ Err error }

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
		objects, err := svc.ListObjects(ctx, bucketName, provider, prefix)
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
