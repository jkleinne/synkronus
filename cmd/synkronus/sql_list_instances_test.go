package main

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"synkronus/internal/config"
	"synkronus/internal/domain"
	"synkronus/internal/domain/sql"
	"synkronus/internal/output"
	"synkronus/internal/provider/factory"
	"synkronus/internal/service"
)

// --- SQL CLI-layer mock types ---

// cmdSqlFactory is a test double for service.SqlProviderFactory.
type cmdSqlFactory struct {
	providers map[string]sql.SQL
	err       error
}

func (f *cmdSqlFactory) GetSqlProvider(_ context.Context, name string) (sql.SQL, error) {
	if f.err != nil {
		return nil, f.err
	}
	client, ok := f.providers[name]
	if !ok {
		return nil, errors.New("unsupported SQL provider: " + name)
	}
	return client, nil
}

// cmdMockSQL is a controllable implementation of sql.SQL for CLI tests.
type cmdMockSQL struct {
	instances   []sql.Instance
	instance    sql.Instance
	err         error
	closeCalled bool
}

func (m *cmdMockSQL) ListInstances(_ context.Context) ([]sql.Instance, error) {
	return m.instances, m.err
}
func (m *cmdMockSQL) DescribeInstance(_ context.Context, _ string) (sql.Instance, error) {
	return m.instance, m.err
}
func (m *cmdMockSQL) ProviderName() domain.Provider { return domain.GCP }
func (m *cmdMockSQL) Close() error {
	m.closeCalled = true
	return nil
}

// newSqlTestApp builds a minimal appContainer for single-resource SQL commands
// (describe). ProviderFactory is nil because describe does not enumerate providers.
func newSqlTestApp(factory service.SqlProviderFactory) *appContainer {
	logger := cmdTestLogger()
	return &appContainer{
		SqlService:   service.NewSqlService(factory, logger),
		OutputFormat: output.FormatJSON,
		Logger:       logger,
	}
}

// newSqlListTestApp builds an appContainer for list-instances tests.
// It wires a real factory (for provider resolution) with a mock SqlService.
// The config marks GCP as configured so the provider resolver returns ["gcp"].
func newSqlListTestApp(sqlFactory service.SqlProviderFactory) *appContainer {
	logger := cmdTestLogger()
	gcpCfg := &config.Config{GCP: &config.GCPConfig{Project: "test-project"}}
	providerFactory := factory.NewFactory(gcpCfg, logger)
	return &appContainer{
		ProviderFactory: providerFactory,
		SqlService:      service.NewSqlService(sqlFactory, logger),
		OutputFormat:    output.FormatJSON,
		Logger:          logger,
	}
}

// --- list-instances tests ---

func TestListInstancesCmd_HappyPath_ReturnsInstances(t *testing.T) {
	wantInstances := []sql.Instance{
		{Name: "prod-db", Provider: domain.GCP, State: "RUNNABLE"},
		{Name: "staging-db", Provider: domain.GCP, State: "RUNNABLE"},
	}
	mock := &cmdMockSQL{instances: wantInstances}
	sqlFactory := &cmdSqlFactory{providers: map[string]sql.SQL{"gcp": mock}}
	app := newSqlListTestApp(sqlFactory)

	var buf bytes.Buffer
	cmd := newListInstancesCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetContext(app.ToContext(context.Background()))
	cmd.SetArgs([]string{})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Output is rendered to os.Stdout (not the cobra writer), so we assert on
	// successful execution rather than captured output content.
}

func TestListInstancesCmd_EmptyResults_PrintsMessage(t *testing.T) {
	mock := &cmdMockSQL{instances: []sql.Instance{}}
	sqlFactory := &cmdSqlFactory{providers: map[string]sql.SQL{"gcp": mock}}
	app := newSqlListTestApp(sqlFactory)

	var buf bytes.Buffer
	cmd := newListInstancesCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetContext(app.ToContext(context.Background()))
	cmd.SetArgs([]string{})

	// Should succeed with no error; the command prints "No SQL instances found."
	// to os.Stdout (not the cobra writer), so we assert only on the error return.
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error for empty instance list: %v", err)
	}
}

func TestListInstancesCmd_InvalidProvider_ReturnsError(t *testing.T) {
	app := newSqlListTestApp(&cmdSqlFactory{})

	var buf bytes.Buffer
	cmd := newListInstancesCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetContext(app.ToContext(context.Background()))
	cmd.SetArgs([]string{"--providers", "azure"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for unsupported provider 'azure', got nil")
	}
	if !strings.Contains(err.Error(), "unsupported") {
		t.Errorf("expected 'unsupported' in error message, got: %v", err)
	}
}

func TestListInstancesCmd_ServiceError_AllProvidersFail_ReturnsError(t *testing.T) {
	serviceErr := errors.New("GCP API unavailable")
	mock := &cmdMockSQL{err: serviceErr}
	sqlFactory := &cmdSqlFactory{providers: map[string]sql.SQL{"gcp": mock}}
	app := newSqlListTestApp(sqlFactory)

	var buf bytes.Buffer
	cmd := newListInstancesCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetContext(app.ToContext(context.Background()))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when all SQL providers fail, got nil")
	}
}
