package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"synkronus/internal/domain"
	"synkronus/internal/domain/sql"
)

// --- Mock types ---

// mockSqlFactory satisfies SqlProviderFactory. If err is set it is returned for
// every GetSqlProvider call; otherwise the call is dispatched by provider name.
type mockSqlFactory struct {
	providers map[string]sql.SQL
	err       error
}

func (m *mockSqlFactory) GetSqlProvider(_ context.Context, name string) (sql.SQL, error) {
	if m.err != nil {
		return nil, m.err
	}
	client, ok := m.providers[name]
	if !ok {
		return nil, fmt.Errorf("unsupported provider: %s", name)
	}
	return client, nil
}

// mockSQL is a controllable implementation of sql.SQL.
type mockSQL struct {
	providerName domain.Provider
	instances    []sql.Instance
	instance     sql.Instance
	err          error
	closeCalled  bool
}

func (m *mockSQL) ListInstances(_ context.Context) ([]sql.Instance, error) {
	return m.instances, m.err
}

func (m *mockSQL) DescribeInstance(_ context.Context, _ string) (sql.Instance, error) {
	return m.instance, m.err
}

func (m *mockSQL) ProviderName() domain.Provider {
	return m.providerName
}

func (m *mockSQL) Close() error {
	m.closeCalled = true
	return nil
}

// newSqlService builds a SqlService wired to the supplied factory.
func newSqlService(factory SqlProviderFactory) *SqlService {
	return NewSqlService(factory, newTestLogger())
}

// --- SqlService tests ---

func TestSqlService_DescribeInstance_HappyPath(t *testing.T) {
	want := sql.Instance{Name: "prod-db", Provider: domain.GCP}
	mock := &mockSQL{instance: want}
	svc := newSqlService(&mockSqlFactory{providers: map[string]sql.SQL{"gcp": mock}})

	got, err := svc.DescribeInstance(context.Background(), "prod-db", "gcp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != want.Name {
		t.Errorf("got instance name %q, want %q", got.Name, want.Name)
	}
}

func TestSqlService_DescribeInstance_ProviderError(t *testing.T) {
	providerErr := errors.New("instance not found")
	mock := &mockSQL{err: providerErr}
	svc := newSqlService(&mockSqlFactory{providers: map[string]sql.SQL{"gcp": mock}})

	_, err := svc.DescribeInstance(context.Background(), "missing-db", "gcp")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, providerErr) {
		t.Errorf("error chain should wrap provider error; got: %v", err)
	}
	if !strings.Contains(err.Error(), "missing-db") {
		t.Errorf("error should contain instance name for context; got: %v", err)
	}
}

func TestSqlService_ListAllInstances_Empty(t *testing.T) {
	svc := newSqlService(&mockSqlFactory{providers: map[string]sql.SQL{}})

	results, err := svc.ListAllInstances(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error for empty provider list: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected empty result, got %d instances", len(results))
	}
}
