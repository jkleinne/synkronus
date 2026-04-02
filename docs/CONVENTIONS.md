# Synkronus — Project Conventions

> This document codifies the patterns already in use in the codebase.
> Follow these conventions for all new code. Where an inconsistency
> existed, the chosen standard is noted along with the old pattern
> to migrate away from.

---

## 1. Project Overview

Synkronus is a multi-cloud infrastructure CLI tool written in Go using
Cobra/Viper. It provides a unified command surface for managing storage
buckets, objects, and SQL database instances across cloud providers.
GCP is fully implemented; AWS storage is stubbed; AWS SQL does not
exist. The codebase is ~3,700 lines across 41 files with a clean
layered architecture.

---

## 2. Tech Stack

| Dependency | Version | Purpose |
|------------|---------|---------|
| Go | 1.25.0 | Language runtime |
| `github.com/spf13/cobra` | v1.8.1 | CLI framework |
| `github.com/spf13/viper` | v1.20.1 | Configuration management |
| `github.com/go-playground/validator/v10` | v10.22.0 | Config struct validation |
| `github.com/go-viper/mapstructure/v2` | v2.3.0 | Strict config unmarshaling |
| `cloud.google.com/go/storage` | v1.56.0 | GCP Cloud Storage client |
| `cloud.google.com/go/monitoring` | v1.24.3 | GCP Cloud Monitoring (bucket usage metrics) |
| `google.golang.org/api` | v0.271.0 | GCP SQL Admin API |
| `google.golang.org/protobuf` | v1.36.11 | Protobuf types for Monitoring API |

- **Testing:** Standard `testing` package only. No assertion libraries.
- **Linting:** `go vet`. No golangci-lint config exists.

---

## 3. File and Folder Structure

```
cmd/synkronus/          CLI entry point and command definitions
  main.go               Binary entry point (imports internal/provider, calls Execute)
  root.go               Root command, global flags (--debug, --output), PersistentPreRunE
  app.go                appContainer definition, context injection, constructor
  errors.go             Shared error sentinels (ErrOperationAborted)
  providers.go          ProviderResolver struct for validating provider names
  storage_cmd.go        Storage parent command (AddCommand only)
  storage_*.go          Individual storage subcommands (one file per command)
  sql_cmd.go            SQL parent command (AddCommand only)
  sql_*.go              Individual SQL subcommands (one file per command)
  config_cmd.go         All config subcommands (set, get, delete, list)

internal/               Private application logic
  config/               Viper-based config management
  domain/               Shared domain types, interfaces, and models
    provider.go         Provider enum (GCP, AWS)
    storage/            Storage interface and models (Bucket, Object, etc.)
    sql/                SQL interface and models (Instance)
  flags/                Centralized flag name constants
  logger/               slog logger initialization
  output/               Unified output rendering
    format.go           Format enum (table, json, yaml) and ParseFormat
    render.go           Render dispatch (TableRenderer interface)
    table.go            ASCII table struct
    storage_views.go    Table renderers for storage types
    sql_views.go        Table renderers for SQL types
  provider/
    imports.go          Blank imports triggering provider init() registration
    registry/           Generic thread-safe provider registry
    factory/            Provider client creation from config (generic helpers)
    storage/gcp/        GCP storage implementation
    storage/aws/        AWS storage implementation (stubbed)
    sql/gcp/            GCP SQL implementation
  service/              Business logic (StorageService, SqlService, fanout)
  ui/prompt/            Interactive confirmation prompts
```

### Placement rules

- **New CLI commands:** Create a new file per subcommand (e.g.,
  `storage_update_bucket.go`). Only create a new `*_cmd.go` parent
  for an entirely new command group (e.g., `compute_cmd.go`).
- **New provider implementations:** Create
  `internal/provider/<service>/<provider>/` with at minimum a
  `client.go` (type, init registration, constructor, Close). Split
  operations into separate files (`buckets.go`, `objects.go`) when
  the provider is fully implemented.
- **New service types:** Add to `internal/service/` with the pattern
  `<service>_service.go`.
- **New domain models:** Add to the existing `model.go` in the
  relevant `internal/domain/<service>/` package. Include `json` and
  `yaml` struct tags on all exported fields.
- **New output view types:** Add to the appropriate views file in
  `internal/output/` (e.g., `storage_views.go` for storage types).
  Implement the `TableRenderer` interface.
- **Tests:** Co-located with source (`foo.go` → `foo_test.go` in the
  same package).

---

## 4. Code Style

### File naming

- Multi-word files: `snake_case.go` (e.g., `storage_cmd.go`,
  `storage_service.go`, `sql_formatter.go`)
- Single-word files: lowercase (e.g., `app.go`, `root.go`, `logger.go`)
- Test files: `<source>_test.go` in the same package

### Package naming

- Lowercase, single word (e.g., `config`, `registry`, `formatter`,
  `gcp`, `aws`)
- No underscores, no camelCase

### Type naming

- Exported types: `PascalCase` (e.g., `StorageService`, `GCPStorage`,
  `ConfigManager`)
- Unexported types: `camelCase` (e.g., `appContainer`, `storageFlags`,
  `contextKey`)

### Function naming

- Exported: `PascalCase` — verb-first (e.g., `ListBuckets`,
  `GetConfiguredProviders`, `FormatBucketList`)
- Unexported: `camelCase` — verb-first (e.g., `isConfigured`,
  `initialize`, `getStorageClient`, `mapInstance`)
- Constructors: `NewXxx` — always (e.g., `NewFactory`,
  `NewGCPStorage`, `NewStorageFormatter`)

### Variable naming

- Booleans: `isXxx`, `hasXxx` prefix (e.g., `isUBLAEnabled`,
  `hasConditions`)
- Error sentinels: `ErrXxx` with `errors.New()` (e.g.,
  `ErrOperationAborted`, `ErrMetricsNotFound`)
- Constants: `PascalCase` for exported, `camelCase` for unexported
  (e.g., `ConfigFileName`, `metricTimeWindow`)

### Method receivers

- Single letter matching the type's first letter: `s` for Service,
  `f` for Factory/Formatter, `g` for GCP types, `r` for Registry,
  `t` for Table
- Two-letter abbreviation when single letter would collide:
  `cm` for ConfigManager
- Always pointer receivers unless the type is immutable

### Import ordering

Three groups separated by blank lines, in this order:

```go
import (
    // 1. Standard library
    "context"
    "fmt"

    // 2. External dependencies
    "github.com/spf13/cobra"

    // 3. Internal packages (synkronus/...)
    "synkronus/internal/config"
    "synkronus/internal/domain/storage"
)
```

This is already 100% consistent across all files. Do not deviate.

### File header comments

**Standard (resolving inconsistency):** `// File: <path>` headers
were present in ~60% of files but absent in the rest. **Decision:
remove them.** They duplicate information the editor/IDE already
provides and become stale if files move. Do not add them to new files.
Existing ones may be removed opportunistically during unrelated edits.

> *Old pattern (do not use):* `// File: internal/config/config.go`

---

## 5. Provider Implementation Patterns

This is the Go/CLI equivalent of "component patterns." Each cloud
provider implementation follows a strict structure.

### Required elements for a new provider

1. **Type definition** with interface compliance check:
   ```go
   type AWSStorage struct { ... }
   var _ storage.Storage = (*AWSStorage)(nil)
   ```

2. **Self-registration via `init()`:**
   ```go
   func init() {
       registry.RegisterProvider("aws", registry.Registration[storage.Storage]{
           ConfigCheck: isConfigured,
           Initializer: initialize,
       })
   }
   ```

3. **Unexported helpers** `isConfigured` and `initialize`:
   ```go
   func isConfigured(cfg *config.Config) bool { ... }
   func initialize(ctx context.Context, cfg *config.Config, logger *slog.Logger) (storage.Storage, error) { ... }
   ```

4. **Constructor** following `NewXxx` pattern:
   ```go
   func NewAWSStorage(...) (*AWSStorage, error) { ... }
   ```

5. **`ProviderName()` and `Close()`** methods

6. **Blank import** in `internal/provider/imports.go`

7. **Config struct** added to `internal/config/config.go`

### File organization within a provider package

- `client.go` — type, init, constructor, ProviderName, Close
- One file per operation group when fully implemented (e.g.,
  `buckets.go`, `objects.go`)
- `mappers.go` — SDK-to-domain-model transformation functions
- `metrics.go` — monitoring/observability API integration (if applicable)
- Single file acceptable for stubs or simple implementations

### Interface definitions

- Defined in the domain package, not alongside implementations
  (`internal/domain/storage/storage.go`, not alongside implementations)
- Named as nouns or `-er` suffix: `Storage`, `SQL`, `Prompter`
- Keep interfaces minimal. The current interfaces (6 methods for
  Storage, 4 methods for SQL including ProviderName/Close) are the
  right size.

---

## 6. Cloud API Interaction

This is the Go/CLI equivalent of "data fetching."

### Client lifecycle

- Clients are created per-command-invocation via the provider factory.
  Each command gets a fresh client, uses it, and defers `Close()`.
- No client pooling, connection reuse, or caching across commands.
  This is correct for a CLI tool.
- The monitoring client within GCPStorage is an exception — it uses
  `sync.Once` for lazy initialization within a single client's
  lifetime.

### Multi-provider operations

- Use `concurrentFanOut[C, T]()` in `internal/service/fanout.go`
- One goroutine per provider, results collected via mutex
- Partial failures return combined results + `errors.Join(errs...)`
- Each goroutine creates its own client and defers `Close()`

### Data mapping

- Cloud SDK response types are never exposed outside the provider
  package
- Map to domain models (`storage.Bucket`, `sql.Instance`) using
  explicit mapper functions
- Mapper functions are unexported and live in the provider package
- Handle nil/missing fields defensively (check before accessing)

### Where data logic lives

```
cmd/synkronus/ → calls service method, passes result to output.Render
internal/service/ → orchestrates: get client, call method, handle errors
internal/provider/<service>/<provider>/ → calls cloud SDK, maps to domain model
```

- **Never call cloud SDKs from `cmd/`.**
- **Never format output in `internal/service/`.**
- **Never import provider packages in `internal/service/`** — only the
  interface packages (`internal/domain/storage/`, `internal/domain/sql/`).

---

## 7. Error Handling

### Error creation

- Sentinel errors: `var ErrXxx = errors.New("...")` for named,
  checkable error conditions
- All other errors: `fmt.Errorf("...: %w", err)` — always wrap with
  `%w`, never `%v`
- Error messages: lowercase, start with action verb, describe what
  failed:
  - `"failed to create GCP storage client: %w"`
  - `"error getting metric data for bucket %s: %w"`

### Error propagation

```
Provider: returns fmt.Errorf("failed to <action>: %w", sdkErr)
    ↓
Service: logs error, returns fmt.Errorf("error initializing provider: %w", err)
    ↓
Command: returns fmt.Errorf("error <action> '<resource>' on %s: %w", provider, err)
    ↓
Root: fmt.Fprintf(os.Stderr, "Error: %v\n", err) + os.Exit(1)
```

- Every layer adds context about what it was doing
- Service layer logs at `slog.Error` level before returning
- Command layer wraps with user-facing context (resource name, provider)
- Root command handles final output to stderr

### Partial failures (multi-provider operations)

- `concurrentFanOut` returns `(results, errors.Join(errs...))`
- Command layer checks: if `err != nil && len(results) == 0`, return
  error. If `err != nil && len(results) > 0`, print warning to stderr
  and display partial results.

### Never swallow errors

- Every `if err != nil` must either return the error or log it at an
  appropriate level
- The only exception: graceful degradation in `DescribeBucket` where
  IAM policy fetch failure is logged as `Warn` and the section shows
  `(Could not retrieve IAM policy - check permissions)`

---

## 8. Environment and Configuration

### Configuration management

- **File:** `~/.config/synkronus/config.json`
- **Library:** Viper with strict unmarshaling (rejects unknown keys)
- **Validation:** `go-playground/validator` with `required` tags
- **Permissions:** directory `0700`, file `0600`

### Adding a new config key

1. Add the field to the appropriate config struct in
   `internal/config/config.go` with `json` and `validate` tags
2. The config system auto-handles persistence, validation, and
   secure file permissions
3. Update `isConfigured()` in the relevant provider to check the
   new field

### Environment variables

- Synkronus does not read environment variables directly
- Cloud SDK credentials are managed by the respective SDK's credential
  chain (e.g., `GOOGLE_APPLICATION_CREDENTIALS` for GCP)
- No `.env` files, no `os.Getenv()` calls in the codebase

---

## 9. Git and Version Control

### Branch naming

- `type/short-description` (e.g., `feat/add-auth`, `fix/nav-crash`)
- Types match commit types: `feat`, `fix`, `docs`, `refactor`,
  `chore`, `test`, `perf`, `ci`

### Commit messages

- **Format:** Conventional Commits — `type(scope): description`
- **Examples from history:**
  - `feat(sql): add support for SQL providers with GCP integration`
  - `refactor(provider): unify duplicated patterns with generics`
  - `fix(storage/aws): return error from stub instead of fake data`
  - `perf(storage/gcp): cache monitoring client with lazy initialization`
- **Subject line:** imperative mood, under 72 characters
- **Scope:** package or feature area (e.g., `storage/gcp`, `sql`,
  `provider`, `cli`, `service`, `formatter`)

### PRs

- One logical change per PR
- PR descriptions summarize what changed, why, and testing done
- Merge to `main` via merge commits (not squash — preserves
  individual commit history)

---

## 10. Dependency Management

- **Pin versions** in `go.mod` (Go modules does this by default)
- **Direct dependencies only:** do not add a dependency when the
  standard library suffices
- **Audit before adding:** check the dependency's maintenance status,
  license, and transitive dependency count
- **Current state:** 8 direct dependencies, all actively used, no
  redundancy (verified in TECH_SPEC.md)
- When AWS is implemented, use `github.com/aws/aws-sdk-go-v2` (v2,
  not v1)

---

## 11. Security Patterns

### Credential handling

- **Never store credentials in Synkronus config.** The config file
  holds only non-secret identifiers (`gcp.project`, `aws.region`).
- Delegate authentication entirely to cloud SDK credential chains.
- Config file permissions are enforced at `0600` with directory at
  `0700`.

### Input validation

- Provider names are validated via `ProviderResolver.Resolve()` against the
  registry (case-insensitive, deduplicated).
- **Gap:** Bucket names and object keys are not validated at the CLI
  boundary. Add validation before passing to providers.
- Config values are validated via struct tags (`required`) and strict
  unmarshaling.

### Destructive operations

- `delete-bucket` requires typing the bucket name to confirm (not
  yes/no). Bypassed only with explicit `--force` flag.
- Use `internal/ui/prompt/Prompter` interface for all future
  destructive operations.

### Error messages

- Never include stack traces, API keys, or internal file paths in
  user-facing errors
- Error messages include resource names and provider names (acceptable
  — these are user-provided inputs)

---

## 12. Common Pitfalls

### Do not register a stub provider as fully functional

The AWS storage stub registers in the provider registry and passes
config checks, creating the impression that AWS works. When AWS
operations are attempted, users see `"AWS ListBuckets is not yet
implemented"` — this looks like a bug, not a known limitation.

```go
// ✗ Wrong: registering a stub that will confuse users
func init() {
    registry.RegisterProvider("aws", registry.Registration[storage.Storage]{
        ConfigCheck: isConfigured,
        Initializer: initialize, // returns a stub that errors on every call
    })
}

// ✓ Right: either don't register until implemented, or return
// explicit "coming soon" errors that don't look like runtime failures
```

### Do not silently skip absent providers in list operations

If a user configures `aws.region` and runs `sql list`, AWS SQL is
silently excluded because it has no registered provider. No error, no
warning. The user sees only GCP results.

```go
// ✗ Wrong: user gets partial results with no indication of omission
// (current behavior when --providers is not specified)

// ✓ Right: warn when a configured provider has no implementation
// for the requested service type
```

### Do not interpolate user input into API filter strings

```go
// ✗ Wrong (metrics.go:90):
Filter: fmt.Sprintf(`...bucket_name="%s"`, bucketName),

// ✓ Right: validate bucket name against naming rules first,
// or use a parameterized query if the API supports it
```

### Do not mix date formats in similar contexts

```go
// ✗ Wrong: RFC3339 in object list, ISO date in bucket list
obj.LastModified.Format(time.RFC3339)     // object list
bucket.CreatedAt.Format("2006-01-02")     // bucket list

// ✓ Right: use "2006-01-02" for all list tables,
// time.RFC1123 for all detail views
```

---

## 13. AI-Assisted Development Instructions

When using Claude Code, Copilot, or other AI coding tools in this
repository:

### Before writing code

- Read the relevant interface definition (`internal/domain/storage/storage.go`
  or `internal/domain/sql/sql.go`) before implementing a provider
- Read existing provider implementations (GCP storage is the
  canonical reference) before creating new ones
- Check `internal/flags/flags.go` for existing flag constants before
  defining new ones
- Check `internal/output/` for existing view types and rendering patterns
  before creating new output renderers
- Check `internal/output/` view types before creating new output renderers

### When writing code

- Follow the import ordering strictly: stdlib, external, internal
  (three groups, blank line between each)
- Use `fmt.Errorf("...: %w", err)` for all error wrapping — never
  `%v`
- Use `var _ Interface = (*Type)(nil)` to verify interface compliance
  at compile time
- Add `defer client.Close()` immediately after acquiring any cloud
  SDK client
- Use the `concurrentFanOut` function for any new multi-provider
  list operation — do not write custom goroutine logic
- Match existing method receiver naming: single letter from type name

### Before presenting code as complete

- Run `go vet ./...` — must pass with no warnings
- Run `go build ./cmd/synkronus` — must compile
- Run `go test ./...` — all existing tests must pass
- Verify no unused imports, variables, or dead code
- Verify error paths are handled (no empty `if err != nil {}` blocks)

### Do not

- Add new dependencies without explicit confirmation
- Create new files without stating purpose and where they fit
- Modify public interfaces (`storage.Storage`, `sql.SQL`) without
  listing all implementations that must be updated
- Add `// File: ...` header comments to files (deprecated convention)
- Add `TODO` comments without a ticket/issue reference
- Write tests that depend on real cloud API credentials
