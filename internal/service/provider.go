// File: internal/service/provider.go
package service

import (
	"context"

	"synkronus/internal/domain/sql"
	"synkronus/internal/domain/storage"
)

// StorageProviderFactory is the interface the StorageService depends on to
// acquire a storage client for a named provider. Keeping this narrow allows
// the service to be tested with a simple mock rather than a real *factory.Factory.
type StorageProviderFactory interface {
	GetStorageProvider(ctx context.Context, name string) (storage.Storage, error)
}

// SqlProviderFactory is the interface the SqlService depends on to acquire a
// SQL client for a named provider.
type SqlProviderFactory interface {
	GetSqlProvider(ctx context.Context, name string) (sql.SQL, error)
}

// ProviderLister exposes provider enumeration used by CLI commands (e.g. the
// --providers resolver) without coupling them to the concrete factory type.
type ProviderLister interface {
	SupportedStorageProviders() []string
	ConfiguredStorageProviders() []string
	SupportedSqlProviders() []string
	ConfiguredSqlProviders() []string
	IsConfigured(providerName string) bool
	IsSqlConfigured(providerName string) bool
}
