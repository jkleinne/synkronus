# Synkronus — Product Requirements Document

> **Version:** 1.0 — 2 April 2026
> **Status:** Draft
> **Author:** Auto-generated from codebase analysis

---

## 1. Executive Summary

Synkronus is a multi-cloud infrastructure CLI tool that provides a single, consistent command surface for managing storage buckets, objects, and SQL database instances across cloud providers. It targets the gap between full infrastructure-as-code platforms (Terraform, Pulumi) — which require declarative configuration and state management — and individual cloud CLIs (gcloud, aws), which force users to context-switch between incompatible tools. The tool is built in Go using Cobra and currently ships a fully functional GCP implementation with a stubbed AWS placeholder.

The codebase is architecturally sound: cleanly layered, extensible via self-registering providers, and equipped with concurrent multi-provider fanout with graceful partial failure handling. At ~3,700 lines of Go across 41 files, the core is compact and well-structured. However, the product is not yet launch-ready. AWS support — a central part of the multi-cloud value proposition — is entirely non-functional. There is no CI/CD pipeline, no integration or end-to-end tests, no documentation beyond the developer-facing CLAUDE.md, and no Azure support. The highest-priority actions are: (1) implement AWS storage and SQL operations to deliver on the multi-cloud promise, (2) add CI/CD and expand test coverage, and (3) produce user-facing documentation including installation instructions, quickstart, and command reference.

---

## 2. Problem Statement and Target Audience

### Problem

Engineering teams operating across multiple cloud providers face a fragmented tooling experience. Each cloud has its own CLI with different authentication models, flag conventions, output formats, and command structures. For routine operational tasks — listing buckets, inspecting database instances, creating or deleting resources — engineers must remember provider-specific syntax and mentally translate between tools. Infrastructure-as-code platforms solve provisioning but are too heavyweight for ad-hoc operational queries. There is no simple, imperative, unified CLI for quick cross-cloud resource management.

### Target Audience

**Primary:** DevOps engineers and platform teams at organizations using two or more cloud providers who need fast, cross-provider visibility and basic resource management without IaC overhead.

**Secondary:** Individual developers working across clouds who want a single tool instead of juggling multiple CLIs.

### User Personas

**Persona 1 — Rahul, 34, Senior Platform Engineer**
Rahul maintains infrastructure across GCP and AWS for a mid-size SaaS company. He spends 20% of his time on ad-hoc tasks: checking which buckets exist across providers, inspecting SQL instance configurations during incident response, and cleaning up orphaned resources. He knows both gcloud and aws CLI but finds the context-switching costly — different flag names, different output formats, different pagination behavior. He wants a single command that lists all buckets everywhere, and a consistent `describe` experience that works the same way regardless of provider.

**Persona 2 — Keiko, 28, Full-Stack Developer**
Keiko builds features that use GCP Cloud SQL and AWS S3. She isn't a cloud infrastructure specialist and finds the native CLIs overwhelming — gcloud has hundreds of subcommands, and the AWS CLI documentation is dense. She wants a focused tool with a small command surface where `synkronus storage list` just works across her configured providers, and `synkronus sql describe my-instance` gives her what she needs without memorizing provider-specific incantations.

---

## 3. Competitive Landscape

| Tool | Type | Scope | Multi-Cloud | Learning Curve | Resource Mutation |
|------|------|-------|-------------|----------------|-------------------|
| **gcloud / aws / az** | Native CLI | Full cloud surface | Single provider each | High (hundreds of commands) | Full |
| **Terraform** | IaC | Full infrastructure | Yes (via providers) | High (HCL, state management) | Declarative only |
| **Pulumi** | IaC | Full infrastructure | Yes (via SDKs) | High (language SDKs, state) | Declarative only |
| **Steampipe** | Query engine | Read-only inventory | Yes (via plugins) | Medium (SQL syntax) | Read-only |
| **Cloudquery** | Data sync | Read-only analytics | Yes (via plugins) | Medium (ETL pipeline) | Read-only |
| **cloud-nuke** | Cleanup | Resource deletion | AWS only | Low | Delete only |
| **Synkronus** | Operational CLI | Storage + SQL | Yes (GCP + AWS planned) | Low (12 commands) | Read + Create + Delete |

### Differentiation

1. **Imperative and instant.** No configuration files, no state management, no plan/apply cycle. Configure credentials once, run commands immediately.
2. **Narrow scope, fast learning.** 12 commands covering storage and SQL — learnable in minutes, not days. Deliberately not trying to manage every cloud service.
3. **Unified output.** Consistent table formatting, flag conventions (`--provider`/`--providers`), and error messages across all clouds.
4. **Concurrent cross-provider queries.** `synkronus storage list-buckets` fans out to all configured providers simultaneously and returns combined results with partial failure tolerance — something no native CLI offers.
5. **Extensible architecture.** Adding a new provider requires implementing an interface and a single `init()` registration — no framework changes needed.

### Competitive Risks

- Native CLIs are always available and always up-to-date. Synkronus must deliver enough convenience to justify installing a separate tool.
- Steampipe's SQL-based querying is more flexible for complex inventory queries. Synkronus competes on simplicity and mutation support, not query power.
- If Terraform or Pulumi add imperative command modes (Terraform already has `import`), the niche narrows.

---

## 4. Feature Inventory

### 4.1 Functional (Working as Intended)

| Feature | Scope | Notes |
|---------|-------|-------|
| **GCP Storage — List Buckets** | Full | Includes usage metrics from Cloud Monitoring API |
| **GCP Storage — Describe Bucket** | Full | IAM policies, ACLs, lifecycle rules, encryption, retention, versioning |
| **GCP Storage — Create Bucket** | Full | Name + location |
| **GCP Storage — Delete Bucket** | Full | Confirmation prompt (type bucket name to confirm) |
| **GCP Storage — List Objects** | Full | Prefix filtering, directory simulation via delimiters |
| **GCP Storage — Describe Object** | Full | Metadata, checksums, encryption, HTTP headers |
| **GCP SQL — List Instances** | Full | Name, region, version, tier, state, address |
| **GCP SQL — Describe Instance** | Full | Complete metadata including labels |
| **Multi-provider fanout** | Full | Concurrent queries with partial failure tolerance |
| **Configuration management** | Full | Set/get/delete/list with strict validation, secure file permissions |
| **Debug logging** | Full | `--debug` flag enables verbose slog output to stderr |
| **Output formatting** | Full | ASCII tables (default), JSON, and YAML via --output flag |
| **Destructive operation safety** | Full | Value-matching confirmation (not yes/no) for deletes |
| **Provider auto-registration** | Full | `init()`-based plugin system, generic thread-safe registry |

### 4.2 Incomplete or Non-Functional

| Feature | Status | Details |
|---------|--------|---------|
| **AWS Storage — All Operations** | Stubbed | All 6 interface methods return "not yet implemented" errors. Registered in provider registry. Config check implemented. |
| **AWS SQL** | Absent | No implementation, no registration, no config. A user configuring `aws.region` would see storage stub errors but SQL commands would silently skip AWS entirely. |
| **IAM Conditional Bindings** | Partial | Detected and logged at debug level but not displayed in output. TODO comment in `pkg/storage/gcp/buckets.go:137-140`. |
| **Cost Estimation** | Abandoned | Git history shows a `feat/storage-describe-cost-estimation` branch with billing client work (commits `6cdca3f`, `1b0f2c3`, `7e62ff8`, `0e7e794`, `1b77f47`) that was stashed and never merged. Billing interface and formatter updates exist only in stash. |

### 4.3 Not Yet Built

| Feature | Priority | Rationale |
|---------|----------|-----------|
| **AWS Storage implementation** | Critical | Core value proposition is multi-cloud |
| **AWS SQL implementation** | Critical | Same as above |
| **Azure provider** | High | Third major cloud; significantly widens addressable market |
| **CI/CD pipeline** | ~~Critical~~ | ~~No automated build, test, or release pipeline~~ **DONE** — GitHub Actions CI (build, vet, test) |
| **User documentation** | Critical | No README, quickstart, installation guide, or command reference |
| **Integration tests** | High | Only unit tests exist; no tests against real or emulated cloud APIs |
| **Shell completions** | Medium | Cobra supports generating bash/zsh/fish completions |
| **Output format options** | ~~Medium~~ | ~~Only ASCII table output; no JSON/YAML/CSV for scripting~~ **DONE** — `--output table\|json\|yaml` implemented |
| **Pagination** | Medium | Large bucket/object lists will produce unwieldy output |
| **Authentication guidance** | High | No documentation on how to authenticate with each provider |
| **Version command** | Low | No `synkronus version` or build metadata |
| **Update/modify operations** | Medium | Can create/delete buckets but not update bucket settings |
| **Object upload/download** | Medium | Can list/describe objects but not transfer them |

---

## 5. Gap Analysis

### 5.1 Critical Gaps (Launch Blockers)

| Gap | Impact | Evidence |
|-----|--------|----------|
| **AWS storage is non-functional** | Undermines the entire multi-cloud value proposition. A user who configures AWS will get "not yet implemented" errors for every operation. | `pkg/storage/aws/aws.go` — all methods return errors |
| **AWS SQL does not exist** | SQL list command silently ignores AWS even if configured. No error, no warning — just missing results. | No AWS entry in `internal/provider/registry` SQL registry |
| **No user documentation** | Users have no way to learn how to install, configure, or use the tool. No README in the repository root. | `docs/` directory is empty |
| ~~**No CI/CD**~~ | ~~No automated testing...~~ **RESOLVED** — GitHub Actions CI in `.github/workflows/ci.yml` | `.github/workflows/ci.yml` |
| **No installation mechanism** | No published binaries, no Homebrew formula, no Docker image. Users must clone and `go build`. | No release artifacts or package manifests |

### 5.2 Integrity Gaps (Misleading or Deceptive Behavior)

| Gap | Impact | Evidence |
|-----|--------|----------|
| **AWS stub errors lack guidance** | "AWS ListBuckets is not yet implemented" tells users what failed but not what to do about it. No indication this is a known limitation. | `pkg/storage/aws/aws.go:52-76` |
| **Silent AWS SQL omission** | When a user runs `synkronus sql list --providers gcp,aws`, AWS is silently skipped because no SQL provider is registered. The user sees only GCP results with no indication AWS was ignored. | `internal/provider/registry/registry.go` — `GetAll` only returns registered providers |
| **Config accepts `aws.region` without functional benefit** | Users can set AWS configuration, creating the expectation that AWS works. The config system validates and persists the value successfully. | `internal/config/config.go` — AWS config struct is fully defined |

### 5.3 Quality Gaps (Should Fix Before or Shortly After Launch)

| Gap | Impact |
|-----|--------|
| ~~**No JSON/YAML output mode**~~ | ~~CLI tools used in automation pipelines need machine-readable output. ASCII tables cannot be parsed reliably.~~ **RESOLVED** — `--output json\|yaml` flag on all data commands |
| **No integration tests** | 4 test files with ~328 lines cover only unit-level logic. No coverage of actual cloud API interactions, CLI argument parsing, or end-to-end workflows. |
| **No shell completions** | Tab completion is expected in modern CLIs. Cobra supports this natively but it hasn't been wired up. |
| **No pagination for large result sets** | Listing thousands of objects will produce unmanageable output with no way to page through results. |
| **Conditional IAM bindings not displayed** | GCP V3 IAM policies with conditions are silently skipped. Users may see incomplete policy views without realizing it. |
| **No help text beyond Cobra defaults** | Commands have short descriptions but no examples, extended help, or usage patterns. |

---

## 6. Recommended MVP Scope

### Phase 1: Ship (Launch Blockers)

| Item | Effort Estimate | Notes |
|------|-----------------|-------|
| **Implement AWS S3 storage operations** | Large | ListBuckets, DescribeBucket, CreateBucket, DeleteBucket, ListObjects, DescribeObject using AWS SDK for Go v2 |
| **Implement AWS RDS SQL operations** | Medium | ListInstances, DescribeInstance using AWS SDK for Go v2 |
| **Write user-facing documentation** | Medium | README with overview, installation, quickstart, configuration guide, command reference, authentication setup per provider |
| **Set up CI/CD** | Medium | GitHub Actions: lint, vet, test on PR; build + release binaries on tag (GoReleaser) **DONE** |
| **Add `--output` flag (table/json)** | Medium | JSON output for scripting; table remains default **DONE** |
| **Add `synkronus version`** | Small | Build-time injection of version, commit, date via ldflags |

### Phase 2: Harden (First Week Post-Launch)

| Item | Notes |
|------|-------|
| **Integration test suite** | Test against real or emulated cloud APIs (e.g., fake-gcs-server, LocalStack) |
| **Shell completion generation** | `synkronus completion bash/zsh/fish` |
| **Improve AWS stub error messages** | Until AWS ships: "AWS support is planned. Track progress at [link]." with `--provider` suggestions |
| **Surface conditional IAM bindings** | Display condition title/expression instead of silently skipping |
| **Add command examples to help text** | Cobra `Example` field for every command |
| **Authentication troubleshooting docs** | Common errors, required permissions, credential setup per provider |

### Defer (Valuable but Not Launch-Critical)

| Item | Notes |
|------|-------|
| **Azure provider** | Third major cloud; significant market expansion but not needed for an initial GCP+AWS launch |
| **Object upload/download** | Useful but overlaps with native CLIs; list/describe is the differentiator |
| **Bucket update operations** | Update labels, lifecycle rules, etc. |
| **Cost estimation** | Previously attempted (see git stash history); valuable but complex — billing APIs vary significantly across providers |
| **Pagination / streaming output** | Important at scale but most users won't hit limits initially |
| **Interactive mode / TUI** | Could improve discovery but adds significant complexity |

### Cut (Remove or Rethink)

| Item | Rationale |
|------|-----------|
| **Cost estimation (current approach)** | The stashed billing integration (`feat/storage-describe-cost-estimation` branch) was abandoned. Billing APIs differ radically between providers, making a unified abstraction fragile. Reassess only after AWS parity is achieved. If revisited, consider a simpler approach: link to each provider's pricing calculator rather than computing estimates. |
| **AWS storage stub in production builds** | Until real AWS support ships, the stub creates a misleading experience. Consider either: (a) gating the stub behind a `--experimental` flag, or (b) replacing stub errors with explicit "coming soon" messages that don't look like runtime failures. |

---

## 7. Success Metrics

| Metric | Measurement Method | Target | Justification |
|--------|-------------------|--------|---------------|
| **Provider parity** | Automated test matrix: % of interface methods with passing tests per provider | 100% for GCP and AWS at launch | A multi-cloud tool with one working provider is a single-cloud tool with extra complexity |
| **Test coverage** | `go test -coverprofile` across all packages | ≥ 70% line coverage | Industry baseline for Go CLI tools; the current ~15% (estimated from 328 test lines / ~3,700 code lines) is far below |
| **Build reliability** | CI pass rate on main branch over 30-day rolling window | ≥ 95% | Standard threshold for healthy CI pipelines |
| **Installation-to-first-command time** | Manual benchmarking: clone → configure → first successful `list-buckets` | Under 5 minutes with documentation | Competitive with native CLIs; exceeding this signals documentation gaps |
| **GitHub stars** | GitHub Insights | 100 in first 3 months | Comparable open-source Go CLIs (e.g., cloud-nuke, steampipe plugins) reach this range organically with good documentation and a clear README |
| **Issue response time** | GitHub Issues median first-response time | < 48 hours | Signals active maintenance; critical for open-source adoption trust |
| **Command execution latency** | Benchmarking harness: time for `list-buckets` across 2 providers | Under 5 seconds for typical usage (< 50 buckets per provider) | Concurrent fanout should keep cross-provider latency near single-provider latency |

---

## 8. Key Assumptions and Risks

### Assumptions

| Assumption | Impact if Wrong |
|------------|----------------|
| **Multi-cloud is common enough to sustain a dedicated tool.** Most organizations use 2+ clouds. | If most teams are single-cloud, the unified CLI value proposition collapses. The tool becomes an inferior wrapper around a single native CLI. **Mitigation:** Validate with user research; even single-cloud users benefit from simpler commands if the tool is good enough. |
| **Storage and SQL are sufficient initial scope.** These are the most universally used cloud services. | If users expect compute, networking, or IAM management, the tool may feel too limited to justify adoption. **Mitigation:** The provider registry architecture supports adding service types without restructuring. Scope can expand based on demand. |
| **Users will authenticate via existing cloud SDK mechanisms** (ADC for GCP, credential chain for AWS). | If authentication is painful, users won't get past setup. **Mitigation:** Clear documentation and `synkronus config` wizard can guide users. |
| **ASCII table output is acceptable as default.** | If most users are in automation/scripting contexts, the lack of JSON output is a blocker even at launch. **Mitigation:** JSON output is in Phase 1 scope. |

### Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| **AWS SDK behavior differs significantly from GCP SDK** — e.g., pagination models, error types, naming conventions — making the unified interface leaky. | High | High | Design the `Storage` and `SQL` interfaces to be provider-agnostic. Accept that `describe` output may include provider-specific fields (already handled via formatter layer). |
| **GCP SDK breaking changes** — Google's Go client libraries are pre-1.0 for some services and change frequently. | Medium | Medium | Pin dependency versions in `go.mod`. Set up Dependabot or Renovate for controlled updates. |
| **Scope creep toward IaC.** Users may request features (drift detection, desired-state enforcement, plan/apply) that push Synkronus toward Terraform's domain. | Medium | High | Maintain a clear product boundary: Synkronus is an operational CLI, not an IaC tool. Document this boundary explicitly. Decline features that require state management. |
| **Single maintainer risk.** The git history suggests a solo developer. Bus factor of 1. | High | High | Prioritize documentation, CI/CD, and contributor guidelines to lower the barrier for outside contribution. |
| **Performance degradation with scale.** The monitoring API call in GCP's ListBuckets adds latency. No pagination means large result sets load entirely into memory. | Medium | Medium | Add pagination support in Phase 2. Consider caching or optional metrics retrieval (`--no-metrics` flag). |
| **Security exposure from credential handling.** The tool relies on ambient cloud credentials. Misconfiguration could expose sensitive infrastructure. | Low | High | Document minimum-privilege IAM roles per provider. Never store credentials in Synkronus config — rely on cloud SDK credential chains. Add a `--dry-run` flag for destructive operations. |

---

## Appendix: Contradictions and Observations

### A.1 — AWS Stub Registers as a Valid Provider

**Observation:** The AWS storage provider registers itself in the provider registry (`pkg/storage/aws/aws.go:14-18`) and passes configuration checks if `aws.region` is set. From the system's perspective, AWS is a supported, configured provider. From the user's perspective, every AWS operation fails.

**Evidence:**
- `pkg/storage/aws/aws.go:14-18` — `init()` registers AWS with `registry.RegisterProvider`
- `pkg/storage/aws/aws.go:52-76` — All methods return "not yet implemented" errors
- `internal/config/config.go` — `AWSConfig` struct accepts and validates `aws.region`

**Impact:** A user who configures `aws.region` and runs `synkronus storage list-buckets` will see AWS errors mixed with GCP results. The multi-provider fanout correctly surfaces this as a partial failure, but the error message ("AWS ListBuckets is not yet implemented") looks like a bug, not a known limitation.

**Recommendation:** Either remove AWS from the registry until implemented, or replace error messages with explicit "coming soon" notices that include a link to track progress.

### A.2 — AWS SQL Provider Is Completely Absent

**Observation:** While AWS storage has a stub implementation, AWS SQL has nothing — no code, no registration, no tests.

**Evidence:**
- `internal/provider/providers.go` — Only imports `_ "synkronus/pkg/sql/gcp"` (no AWS SQL import)
- `pkg/sql/` — Contains only `gcp/` subdirectory
- `internal/provider/registry/registry.go` — SQL registry has no AWS entry

**Impact:** `synkronus sql list --providers gcp,aws` silently returns only GCP results. There is no error, no warning, and no indication that AWS was requested but unavailable. This is worse than the storage stub, which at least surfaces an error.

### A.3 — Abandoned Cost Estimation Branch

**Observation:** Git history contains a `feat/storage-describe-cost-estimation` branch with substantial work (billing interface, formatter updates, GCP billing client) that was stashed and never merged.

**Evidence:**
- Commits `6cdca3f`, `1b0f2c3`, `7e62ff8`, `0e7e794`, `1b77f47` in git log
- Stash entry `358b973` — "Auto-stashing changes for checking out main"

**Impact:** None currently (code is not in `main`). The abandoned branch suggests billing API integration was attempted and found to be complex enough to defer. This informs the recommendation to cut cost estimation from near-term scope.

### A.4 — Binary Artifacts in Repository

**Observation:** Two binary files appear in git status as modified/untracked.

**Evidence:**
- `cmd/synkronus/synkronus` — Modified binary in source tree
- `synkronus` — Untracked binary at repository root

**Impact:** Binaries in source control increase repository size and can cause merge conflicts. These should be added to `.gitignore`.

**RESOLVED** — `.gitignore` added, binaries removed from git tracking.
