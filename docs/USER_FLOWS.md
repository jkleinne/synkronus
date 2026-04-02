# Synkronus — User Flows

> **Version:** 1.0 — 2 April 2026
>
> Synkronus is a terminal CLI tool, not a web or mobile application.
> This document maps user flows through the **command tree** rather than
> screens/pages. The CLI equivalents are:
>
> | Web/Mobile Concept | CLI Equivalent |
> |--------------------|----------------|
> | Screen / page | Command output |
> | Route / URL | Command + flags |
> | Navigation / sidebar | Command tree (`synkronus <group> <action>`) |
> | Modal / dialog | Interactive confirmation prompt |
> | Authentication | Cloud SDK credential state |
> | Deep link | Full command string (shareable, scriptable) |
> | Responsive layout | Terminal width handling |

---

## 1. Entry Points

Users interact with Synkronus exclusively through the terminal. There is
no web interface, no GUI, no API server.

| Entry Point | How User Arrives | What Happens |
|-------------|-----------------|--------------|
| **Direct invocation** | Types `synkronus <command>` in terminal | Cobra parses command, runs `PersistentPreRunE` (builds `appContainer`), executes command |
| **First run (no config)** | Types any command before running `config set` | Config loads with empty/nil provider sections. Commands that need a provider fail with: `No providers configured. Use 'synkronus config set'.` |
| **Shell script / automation** | CLI command embedded in a script | Same as direct invocation. No interactive mode. Confirmation prompts can be bypassed with `--force`. |
| **Help flag** | `synkronus --help` or `synkronus <group> --help` | Cobra auto-generated help text. No custom help pages. |
| **No arguments** | Types `synkronus` alone | Cobra prints root help text (description + available commands). |

**Not implemented:**
- No shell completions (tab-complete for commands/flags)
- No installer or first-run wizard
- No `synkronus init` command for guided setup
- No man pages

---

## 2. Command Inventory

Every command that exists in the codebase, numbered for flow references.

### Config Commands

| # | Command | Purpose |
|---|---------|---------|
| C1 | `synkronus config set <key> <value>` | Set a provider configuration value |
| C2 | `synkronus config get <key>` | Retrieve a single configuration value |
| C3 | `synkronus config delete <key>` | Remove a configuration value |
| C4 | `synkronus config list` | Display all current configuration |

### Storage Commands

| # | Command | Purpose |
|---|---------|---------|
| S1 | `synkronus storage list-buckets [--providers p1,p2]` | List buckets across one or more providers |
| S2 | `synkronus storage describe-bucket <name> --provider p` | Show detailed bucket metadata |
| S3 | `synkronus storage create-bucket <name> --provider p --location l` | Create a new storage bucket |
| S4 | `synkronus storage delete-bucket <name> --provider p [--force]` | Delete a bucket (with confirmation) |
| S5 | `synkronus storage list-objects --provider p --bucket b [--prefix x]` | List objects/directories in a bucket |
| S6 | `synkronus storage describe-object <key> --provider p --bucket b` | Show detailed object metadata |

### SQL Commands

| # | Command | Purpose |
|---|---------|---------|
| Q1 | `synkronus sql list [--providers p1,p2]` | List SQL instances across providers |
| Q2 | `synkronus sql describe <name> --provider p` | Show detailed SQL instance metadata |

### Meta Commands

| # | Command | Purpose |
|---|---------|---------|
| M1 | `synkronus` (no args) | Display root help text |
| M2 | `synkronus --help` | Display root help with flag details |
| M3 | `synkronus <group> --help` | Display group-level help |
| M4 | `synkronus <group> <command> --help` | Display command-level help |

**Total: 16 commands** (4 config + 6 storage + 2 SQL + 4 help variants)

---

## 3. Primary User Flows

### Flow 1: First-Time Setup

The onboarding flow for a new user who has just built/installed the binary.

```
User installs binary (go build / download)
│
├─→ synkronus (M1)
│   Output: Root help text listing available command groups
│   User sees: "storage", "sql", "config" subcommands
│
├─→ synkronus config set gcp.project my-project-id (C1)
│   Output: "Configuration set: gcp.project = my-project-id"
│   Effect: Creates ~/.config/synkronus/config.json with 0600 permissions
│
├─→ synkronus config set aws.region us-east-1 (C1)
│   Output: "Configuration set: aws.region = us-east-1"
│   Effect: Adds AWS config to existing config file
│
└─→ synkronus config list (C4)
    Output:
      Current configuration:
        aws.region = us-east-1
        gcp.project = my-project-id
```

**Prerequisite not handled by Synkronus:** The user must have already
authenticated with their cloud provider (e.g., `gcloud auth application-default login`
for GCP). Synkronus provides no guidance on this step.

---

### Flow 2: List and Inspect Storage Buckets

The most common operational flow — surveying storage resources.

```
synkronus storage list-buckets (S1)
│   Output: Table of all buckets across all configured providers
│   Columns: BUCKET NAME | PROVIDER | LOCATION | USAGE | STORAGE CLASS | CREATED
│
│   User identifies a bucket of interest
│
└─→ synkronus storage describe-bucket my-bucket --provider gcp (S2)
    Output: Full detail view with sections:
      == Bucket: my-bucket ==
      -- Overview --           (provider, location, storage class, usage, dates)
      -- Access Control --     (UBLA, public access, logging, IAM, ACLs)
      -- Data Protection --    (encryption, versioning, soft delete, retention)
      -- Lifecycle Rules --    (if any)
      -- Labels --             (if any)
```

**Multi-provider variant:**
```
synkronus storage list-buckets --providers gcp,aws (S1)
│   GCP results succeed, AWS returns "not yet implemented" errors
│   Output:
│     stderr: "Warning: some providers failed: provider aws: AWS ListBuckets is not yet implemented"
│     stdout: Table with only GCP buckets
```

---

### Flow 3: Manage Storage Objects

Browsing the contents of a specific bucket.

```
synkronus storage list-objects --provider gcp --bucket my-bucket (S5)
│   Output:
│     Listing objects in bucket: my-bucket
│     Table: KEY | SIZE | STORAGE CLASS | LAST MODIFIED
│     Directories shown as "(DIR)" entries
│
├─→ (Drill into a directory)
│   synkronus storage list-objects --provider gcp --bucket my-bucket --prefix logs/ (S5)
│   Output: Objects and subdirectories under logs/ prefix
│
└─→ (Inspect a specific object)
    synkronus storage describe-object logs/2024-01-15.csv --provider gcp --bucket my-bucket (S6)
    Output: Full detail view with sections:
      == Object: logs/2024-01-15.csv ==
      -- Overview --              (bucket, provider, size, class, dates, encryption, checksums)
      -- HTTP Headers --          (content-type, encoding, cache-control — if set)
      -- User-Defined Metadata -- (custom key-value pairs — if any)
```

---

### Flow 4: Create and Delete Storage Buckets

Mutating operations with safety mechanisms.

**Create:**
```
synkronus storage create-bucket new-bucket --provider gcp --location us-central1 (S3)
│   Output: "Bucket 'new-bucket' created successfully in us-central1 on provider gcp."
│
└─→ synkronus storage describe-bucket new-bucket --provider gcp (S2)
    Output: Detail view of newly created bucket
```

**Delete (interactive confirmation):**
```
synkronus storage delete-bucket old-bucket --provider gcp (S4)
│   Output:
│     WARNING: You are about to delete the bucket 'old-bucket' on provider 'GCP'.
│     This action CANNOT be undone and may result in permanent data loss.
│     To confirm, please type the name 'old-bucket': _
│
├─→ User types "old-bucket" (matches)
│   Output: "Bucket 'old-bucket' deleted successfully from provider gcp."
│
└─→ User types anything else or Ctrl+C
    Output: "Deletion aborted: Confirmation mismatch or cancelled."
    Exit code: non-zero (ErrOperationAborted)
```

**Delete (force, no confirmation):**
```
synkronus storage delete-bucket old-bucket --provider gcp --force (S4)
│   No prompt. Immediate deletion.
│   Output: "Bucket 'old-bucket' deleted successfully from provider gcp."
```

---

### Flow 5: List and Inspect SQL Instances

```
synkronus sql list (Q1)
│   Output: Table of all SQL instances across configured SQL providers
│   Columns: INSTANCE NAME | PROVIDER | REGION | VERSION | TIER | STATE | ADDRESS | CREATED
│
│   User identifies an instance
│
└─→ synkronus sql describe my-db --provider gcp (Q2)
    Output: Full detail view with sections:
      == SQL Instance: my-db ==
      -- Overview --    (provider, region, version, tier, state, address, storage, project, connection name, created)
      -- Labels --      (if any)
```

---

### Flow 6: Manage Configuration

```
synkronus config list (C4)
│   Output: All key-value pairs, sorted alphabetically
│   Or: "No configuration values set. Use 'synkronus config set <key> <value>'."
│
├─→ synkronus config set gcp.project new-project (C1)
│   Output: "Configuration set: gcp.project = new-project"
│
├─→ synkronus config get gcp.project (C2)
│   Output: "gcp.project = new-project"
│
└─→ synkronus config delete aws.region (C3)
    Output: "Configuration key 'aws.region' deleted"
```

---

## 4. Authentication States

Synkronus has no built-in authentication. It delegates entirely to cloud
SDK credential chains. The tool has three effective credential states:

### State 1: No Cloud Credentials

**How:** User has not run `gcloud auth application-default login` or
equivalent. No service account key is available.

**Behavior:** Config commands (C1–C4) work normally. Any storage or SQL
command fails at the provider initialization step with a cloud SDK
authentication error.

```
synkronus storage list-buckets --providers gcp
│   Error: error initializing provider: <GCP SDK auth error>
```

**Gap:** The error message is the raw SDK error, which can be cryptic.
No guidance is provided on how to authenticate. There is no
`synkronus auth` or `synkronus doctor` command to diagnose credential
issues.

### State 2: Cloud Credentials Valid, No Synkronus Config

**How:** User has valid cloud credentials but hasn't run `config set`.

**Behavior:** The `appContainer` initializes successfully (config loads
as empty). Commands that need a provider see no configured providers.

```
synkronus storage list-buckets
│   Output: "No providers configured. Use 'synkronus config set'.
│            Supported providers: aws, gcp"
```

This is handled correctly — actionable error message with next step.

### State 3: Cloud Credentials Valid, Config Set

**How:** User has both valid credentials and `~/.config/synkronus/config.json`
with at least one provider configured.

**Behavior:** All commands work as expected.

### State 4: Cloud Credentials Valid, Insufficient Permissions

**How:** Credentials are valid but the IAM role lacks required permissions
(e.g., `storage.buckets.list`, `storage.buckets.getIamPolicy`).

**Behavior:** Depends on the operation:
- `list-buckets`: Returns empty list or API error
- `describe-bucket`: Returns partial data. IAM policy section degrades
  gracefully with `(Could not retrieve IAM policy - check permissions)`
  (`storage_formatter.go:153`). This is the only permission-aware
  degradation in the codebase.
- All other operations: Return the raw API permission error.

### Auth Guards

| Command Group | Auth Required | Guard Mechanism |
|---------------|--------------|-----------------|
| `config *` | No | Config system is local-only, no cloud API calls |
| `storage *` | Yes (per provider) | Provider factory calls SDK, which checks credentials at client init |
| `sql *` | Yes (per provider) | Same as storage |
| Root / help | No | No cloud API calls |

**There are no explicit auth guards in Synkronus.** Authentication
failures surface as runtime errors from the cloud SDKs. This is standard
for CLI tools that wrap cloud APIs.

---

## 5. Broken or Incomplete Flows

### 5.1 AWS Storage — All Operations Dead-End

**Severity:** Critical

**Affected commands:** S1–S6 when `--provider aws` is specified

**What happens:** Every AWS storage method returns a "not yet implemented"
error. In multi-provider list operations, this surfaces as a partial
failure warning on stderr alongside working GCP results.

**Where it breaks:** `internal/provider/storage/aws/aws.go:52-76` — all interface
methods return `fmt.Errorf("AWS <Operation> is not yet implemented")`.

```
synkronus storage describe-bucket my-bucket --provider aws
│   Error: error describing bucket 'my-bucket' on aws:
│          error initializing provider: AWS DescribeBucket is not yet implemented
│
│   ✗ Dead end. No recovery path. No "coming soon" guidance.
```

### 5.2 AWS SQL — Silently Absent

**Severity:** High

**Affected commands:** Q1 when AWS is configured but `--providers` flag
is omitted

**What happens:** The SQL provider registry has no AWS entry. When a
user runs `sql list` without `--providers`, `ProviderResolver.Resolve()` calls
`GetConfiguredSqlProviders()` which only returns registered providers.
AWS is never mentioned.

**Where it breaks:** `internal/provider/factory/factory.go:34-36` —
`GetConfiguredSqlProviders()` only returns providers that are both
registered and configured. AWS SQL is not registered.

```
synkronus sql list
│   Output: (only GCP instances, no mention of AWS)
│   User has no indication AWS was silently excluded
│
│   ✗ Silent omission. No warning, no error.

synkronus sql list --providers gcp,aws
│   Error: unsupported SQL providers requested: [aws].
│          Supported SQL providers are: [gcp]
│
│   ✓ Explicit error when user requests AWS specifically.
│     But user must already know to ask.
```

### 5.3 No Object Upload/Download

**Severity:** Medium (functionality gap, not a broken flow)

**What happens:** Users can list and describe objects but cannot upload,
download, copy, or move them.

```
synkronus storage list-objects --provider gcp --bucket my-bucket
│   Output: Objects listed successfully
│
│   User wants to download an object
│
│   ✗ No command exists. Must fall back to gcloud/aws CLI.
```

### 5.4 No Bucket Update Operations

**Severity:** Medium (functionality gap)

**What happens:** Users can create and delete buckets but cannot modify
bucket settings (labels, lifecycle rules, versioning, etc.).

```
synkronus storage describe-bucket my-bucket --provider gcp
│   Output: Shows current lifecycle rules, labels, etc.
│
│   User wants to add a label or change versioning
│
│   ✗ No update command exists.
```

### 5.5 Config Delete Validation Gap

**Severity:** Low

**Affected command:** C3

**What happens:** Deleting the only configured field for a provider
(e.g., `gcp.project`) succeeds but leaves behind an empty `"gcp": {}`
structure in the config file. The validator uses `validate:"omitempty"`
on the provider struct, so an empty struct passes validation. The
provider appears unconfigured (config check fails because `Project` is
empty), which is correct behavior. But the empty key persists in the
file.

**Where:** `internal/config/config.go:129-153` — `DeleteValue` sets
value to `""` rather than removing the key entirely.

---

## 6. Edge Case Coverage

### Config Commands

| Command | Empty State | Error State | Permission Denied |
|---------|------------|-------------|-------------------|
| C1 `config set` | N/A | Handled: validation error if unknown key or invalid value (`config_cmd.go:34`) | Handled: filesystem write error propagated (`config.go:92-94`) |
| C2 `config get` | Handled: `"configuration key 'x' not found or not set"` (`config_cmd.go:57`) | Same as empty | N/A (local file) |
| C3 `config delete` | Handled: `"configuration key 'x' not found"` (`config_cmd.go:83`) | Handled: validation errors on delete (`config.go:143-145`) | N/A (local file) |
| C4 `config list` | Handled: `"No configuration values set."` with actionable hint (`config_cmd.go:114`) | N/A | N/A (local file) |

### Storage Commands

| Command | Empty State | Loading State | Error State | Permission Denied |
|---------|------------|---------------|-------------|-------------------|
| S1 `list-buckets` | Handled: `"No buckets found."` or `"No providers configured."` with hint | **Not handled:** No spinner or progress indicator during API calls | Handled: Partial failures print warning to stderr, complete failures return error | Not specifically handled; raw API error shown |
| S2 `describe-bucket` | N/A (requires specific bucket name) | Not handled | Handled: error wrapped with context (`storage_cmd.go:101`) | **Partially handled:** IAM section degrades gracefully; other sections show raw API error |
| S3 `create-bucket` | N/A | Not handled | Handled: error wrapped with context (`storage_cmd.go:127`) | Raw API error |
| S4 `delete-bucket` | N/A | Not handled | Handled: error wrapped with context (`storage_cmd.go:170`) | Raw API error |
| S5 `list-objects` | Handled: `"No objects or directories found."` (`storage_formatter.go:326`) | Not handled | Handled: error wrapped with context (`storage_cmd.go:200`) | Raw API error |
| S6 `describe-object` | N/A | Not handled | Handled: error wrapped with context (`storage_cmd.go:230`) | Raw API error |

### SQL Commands

| Command | Empty State | Loading State | Error State | Permission Denied |
|---------|------------|---------------|-------------|-------------------|
| Q1 `list` | Handled: `"No SQL instances found."` or `"No SQL providers configured."` with hint | Not handled | Handled: partial failure warning on stderr | Raw API error |
| Q2 `describe` | N/A | Not handled | Handled: error wrapped with context (`sql_cmd.go:93`) | Raw API error |

**Summary of gaps:**
- **Loading state:** No command has a loading indicator. All block silently until results arrive.
- **Permission denied:** Only `describe-bucket`'s IAM section handles this gracefully. All other commands expose raw cloud SDK permission errors.
- **Empty states:** All list commands handle empty results with clear messages. Detail commands require a valid resource name so "empty" doesn't apply.

---

## 7. Navigation Structure

### Command Tree

```
synkronus
├── storage
│   ├── list-buckets    (alias: list)
│   ├── describe-bucket (alias: describe)
│   ├── create-bucket   (alias: create)
│   ├── delete-bucket   (alias: delete)
│   ├── list-objects
│   └── describe-object
├── sql
│   ├── list            (alias: list-instances)
│   └── describe        (alias: describe-instance)
└── config
    ├── set
    ├── get
    ├── delete
    └── list
```

### Navigation Characteristics

| Aspect | Implementation |
|--------|---------------|
| **Discovery** | `--help` on any command or group shows available subcommands and flags |
| **Back/up** | N/A — each command is a standalone invocation. No session state. |
| **Breadcrumbs** | N/A — command path is explicit in the invocation (`synkronus storage list-buckets`) |
| **Modal/dialog** | Delete confirmation prompt (`internal/ui/prompt/prompt.go`). Only interactive element in the tool. Bypassed with `--force`. |
| **Tab completion** | **Not implemented.** Cobra supports generating completion scripts but this hasn't been wired up. |
| **Command aliases** | Exist but not documented in help text. Users discover them only from source code or if alias happens to match their guess. |

### Interactive vs Non-Interactive

| Mode | Commands | Behavior |
|------|----------|----------|
| **Non-interactive** (default) | All commands except S4 without `--force` | Execute, print output, exit |
| **Interactive** | S4 `delete-bucket` (without `--force`) | Prints warning, waits for user input, validates, then executes or aborts |
| **Pipe-safe** | All commands | stdout contains only data output (tables, messages). stderr contains errors and warnings. Debug logging goes to stderr. |

---

## 8. Platform and Terminal Differences

### Terminal Width

**Not handled.** No terminal width detection exists anywhere in the
codebase. The table formatter (`internal/output/table.go`) auto-sizes
columns based on content width with no upper bound.

| Table | Columns | Approximate Minimum Width |
|-------|---------|--------------------------|
| Bucket list | 6 | ~100 characters |
| Instance list | 8 | ~120 characters |
| Object list | 4 | ~80 characters |
| Key-value detail | 2 | ~60 characters |

Tables with long values (bucket names, object keys, IAM roles) can
exceed 200 characters wide. In terminals narrower than the table, rows
wrap mid-cell, breaking alignment.

### Platform-Specific Behavior

| Aspect | macOS/Linux | Windows |
|--------|-------------|---------|
| Config path | `~/.config/synkronus/config.json` | `%USERPROFILE%\.config\synkronus\config.json` (Go's `os.UserHomeDir`) |
| File permissions | `0700`/`0600` enforced | Permissions are set but may not be enforced by Windows filesystem |
| Pipe detection | Works (`os.Stdout.Fd()` if implemented) | Works but no color is used so irrelevant currently |
| Terminal encoding | UTF-8 standard | May need console code page consideration for non-ASCII bucket names |

### No Responsive Differences

There is no responsive behavior. The output is identical regardless of
terminal dimensions. The only adaptation point is the table column
auto-sizing, which expands to fit content but never shrinks or truncates.

---

## 9. Unreachable Code and Orphaned Components

### Unreachable Commands

**None.** All 12 commands (excluding help variants) are registered in
the Cobra command tree and reachable through normal invocation.

### Orphaned Code

| Item | Location | Status |
|------|----------|--------|
| `common.AWS` constant | `internal/domain/provider.go` | Defined but never returned by any working provider's `ProviderName()`. Only used by the AWS storage stub. Will become active when AWS is implemented. |
| AWS logging prefix `"s3://"` | `internal/output/storage_views.go` | Formatter contains AWS-specific logic for bucket logging URIs (`s3://...`), but the AWS provider never returns logging data. Dead code until AWS is implemented. |
| Stashed billing branch | Git stash (not in main) | Abandoned billing/cost estimation work. Not in the codebase but present in git history. Should be cleaned up. |
| Built binaries | `cmd/synkronus/synkronus`, `synkronus` (root) | Compiled binaries tracked in git status. **RESOLVED** — Added to `.gitignore`, removed from tracking. |

### Alias Documentation Gap

Command aliases exist but are invisible to users:

| Primary Command | Hidden Alias | Discoverable? |
|-----------------|-------------|---------------|
| `storage list-buckets` | `storage list` | Only via `--help` (shown as `Aliases: list`) |
| `storage describe-bucket` | `storage describe` | Same |
| `storage create-bucket` | `storage create` | Same |
| `storage delete-bucket` | `storage delete` | Same |
| `sql list` | `sql list-instances` | Same |
| `sql describe` | `sql describe-instance` | Same |

Cobra does show aliases in `--help` output, so they are technically
discoverable but not prominently surfaced.
