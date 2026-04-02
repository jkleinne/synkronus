# Synkronus — Technical Specification

> **Version:** 1.0 — 2 April 2026
> **Status:** Draft
> **Codebase snapshot:** commit `f557a9b` (main)

---

## 1. Executive Summary

Synkronus is a multi-cloud CLI tool (~3,700 lines of Go) built on Cobra/Viper with a clean layered architecture: CLI commands → service orchestration → provider implementations → cloud SDKs. The codebase is well-structured with a generic, thread-safe provider registry, concurrent multi-provider fanout with partial failure tolerance, and proper dependency injection via context. GCP storage and SQL are fully functional; AWS storage is stubbed (all methods return errors) and AWS SQL does not exist. There is no CI/CD pipeline, no integration tests, and test coverage is approximately 12% overall (4 packages tested out of 16).

**Production-readiness assessment: approximately 40% ready.** The architecture and GCP implementation are solid, but the tool cannot deliver its core value proposition (multi-cloud management) with only one working provider. No build pipeline, no release mechanism, no user documentation, and minimal test coverage.

---

## 2. Architecture Overview

### 2.1 Component Diagram

```
┌─────────────────────────────────────────────────────────┐
│                    CLI Layer (Cobra)                     │
│   cmd/synkronus/                                        │
│   ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐  │
│   │ root.go  │ │storage_  │ │ sql_     │ │config_   │  │
│   │          │ │cmd.go    │ │ cmd.go   │ │cmd.go    │  │
│   └────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘  │
│        │    PersistentPreRunE     │             │        │
│        ▼             │           │             │        │
│   ┌──────────────┐   │           │             │        │
│   │ appContainer │◄──┘───────────┘─────────────┘        │
│   │ (DI root)    │  injected via context.Context        │
│   └──────┬───────┘                                      │
└──────────┼──────────────────────────────────────────────┘
           │
           ▼
┌──────────────────────────────────────────────────────────┐
│                  Service Layer                           │
│   internal/service/                                      │
│   ┌─────────────────┐  ┌──────────────┐                  │
│   │ StorageService   │  │ SqlService   │                  │
│   │ (6 operations)   │  │ (2 operations)│                 │
│   └────────┬─────────┘  └──────┬───────┘                 │
│            │                   │                          │
│            ▼                   ▼                          │
│   ┌──────────────────────────────────┐                    │
│   │ concurrentFanOut[C,T]()          │  (list operations) │
│   │ goroutine-per-provider + mutex   │                    │
│   └──────────────┬───────────────────┘                    │
└──────────────────┼───────────────────────────────────────┘
                   │
                   ▼
┌──────────────────────────────────────────────────────────┐
│            Provider Factory + Registry                   │
│   internal/provider/                                     │
│   ┌───────────────────┐  ┌─────────────────────────┐     │
│   │ Factory           │  │ Registry[T] (generic)   │     │
│   │ GetStorageProvider│──│ storageRegistry          │     │
│   │ GetSqlProvider    │  │ sqlRegistry              │     │
│   └───────────────────┘  └────────────┬────────────┘     │
└───────────────────────────────────────┼──────────────────┘
                                        │ init() registration
                   ┌────────────────────┼───────────────┐
                   ▼                    ▼               ▼
┌──────────────────────────────────────────────────────────┐
│              Provider Implementations                    │
│   internal/provider/storage/    internal/provider/sql/   │
│   ┌──────────┐ ┌──────────┐ ┌──────────┐                │
│   │ gcp/     │ │ aws/     │ │ gcp/     │                 │
│   │ (full)   │ │ (stub)   │ │ (full)   │                 │
│   └────┬─────┘ └──────────┘ └────┬─────┘                │
│        │                         │                       │
└────────┼─────────────────────────┼───────────────────────┘
         ▼                         ▼
┌──────────────────────────────────────────────────────────┐
│             Cloud SDKs                                   │
│   cloud.google.com/go/storage    google.golang.org/api   │
│   cloud.google.com/go/monitoring (sqladmin v1)           │
└──────────────────────────────────────────────────────────┘
```

### 2.2 Supporting Components

```
┌──────────────────────┐  ┌──────────────────────┐
│ internal/config/     │  │ internal/output/      │
│ Viper-based config   │  │ Unified output        │
│ ~/.config/synkronus/ │  │ rendering (table,     │
│ config.json          │  │ JSON, YAML)           │
└──────────────────────┘  └──────────────────────┘
┌──────────────────────┐  ┌──────────────────────┐
│ internal/logger/     │  │ internal/ui/prompt/   │
│ slog to stderr       │  │ Value-matching        │
│                      │  │ confirmation          │
└──────────────────────┘  └──────────────────────┘
┌──────────────────────┐
│ internal/domain/     │
│ Domain types         │
│ (Provider enum,      │
│ models, interfaces)  │
└──────────────────────┘
```

### 2.3 Development vs Production

There is no distinction between development and production environments. The tool is a local CLI binary that authenticates using ambient cloud credentials (GCP Application Default Credentials, AWS credential chain). No proxy, mock service, or environment switching mechanism exists. Building is `go build ./cmd/synkronus`; running tests is `go test ./...`.

---

## 3. Data Model

### 3.1 Overview

Synkronus does not use a database. It is a stateless CLI that queries cloud APIs on each invocation and formats results for terminal output. The only persistent state is the configuration file at `~/.config/synkronus/config.json`.

### 3.2 Domain Models

All domain models are defined as Go structs in `internal/domain/storage/model.go` and `internal/domain/sql/model.go`. They are provider-agnostic representations mapped from cloud SDK response types.

#### Storage Domain

```
Bucket
├── Name               string
├── Provider            domain.Provider
├── Location            string
├── LocationType        string         (GCP-specific: "multi-region", "dual-region", etc.)
├── StorageClass        string
├── CreatedAt           time.Time
├── UpdatedAt           time.Time
├── UsageBytes          int64          (-1 = unknown)
├── RequesterPays       bool
├── Labels              map[string]string
├── Autoclass           *Autoclass     (Enabled bool)
├── IAMPolicy           *IAMPolicy
│   ├── Bindings        []IAMBinding   (Role + Principals)
│   └── HasConditions   bool           (conditional bindings detected but not displayed)
├── ACLs                []ACLRule      (Entity + Role)
├── LifecycleRules      []LifecycleRule
│   ├── Action          string
│   └── Condition       LifecycleCondition (Age, CreatedBefore, MatchesStorageClass, NumNewerVersions)
├── Logging             *Logging       (LogBucket + LogObjectPrefix)
├── Versioning          *Versioning    (Enabled bool)
├── SoftDeletePolicy    *SoftDeletePolicy (RetentionDuration)
├── UniformBucketLevelAccess *UniformBucketLevelAccess (Enabled bool)
├── PublicAccessPrevention  string
├── Encryption          *Encryption    (KmsKeyName + Algorithm)
└── RetentionPolicy     *RetentionPolicy (RetentionPeriod + IsLocked)

ObjectList
├── BucketName          string
├── Prefix              string
├── Objects             []Object
└── CommonPrefixes      []string       (simulated directories)

Object
├── Key                 string
├── Bucket              string
├── Provider            domain.Provider
├── Size                int64
├── StorageClass        string
├── LastModified        time.Time
├── CreatedAt           time.Time
├── UpdatedAt           time.Time
├── ETag                string
├── ContentType/Encoding/Language/CacheControl/Disposition  string
├── MD5Hash             string         (Base64)
├── CRC32C              string         (GCP-specific, Base64)
├── Generation          int64          (GCP-specific versioning)
├── Metageneration      int64          (GCP-specific versioning)
├── Encryption          *Encryption
└── Metadata            map[string]string
```

#### SQL Domain

```
Instance
├── Name                string
├── Provider            domain.Provider
├── Region              string
├── DatabaseVersion     string
├── Tier                string
├── State               string
├── PrimaryAddress      string
├── StorageSizeGB       int64
├── CreatedAt           time.Time
├── Project             string         (GCP-specific)
├── ConnectionName      string         (GCP-specific)
└── Labels              map[string]string
```

### 3.3 Data Flow

```
Cloud API  ──SDK response──►  Mapper (gcp/mappers.go)  ──domain model──►  Service  ──►  output.Render  ──►  stdout
```

Each provider package contains mapper functions that translate cloud SDK types into these domain models. GCP storage mappers are in `internal/provider/storage/gcp/mappers.go` (113 lines); GCP SQL mapping is inline in `internal/provider/sql/gcp/gcp.go:96-128`.

### 3.4 Configuration Schema

```json
{
  "gcp": {
    "project": "string (required if GCP enabled)"
  },
  "aws": {
    "region": "string (required if AWS enabled)"
  }
}
```

Stored at `~/.config/synkronus/config.json` with permissions `0600` (file) and `0700` (directory). Strict unmarshaling rejects unknown keys. Validation uses `go-playground/validator`.

### 3.5 Schema Issues

| Issue | Severity | Details |
|-------|----------|---------|
| GCP-specific fields on shared models | Low | `LocationType`, `CRC32C`, `Generation`, `Metageneration`, `Project`, `ConnectionName` are GCP concepts with no AWS equivalent. These fields will be zero-valued when populated by AWS providers, which is acceptable for display but should be documented. |
| `UsageBytes` sentinel value | Low | Uses `-1` to indicate "unknown". A pointer (`*int64`) or a wrapper type would be more idiomatic, but the current approach works and is documented in the struct comment. |

---

## 4. API Surface

Synkronus is a CLI tool, not a server. Its "API surface" is the set of Cobra commands exposed to users.

### 4.1 Command Reference

#### Storage Commands (`cmd/synkronus/storage_*.go`)

| Command | Args | Required Flags | Optional Flags | Purpose |
|---------|------|----------------|----------------|---------|
| `storage list-buckets` | — | — | `--providers, -p` (csv) | List buckets across providers |
| `storage describe-bucket <name>` | bucket name | `--provider, -p` | — | Detailed bucket metadata |
| `storage create-bucket <name>` | bucket name | `--provider, -p`, `--location, -l` | — | Create a new bucket |
| `storage delete-bucket <name>` | bucket name | `--provider, -p` | `--force, -f` | Delete bucket (with confirmation) |
| `storage list-objects` | — | `--provider, -p`, `--bucket, -b` | `--prefix` | List objects in bucket |
| `storage describe-object <key>` | object key | `--provider, -p`, `--bucket, -b` | — | Object metadata |

#### SQL Commands (`cmd/synkronus/sql_*.go`)

| Command | Args | Required Flags | Optional Flags | Purpose |
|---------|------|----------------|----------------|---------|
| `sql list` | — | — | `--providers, -p` (csv) | List SQL instances across providers |
| `sql describe <name>` | instance name | `--provider, -p` | — | Detailed instance metadata |

#### Config Commands (`cmd/synkronus/config_cmd.go`)

| Command | Args | Purpose |
|---------|------|---------|
| `config set <key> <value>` | key, value | Set config value with validation |
| `config get <key>` | key | Retrieve config value |
| `config delete <key>` | key | Remove config key |
| `config list` | — | Display all settings |

#### Global Flags

| Flag | Type | Default | Purpose |
|------|------|---------|---------|
| `--debug, -d` | bool | false | Enable `slog.LevelDebug` logging to stderr |
| `--output, -o` | string | table | Output format: table, json, yaml |

### 4.2 Command Aliases

| Primary | Alias |
|---------|-------|
| `storage list-buckets` | `storage list` |
| `storage describe-bucket` | `storage describe` |
| `storage create-bucket` | `storage create` |
| `storage delete-bucket` | `storage delete` |
| `sql list` | `sql list-instances` |
| `sql describe` | `sql describe-instance` |

### 4.3 Output Format

Output defaults to ASCII table format to stdout. JSON and YAML output are available via the `--output` flag (e.g., `--output json`). Errors and warnings go to stderr.

### 4.4 Authentication

Authentication is delegated entirely to the cloud SDKs:
- **GCP:** Application Default Credentials (ADC) — `gcloud auth application-default login` or service account key
- **AWS:** Standard credential chain — environment variables, `~/.aws/credentials`, IAM role

Synkronus stores no credentials. The config file contains only non-secret identifiers (`gcp.project`, `aws.region`).

### 4.5 Inconsistencies and Gaps

| Issue | Severity | Details |
|-------|----------|---------|
| No `version` command | Low | No way to identify installed version or build info |
| No command examples in help | Low | Cobra `Example` field unused on all commands |
| Alias direction inconsistent | Low | Storage uses verb-noun as primary (`list-buckets`) with short alias (`list`); SQL uses short as primary (`list`) with verb-noun alias (`list-instances`) |

---

## 5. Dependencies Audit

### 5.1 Direct Dependencies (9 packages)

| Package | Version | Status | Purpose |
|---------|---------|--------|---------|
| `cloud.google.com/go/monitoring` | v1.24.3 | **Active** | Bucket usage metrics via Cloud Monitoring API |
| `cloud.google.com/go/storage` | v1.56.0 | **Active** | GCP Cloud Storage client |
| `github.com/go-playground/validator/v10` | v10.22.0 | **Active** | Config struct validation |
| `github.com/go-viper/mapstructure/v2` | v2.3.0 | **Active** | Strict config unmarshaling |
| `github.com/spf13/cobra` | v1.8.1 | **Active** | CLI framework |
| `github.com/spf13/viper` | v1.20.1 | **Active** | Configuration management |
| `google.golang.org/api` | v0.271.0 | **Active** | GCP SQL Admin API (sqladmin/v1) |
| `google.golang.org/protobuf` | v1.36.11 | **Active** | Protobuf types for Monitoring API requests |
| `gopkg.in/yaml.v3` | v3.0.1 | **Active** | YAML output serialization |

### 5.2 Assessment

**The dependency tree is clean.** All 9 direct dependencies are actively used, none are redundant, and all serve distinct purposes. There are no unused imports or dead dependencies.

### 5.3 Indirect Dependencies

The project pulls ~70 transitive dependencies, primarily from the GCP SDK ecosystem (OpenTelemetry, gRPC, OAuth2, x/crypto, etc.). This is expected and unavoidable when using Google Cloud Go client libraries.

### 5.4 Go Version

The project targets **Go 1.25.0** (per `go.mod`). This is the current stable release and gives access to all modern Go features including generics (used in the registry and fanout).

### 5.5 Missing Dependencies

When AWS is implemented, the following will be needed:
- `github.com/aws/aws-sdk-go-v2` (core)
- `github.com/aws/aws-sdk-go-v2/service/s3`
- `github.com/aws/aws-sdk-go-v2/service/rds`
- `github.com/aws/aws-sdk-go-v2/config`

---

## 6. Environment and Configuration

### 6.1 Configuration Management

| Aspect | Implementation | File Reference |
|--------|---------------|----------------|
| Config file path | `~/.config/synkronus/config.json` | `internal/config/config.go:51` |
| File permissions | `0600` (rw-------) | `internal/config/config.go:22` |
| Directory permissions | `0700` (rwx------) | `internal/config/config.go:20` |
| Fallback search path | Current working directory | `internal/config/config.go:56` |
| Strict parsing | Rejects unknown keys via mapstructure `ErrorUnused` | `internal/config/config.go:171` |
| Validation | `go-playground/validator` with `required` tags | `internal/config/config.go:187-204` |
| Revert on failure | `v.ReadInConfig()` called on validation failure | `internal/config/config.go:109,114,139,144` |

### 6.2 Environment Variables

**Synkronus itself does not read environment variables.** All configuration is file-based. Cloud SDK authentication relies on environment variables managed by the respective SDKs:

- `GOOGLE_APPLICATION_CREDENTIALS` — GCP service account key path
- `GOOGLE_CLOUD_PROJECT` — (not used; Synkronus requires explicit `gcp.project` in config)
- `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_REGION` — AWS SDK standard variables (not yet used)

### 6.3 Issues

| Issue | Severity | Details |
|-------|----------|---------|
| No `.env.example` or config template | Low | Users must discover valid config keys through `config set` error messages or docs. No example config file ships with the project. |
| No config validation for values | Low | The config system validates presence of keys (`required` tag) but not value format. `gcp.project = "!!!"` would pass validation but fail at API call time. |
| Config fallback to CWD | Low | `internal/config/config.go:56` adds `.` to Viper search path. A `config.json` in the working directory would override `~/.config/synkronus/config.json`. This could be surprising but is a standard Viper pattern. |

### 6.4 Secrets Management

**No secrets are stored by Synkronus.** The configuration file contains only:
- `gcp.project` — GCP project ID (not a secret)
- `aws.region` — AWS region name (not a secret)

Authentication credentials are managed entirely by the cloud SDKs using their native credential chain. This is the correct approach.

---

## 7. Security Assessment

### 7.1 Critical Severity

**None identified.**

### 7.2 Moderate Severity

| Finding | Location | Details | Recommendation |
|---------|----------|---------|----------------|
| **Unescaped string interpolation in Monitoring API filter** | `internal/provider/storage/gcp/metrics.go:90` | Bucket names are interpolated directly into the Cloud Monitoring API filter string: `fmt.Sprintf(\`...\"%s\"\`, bucketName)`. A bucket name containing double quotes could break the filter syntax or alter query semantics. While this cannot execute arbitrary code (it's a read-only monitoring query, not SQL), it can cause unexpected behavior or information disclosure. | Validate bucket names against cloud naming rules (3-63 chars, lowercase, alphanumeric, hyphens, dots) before use in API calls. |
| **No input validation on resource names** | `cmd/synkronus/storage_*.go` | Bucket names and object keys from CLI args are passed directly to provider methods without validation. Invalid names produce cloud API errors that may expose internal details. | Add validation at the command level that matches cloud provider naming rules. |

### 7.3 Low Severity

| Finding | Location | Details |
|---------|----------|---------|
| **Config values visible in `config list`** | `cmd/synkronus/config_cmd.go:126` | Currently only non-secret values (project ID, region) are stored. If credentials are ever added to config, they would be displayed in plain text. |
| **Debug logging may expose operational details** | `internal/logger/logger.go` | When `--debug` is enabled, structured log fields include bucket names, provider names, and error details. This is appropriate for a CLI tool but should be documented. |
| **No request timeouts** | Service layer | No explicit timeouts are set on cloud API calls beyond SDK defaults. A hung API call would block the CLI indefinitely. |

### 7.4 Positive Security Characteristics

- **No credential storage** — authentication is fully delegated to cloud SDK credential chains
- **Secure config file permissions** — defense-in-depth with both directory (`0700`) and file (`0600`) permissions
- **Strict config parsing** — rejects unknown keys, preventing config injection
- **Config revert on failure** — invalid values cannot corrupt the config file
- **Destructive operation confirmation** — delete operations require typing the resource name (not just "yes")
- **No network listeners** — the tool is a CLI binary, not a server; no attack surface from incoming connections

---

## 8. Performance Assessment

### 8.1 Applicability

Synkronus is a CLI tool, not a web application. Lighthouse scores, bundle size analysis, Core Web Vitals, and render-blocking resource assessments are **not applicable**.

### 8.2 CLI Performance Characteristics

| Aspect | Assessment | Details |
|--------|-----------|---------|
| **Binary size** | Acceptable | Standard Go binary with GCP SDK dependencies. Expected ~30-50MB (typical for GCP client libraries). |
| **Startup time** | Good | Initialization is lazy — cloud SDK clients are created only when a command runs, not at binary startup. The `appContainer` is constructed in `PersistentPreRunE`, after flag parsing. |
| **List operation latency** | Concern | `ListBuckets` makes 2 API calls: one to GCS for bucket metadata, one to Cloud Monitoring for usage metrics. The monitoring call has a 72-hour aggregation window. With many providers, the fanout runs all in parallel, bounded by the slowest provider. |
| **Client instantiation** | Concern | Each single-resource command creates a new cloud SDK client, uses it once, and closes it. There is no client pooling or caching across commands. For a CLI tool where each invocation is independent, this is acceptable. |
| **Memory usage** | Acceptable | All results are loaded into memory before formatting. For extremely large result sets (thousands of objects), this could be an issue, but typical usage won't hit this. |
| **Monitoring client caching** | Good | `sync.Once` ensures the Cloud Monitoring client is initialized once per GCP storage client lifetime (`internal/provider/storage/gcp/client.go:67-72`). |

### 8.3 Performance Risks

| Risk | Severity | Affected Operation | Mitigation |
|------|----------|--------------------|------------|
| **No pagination** | Medium | `list-objects` with large buckets | All objects matching a prefix are fetched and held in memory. Buckets with millions of objects could cause high memory usage and long response times. Add pagination or streaming output. |
| **Monitoring API adds latency to ListBuckets** | Low | `storage list-buckets` | Usage metrics are fetched for every ListBuckets call. Consider a `--no-metrics` flag or caching metrics locally with a TTL. |
| **No request timeouts** | Low | All operations | Cloud SDK default timeouts apply. An explicit timeout (e.g., 30 seconds) via `context.WithTimeout` would prevent indefinite hangs. |

---

## 9. Technical Debt

| # | Issue | Severity | Affected Files | Effort | Recommended Fix |
|---|-------|----------|---------------|--------|-----------------|
| 1 | **AWS storage stub registers as a valid provider** — users get confusing "not yet implemented" errors mixed with real results | High | `internal/provider/storage/aws/aws.go`, `internal/provider/imports.go` | Small | Either remove from registry until implemented, or replace errors with clear "coming soon" messages. |
| 2 | **AWS SQL is completely absent but silently skipped** — `sql list --providers gcp,aws` returns only GCP results with no indication AWS was ignored | High | `internal/provider/imports.go`, `cmd/synkronus/sql_*.go` | Small | The `ProviderResolver` call in sql commands already validates against `registry.IsSqlSupported`, so requesting AWS SQL returns an error. However, when no `--providers` flag is passed, AWS is silently excluded from defaults because it's not registered. Document this behavior or warn when configured providers have no SQL support. |
| 3 | **No machine-readable output format** (RESOLVED — `--output` flag with table/json/yaml now implemented in `internal/output/`) | Medium | `internal/output/` | Medium | Add `--output json` flag. The formatter layer already centralizes all output; add a JSON formatter alongside the table formatter. |
| 4 | **Conditional IAM bindings silently skipped** — TODO at `internal/provider/storage/gcp/buckets.go` | Medium | `internal/provider/storage/gcp/buckets.go` | Small | Display condition title and expression. The data is already available from the API response. |
| 5 | **Abandoned cost estimation branch** — billing integration work stashed and never merged | Low | Git stash (commits `6cdca3f`, `1b0f2c3`, `7e62ff8`) | N/A | Delete stashed branch to reduce confusion. If cost estimation is revisited, start fresh. |
| 6 | **Binary artifacts in repository** — `cmd/synkronus/synkronus` and `synkronus` at root (RESOLVED — `.gitignore` added, binary artifacts removed from tracking) | Low | Repository root | Small | Add to `.gitignore`. |
| 7 | **SQL Instance model has GCP-specific fields** — `Project` and `ConnectionName` are GCP concepts with no AWS equivalent | Low | `internal/domain/sql/model.go` | Small | Acceptable for now. When AWS SQL is implemented, these fields will be zero-valued for AWS instances. Consider a provider-specific metadata map if more provider-specific fields emerge. |
| 8 | **Formatter package has no tests** — 704 lines of formatting logic with zero test coverage (RESOLVED — Output rendering moved to `internal/output/` with 26 tests covering table, format, render, and view types) | Medium | `internal/output/` | Medium | Formatter bugs silently produce incorrect output. Add table snapshot tests. |
| 9 | **GCP storage provider has no tests** — 584 lines covering buckets, objects, metrics, and mappers | Medium | `internal/provider/storage/gcp/` | Large | Most complex package in the codebase. Integration tests against fake-gcs-server would provide the most value. Unit tests for `mappers.go` are straightforward. |

---

## 10. Testing Strategy

### 10.1 Current State

| Package | Test File | Coverage | What's Tested |
|---------|-----------|----------|---------------|
| `internal/provider/registry` | `registry_test.go` | **71.4%** | Registration, deduplication (panics), lookup, case-insensitive matching, defensive copy |
| `internal/service` | `fanout_test.go` | **26.1%** | Concurrent fanout: all succeed, all fail, partial failure |
| `internal/provider/sql/gcp` | `gcp_test.go` | **36.1%** | Instance mapping: nil settings, full settings |
| `internal/provider/storage/aws` | `aws_test.go` | **62.5%** | All stub methods return "not yet implemented" |
| `internal/output` | multiple | **~high** | 26 tests covering table, format, render, and view types |
| `internal/config` | config_test.go | **~high** | 10 tests: set/get/delete/list round-trip, strict unmarshaling, validation |
| `internal/ui/prompt` | prompt_test.go | **~high** | 4 tests: correct name, wrong name, EOF, empty input |
| `cmd/synkronus` | multiple | **~high** | 13 tests: 6 ProviderResolver + 7 integration |
| All other packages (8) | — | **0.0%** | Not tested |

**Overall estimated coverage: ~30%** (weighted by lines of code).

**Framework:** Standard `testing` package. No test runner, assertion library, or mocking framework installed.

**No integration tests.** No end-to-end tests. No test infrastructure (Docker, emulators, CI).

### 10.2 Recommended Testing Plan

#### Pre-Launch (Ordered by Risk)

| Priority | Test | Package | Type | Effort | Rationale |
|----------|------|---------|------|--------|-----------|
| 1 ✓ | **Config set/get/delete/list round-trip** | `internal/config` | Unit | Small | Config is the only persistent state. Corruption means data loss. Tests should cover: set valid key, reject unknown key, revert on validation failure, strict unmarshaling. |
| 2 | **GCP mapper unit tests** | `internal/provider/storage/gcp` | Unit | Small | `mappers.go` translates SDK responses to domain models. Incorrect mapping silently corrupts output. Test each mapper with representative SDK response structs. |
| 3 ✓ | **Output views tests** (formerly formatter snapshot tests) | `internal/output` | Unit | Medium | 26 tests covering table, format, render, and view types. |
| 4 ✓ | **Command-level integration tests** | `cmd/synkronus` | Integration | Medium | Test that Cobra commands parse flags correctly, resolve providers, and call services with expected arguments. Use a mock factory. |
| 5 ✓ | **Provider resolution edge cases** | `cmd/synkronus` | Unit | Small | `ProviderResolver` handles deduplication, normalization, validation. Test: empty input, duplicates, unknown providers, mixed valid/invalid. |

#### Post-Launch

| Test | Type | Effort | Rationale |
|------|------|--------|-----------|
| **GCP storage integration tests** | Integration | Large | Test against `fake-gcs-server` or a dedicated test project. Covers the full create → list → describe → delete lifecycle. |
| **AWS provider tests** | Integration | Large | Test against LocalStack. Must be done when AWS is implemented. |
| **Concurrent fanout race detection** | Unit | Small | Run fanout tests with `-race` flag. No dedicated race tests exist. |
| ✓ **Prompt confirmation tests** | Unit | Small | Test `StandardPrompter` with various inputs: correct name, wrong name, EOF, empty input. |

### 10.3 Recommended Setup

- **Framework:** Continue with standard `testing` package (no need for testify/gomega — the codebase is idiomatic Go)
- **CI integration:** `go test -race -cover ./...` in GitHub Actions
- **Golden files:** For formatter tests, use `testdata/` directories with `.golden` files and `go test -update` flag pattern

---

## 11. Deployment Readiness

### 11.1 Current State

| Component | Status | Details |
|-----------|--------|---------|
| **CI/CD pipeline** | Implemented | GitHub Actions: build, vet, test on PRs and pushes to main (`.github/workflows/ci.yml`) |
| **Release mechanism** | Missing | No GoReleaser, no published binaries, no Homebrew formula |
| **Version embedding** | Missing | No `ldflags` injection of version/commit/date. No `version` command. |
| **Container image** | Missing | No Dockerfile for containerized usage |
| **Monitoring** | N/A | CLI tool — no server monitoring needed |
| **Logging** | Implemented | `slog` to stderr, debug toggle via `--debug` |
| **Shell completions** | Missing | Cobra supports `completion` subcommand generation but it hasn't been wired up |
| **Man pages** | Missing | Cobra supports `doc` generation but it hasn't been configured |
| **`.gitignore`** | Implemented | Binary artifacts excluded |

### 11.2 Pre-Production Checklist

#### Must-Have (Launch Blockers)

| # | Item | Effort | Details |
|---|------|--------|---------|
| 1 | **Implement AWS S3 storage** | Large | Implement all 6 `storage.Storage` interface methods using AWS SDK for Go v2. |
| 2 | **Implement AWS RDS SQL** | Medium | Implement `sql.SQL` interface. Register via `init()`. Add blank import. |
| 3 ✓ | **Set up CI/CD** | Medium | GitHub Actions workflow: `go vet`, `go test -race -cover`, `go build` on every PR. |
| 4 | **Set up release pipeline** | Medium | GoReleaser config for cross-compilation and binary publishing. GitHub Releases as distribution channel. |
| 5 | **Write user documentation** | Medium | README with: project description, installation, quickstart (configure → first command), command reference, authentication setup per provider. |
| 6 ✓ | **Add `--output json` flag** | Medium | Implemented: table, json, yaml via `--output` flag. |
| 7 | **Add `version` command** | Small | Inject version, commit SHA, and build date via `ldflags` at build time. |
| 8 ✓ | **Add `.gitignore`** | Small | Exclude built binaries, IDE files, and OS-specific files. |
| 9 ✓ | **Add config validation tests** | Small | Round-trip tests for the config system (see testing plan). |
| 10 | **Validate resource names at CLI boundary** | Small | Add bucket name and object key validation before passing to providers. |

#### Should-Have (First Week Post-Launch)

| # | Item | Effort | Details |
|---|------|--------|---------|
| 11 | **Shell completion support** | Small | `synkronus completion bash/zsh/fish/powershell` |
| 12 ✓ | **Formatter tests** | Medium | Output views tests implemented in `internal/output/` (26 tests). |
| 13 | **GCP mapper tests** | Small | Unit tests for `internal/provider/storage/gcp/mappers.go`. |
| 14 | **Add command examples to help text** | Small | Populate Cobra `Example` field for every command. |
| 15 | **Add request timeouts** | Small | `context.WithTimeout` wrapper in service layer with a sensible default (30s). |
| 16 | **Display conditional IAM bindings** | Small | Resolve TODO at `internal/provider/storage/gcp/buckets.go`. |

#### Nice-to-Have (Defer)

| # | Item | Effort | Details |
|---|------|--------|---------|
| 17 | **Azure provider** | Large | Third major cloud provider. Same pattern as GCP/AWS. |
| 18 | **Pagination for list-objects** | Medium | Streaming or paged output for large result sets. |
| 19 | **Homebrew formula** | Small | `brew install synkronus` for macOS users. |
| 20 | **Docker image** | Small | Multi-stage Dockerfile for containerized usage in CI/CD pipelines. |
| 21 | **Integration test infrastructure** | Large | fake-gcs-server and LocalStack for provider integration tests. |
