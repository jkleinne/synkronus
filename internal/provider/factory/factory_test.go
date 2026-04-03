package factory

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"synkronus/internal/config"
	"synkronus/internal/domain"
	"synkronus/internal/domain/sql"
	"synkronus/internal/domain/storage"
	"synkronus/internal/provider/registry"
	"testing"
)

// --- Fake implementations ---

type fakeStorage struct{ name string }

func (f *fakeStorage) ListBuckets(ctx context.Context) ([]storage.Bucket, error) { return nil, nil }
func (f *fakeStorage) DescribeBucket(ctx context.Context, n string) (storage.Bucket, error) {
	return storage.Bucket{}, nil
}
func (f *fakeStorage) CreateBucket(ctx context.Context, opts storage.CreateBucketOptions) error {
	return nil
}
func (f *fakeStorage) DeleteBucket(ctx context.Context, n string) error                 { return nil }
func (f *fakeStorage) ListObjects(ctx context.Context, b, p string) (storage.ObjectList, error) {
	return storage.ObjectList{}, nil
}
func (f *fakeStorage) DescribeObject(ctx context.Context, b, k string) (storage.Object, error) {
	return storage.Object{}, nil
}
func (f *fakeStorage) ProviderName() domain.Provider { return domain.Provider(f.name) }
func (f *fakeStorage) Close() error                  { return nil }

type fakeSql struct{ name string }

func (f *fakeSql) ListInstances(ctx context.Context) ([]sql.Instance, error) { return nil, nil }
func (f *fakeSql) DescribeInstance(ctx context.Context, n string) (sql.Instance, error) {
	return sql.Instance{}, nil
}
func (f *fakeSql) ProviderName() domain.Provider { return domain.Provider(f.name) }
func (f *fakeSql) Close() error                  { return nil }

// --- Helpers ---

func gcpConfigured() *config.Config {
	return &config.Config{GCP: &config.GCPConfig{Project: "test-project"}}
}

func emptyConfig() *config.Config {
	return &config.Config{}
}

func testLogger() *slog.Logger {
	return slog.Default()
}

// registerTestStorage registers a storage provider with a unique test-prefixed name.
// Returns the normalized name used for lookups.
func registerTestStorage(t *testing.T, name string, configured bool, initErr error) string {
	t.Helper()
	reg := registry.Registration[storage.Storage]{
		ConfigCheck: func(cfg *config.Config) bool { return configured },
		Initializer: func(ctx context.Context, cfg *config.Config, logger *slog.Logger) (storage.Storage, error) {
			if initErr != nil {
				return nil, initErr
			}
			return &fakeStorage{name: name}, nil
		},
	}
	registry.RegisterProvider(name, reg)
	return strings.ToLower(name)
}

func registerTestSql(t *testing.T, name string, configured bool, initErr error) string {
	t.Helper()
	reg := registry.Registration[sql.SQL]{
		ConfigCheck: func(cfg *config.Config) bool { return configured },
		Initializer: func(ctx context.Context, cfg *config.Config, logger *slog.Logger) (sql.SQL, error) {
			if initErr != nil {
				return nil, initErr
			}
			return &fakeSql{name: name}, nil
		},
	}
	registry.RegisterSqlProvider(name, reg)
	return strings.ToLower(name)
}

// --- Tests ---

func TestGetConfiguredProviders_FiltersUnconfigured(t *testing.T) {
	registerTestStorage(t, "filtcfg-alpha", true, nil)
	registerTestStorage(t, "filtcfg-beta", false, nil)
	registerTestStorage(t, "filtcfg-gamma", true, nil)

	f := NewFactory(emptyConfig(), testLogger())
	result := f.GetConfiguredProviders()

	found := map[string]bool{}
	for _, name := range result {
		found[name] = true
	}
	if !found["filtcfg-alpha"] || !found["filtcfg-gamma"] {
		t.Errorf("expected filtcfg-alpha and filtcfg-gamma in results, got %v", result)
	}
	if found["filtcfg-beta"] {
		t.Errorf("filtcfg-beta should not be in results (unconfigured), got %v", result)
	}
}

func TestGetConfiguredProviders_ReturnsSorted(t *testing.T) {
	registerTestStorage(t, "sorttest-zebra", true, nil)
	registerTestStorage(t, "sorttest-apple", true, nil)

	f := NewFactory(emptyConfig(), testLogger())
	result := f.GetConfiguredProviders()

	// Find our test entries
	var testResults []string
	for _, name := range result {
		if strings.HasPrefix(name, "sorttest-") {
			testResults = append(testResults, name)
		}
	}
	if len(testResults) < 2 {
		t.Fatalf("expected at least 2 sorttest entries, got %v", testResults)
	}
	for i := 1; i < len(testResults); i++ {
		if testResults[i-1] > testResults[i] {
			t.Errorf("expected sorted order, got %v", testResults)
		}
	}
}

func TestGetConfiguredProviders_EmptyRegistry(t *testing.T) {
	// Test the generic helper directly with an empty map
	result := getConfigured(map[string]registry.Registration[storage.Storage]{}, emptyConfig())
	if len(result) != 0 {
		t.Errorf("expected empty result, got %v", result)
	}
}

func TestGetConfiguredSqlProviders_FiltersUnconfigured(t *testing.T) {
	registerTestSql(t, "sqlfilt-alpha", true, nil)
	registerTestSql(t, "sqlfilt-beta", false, nil)

	f := NewFactory(emptyConfig(), testLogger())
	result := f.GetConfiguredSqlProviders()

	found := map[string]bool{}
	for _, name := range result {
		found[name] = true
	}
	if !found["sqlfilt-alpha"] {
		t.Errorf("expected sqlfilt-alpha in results, got %v", result)
	}
	if found["sqlfilt-beta"] {
		t.Errorf("sqlfilt-beta should not be in results, got %v", result)
	}
}

func TestIsConfigured_RegisteredAndConfigured(t *testing.T) {
	registerTestStorage(t, "iscfg-yes", true, nil)
	f := NewFactory(emptyConfig(), testLogger())
	if !f.IsConfigured("iscfg-yes") {
		t.Error("expected IsConfigured=true for registered and configured provider")
	}
}

func TestIsConfigured_RegisteredButUnconfigured(t *testing.T) {
	registerTestStorage(t, "iscfg-no", false, nil)
	f := NewFactory(emptyConfig(), testLogger())
	if f.IsConfigured("iscfg-no") {
		t.Error("expected IsConfigured=false for unconfigured provider")
	}
}

func TestIsConfigured_NotRegistered(t *testing.T) {
	f := NewFactory(emptyConfig(), testLogger())
	if f.IsConfigured("nonexistent-provider") {
		t.Error("expected IsConfigured=false for unregistered provider")
	}
}

func TestIsSqlConfigured(t *testing.T) {
	registerTestSql(t, "sqlcfg-yes", true, nil)
	registerTestSql(t, "sqlcfg-no", false, nil)

	f := NewFactory(emptyConfig(), testLogger())

	tests := []struct {
		name     string
		provider string
		expected bool
	}{
		{"RegisteredAndConfigured", "sqlcfg-yes", true},
		{"RegisteredButUnconfigured", "sqlcfg-no", false},
		{"NotRegistered", "sqlcfg-missing", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if f.IsSqlConfigured(tt.provider) != tt.expected {
				t.Errorf("IsSqlConfigured(%q) = %v, want %v", tt.provider, !tt.expected, tt.expected)
			}
		})
	}
}

func TestGetStorageProvider_UnsupportedProvider(t *testing.T) {
	f := NewFactory(emptyConfig(), testLogger())
	_, err := f.GetStorageProvider(context.Background(), "totally-fake-provider")
	if err == nil {
		t.Fatal("expected error for unsupported provider")
	}
	if !strings.Contains(err.Error(), "unsupported storage provider") {
		t.Errorf("expected 'unsupported storage provider' in error, got: %v", err)
	}
}

func TestGetStorageProvider_UnconfiguredProvider(t *testing.T) {
	registerTestStorage(t, "getstg-uncfg", false, nil)
	f := NewFactory(emptyConfig(), testLogger())
	_, err := f.GetStorageProvider(context.Background(), "getstg-uncfg")
	if err == nil {
		t.Fatal("expected error for unconfigured provider")
	}
	if !strings.Contains(err.Error(), "not configured") {
		t.Errorf("expected 'not configured' in error, got: %v", err)
	}
}

func TestGetStorageProvider_InitializerFails(t *testing.T) {
	initErr := fmt.Errorf("auth failed")
	registerTestStorage(t, "getstg-fail", true, initErr)
	f := NewFactory(emptyConfig(), testLogger())
	_, err := f.GetStorageProvider(context.Background(), "getstg-fail")
	if err == nil {
		t.Fatal("expected error when initializer fails")
	}
	if !strings.Contains(err.Error(), "auth failed") {
		t.Errorf("expected wrapped error with 'auth failed', got: %v", err)
	}
}

func TestGetStorageProvider_Success(t *testing.T) {
	registerTestStorage(t, "getstg-ok", true, nil)
	f := NewFactory(emptyConfig(), testLogger())
	client, err := f.GetStorageProvider(context.Background(), "getstg-ok")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestGetStorageProvider_CaseInsensitive(t *testing.T) {
	registerTestStorage(t, "getstg-case", true, nil)
	f := NewFactory(emptyConfig(), testLogger())

	for _, name := range []string{"getstg-case", "GETSTG-CASE", "GetStg-Case"} {
		client, err := f.GetStorageProvider(context.Background(), name)
		if err != nil {
			t.Errorf("GetStorageProvider(%q): unexpected error: %v", name, err)
		}
		if client == nil {
			t.Errorf("GetStorageProvider(%q): expected non-nil client", name)
		}
	}
}

func TestUpdateConfig_SubsequentCallsUseNewConfig(t *testing.T) {
	// Register provider that checks GCP config
	registry.RegisterProvider("updatecfg-test", registry.Registration[storage.Storage]{
		ConfigCheck: func(cfg *config.Config) bool {
			return cfg.GCP != nil && cfg.GCP.Project != ""
		},
		Initializer: func(ctx context.Context, cfg *config.Config, logger *slog.Logger) (storage.Storage, error) {
			return &fakeStorage{name: "updatecfg-test"}, nil
		},
	})

	f := NewFactory(emptyConfig(), testLogger())

	// Initially not configured
	if f.IsConfigured("updatecfg-test") {
		t.Error("expected unconfigured with empty config")
	}

	// Update config to include GCP
	f.UpdateConfig(gcpConfigured())

	// Now should be configured
	if !f.IsConfigured("updatecfg-test") {
		t.Error("expected configured after UpdateConfig")
	}
}

func TestGetSqlProvider_UnsupportedProvider(t *testing.T) {
	f := NewFactory(emptyConfig(), testLogger())
	_, err := f.GetSqlProvider(context.Background(), "totally-fake-sql")
	if err == nil {
		t.Fatal("expected error for unsupported SQL provider")
	}
	if !strings.Contains(err.Error(), "unsupported SQL provider") {
		t.Errorf("expected 'unsupported SQL provider' in error, got: %v", err)
	}
}

func TestGetSqlProvider_Success(t *testing.T) {
	registerTestSql(t, "getsql-ok", true, nil)
	f := NewFactory(emptyConfig(), testLogger())
	client, err := f.GetSqlProvider(context.Background(), "getsql-ok")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil SQL client")
	}
}
