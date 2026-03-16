// File: internal/provider/registry/registry.go
package registry

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"sync"
	"synkronus/internal/config"
	"synkronus/pkg/sql"
	"synkronus/pkg/storage"
)

// ProviderConfigCheck defines the function signature for checking if a provider is configured
type ProviderConfigCheck func(cfg *config.Config) bool

// Registration holds the necessary functions to check configuration and initialize a provider
type Registration[T any] struct {
	ConfigCheck ProviderConfigCheck
	Initializer func(ctx context.Context, cfg *config.Config, logger *slog.Logger) (T, error)
}

// Registry is a thread-safe, generic provider registry
type Registry[T any] struct {
	mu       sync.RWMutex
	entries  map[string]Registration[T]
	typeName string // e.g. "storage" or "sql", used in panic messages
}

// NewRegistry creates a new generic Registry
func NewRegistry[T any](typeName string) *Registry[T] {
	return &Registry[T]{
		entries:  make(map[string]Registration[T]),
		typeName: typeName,
	}
}

func (r *Registry[T]) Register(name string, reg Registration[T]) {
	r.mu.Lock()
	defer r.mu.Unlock()

	normalizedName := strings.ToLower(name)
	if _, exists := r.entries[normalizedName]; exists {
		panic(fmt.Sprintf("%s provider %s already registered", r.typeName, normalizedName))
	}
	if reg.ConfigCheck == nil {
		panic(fmt.Sprintf("%s provider %s registration missing ConfigCheck", r.typeName, normalizedName))
	}
	if reg.Initializer == nil {
		panic(fmt.Sprintf("%s provider %s registration missing Initializer", r.typeName, normalizedName))
	}
	r.entries[normalizedName] = reg
}

func (r *Registry[T]) GetSupported() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.entries))
	for name := range r.entries {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (r *Registry[T]) IsSupported(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.entries[strings.ToLower(name)]
	return exists
}

func (r *Registry[T]) Get(name string) (Registration[T], bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	reg, exists := r.entries[strings.ToLower(name)]
	return reg, exists
}

func (r *Registry[T]) GetAll() map[string]Registration[T] {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]Registration[T], len(r.entries))
	for k, v := range r.entries {
		result[k] = v
	}
	return result
}

// Typed registry instances
var (
	storageRegistry = NewRegistry[storage.Storage]("storage")
	sqlRegistry     = NewRegistry[sql.SQL]("sql")
)

// --- Storage convenience functions ---

func RegisterProvider(name string, reg Registration[storage.Storage]) {
	storageRegistry.Register(name, reg)
}

func GetSupportedProviders() []string {
	return storageRegistry.GetSupported()
}

func IsSupported(name string) bool {
	return storageRegistry.IsSupported(name)
}

func GetRegistration(name string) (Registration[storage.Storage], bool) {
	return storageRegistry.Get(name)
}

func GetAllRegistrations() map[string]Registration[storage.Storage] {
	return storageRegistry.GetAll()
}

// --- SQL convenience functions ---

func RegisterSqlProvider(name string, reg Registration[sql.SQL]) {
	sqlRegistry.Register(name, reg)
}

func GetSupportedSqlProviders() []string {
	return sqlRegistry.GetSupported()
}

func IsSqlSupported(name string) bool {
	return sqlRegistry.IsSupported(name)
}

func GetSqlRegistration(name string) (Registration[sql.SQL], bool) {
	return sqlRegistry.Get(name)
}

func GetAllSqlRegistrations() map[string]Registration[sql.SQL] {
	return sqlRegistry.GetAll()
}
