# Synkronus — CLI Output Style Guide

> **Version:** 1.0 — 2 April 2026
> **Scope:** Terminal output formatting for a Go CLI tool
>
> Synkronus is a terminal application, not a web application. This style
> guide covers the actual UI surface: ASCII-formatted terminal output.
> Sections on CSS, Tailwind, font families, responsive breakpoints,
> Lighthouse scores, ARIA attributes, and hex color palettes are **not
> applicable** and are omitted.

---

## 1. Color

### 1.1 Current State

**No ANSI colors are used anywhere in the codebase.** All output is plain
uncolored text. There are no escape sequences, no color library imports,
and no terminal capability detection.

### 1.2 Recommendation

Add color selectively for three categories using a library like
[`fatih/color`](https://github.com/fatih/color) or
[`muesli/termenv`](https://github.com/muesli/termenv). Both auto-detect
terminal capabilities and degrade gracefully when piped to a file or a
terminal that doesn't support color.

| Category | Where Used | Color | Rationale |
|----------|-----------|-------|-----------|
| **Error** | `Error: ...` on stderr | Red | Universal convention for errors |
| **Warning** | `Warning: some providers failed: ...` on stderr | Yellow | Distinguishes partial failures from fatal errors |
| **Success** | `Bucket 'x' created successfully...` | Green | Confirms completion of mutating operations |
| **Section headers** | `== Bucket: my-bucket ==` and `-- Overview --` | Bold (no color) | Adds visual hierarchy without color dependency |
| **Table headers** | `BUCKET NAME`, `PROVIDER`, etc. | Bold (no color) | Differentiates headers from data rows |
| **Info labels** | `(DIR)`, `N/A`, `Not configured` | Dim / gray | De-emphasizes placeholder values |

**Rules:**
- Always respect `NO_COLOR` environment variable (see https://no-color.org/)
- Disable color when stdout is not a TTY (piped to file or another program)
- Never convey meaning through color alone — the plain-text output must be fully understandable without color
- Dark/light mode: Not applicable. Terminal color schemes are user-controlled; the above ANSI colors render correctly in both

---

## 2. Typography

### 2.1 Current State

All output is monospaced terminal text. There are no font choices —
the user's terminal font applies universally. The only typographic
variation comes from:

- **UPPERCASE** headers in list tables (`BUCKET NAME`, `PROVIDER`, etc.)
- **Title Case** headers in detail key-value tables (`Parameter`, `Value`)
- **ASCII decoration** for section separators (`===` and `-- --`)

### 2.2 Inconsistencies Found

| Location | Pattern | Example |
|----------|---------|---------|
| List table headers | SCREAMING_CAPS | `BUCKET NAME`, `STORAGE CLASS` |
| Detail table headers | Title Case | `Parameter`, `Value` |
| Section titles | Sentence case | `-- Access Control & Logging --` |
| Page headers | Sentence case | `== Bucket: my-bucket ==` |
| IAM sub-header | Sentence case, no decorator | `Identity and Access Management (IAM) Policy:` |
| ACL sub-header | Sentence case, no decorator | `Access Control List (ACLs) - (Fine-grained object control):` |

### 2.3 Canonical Rules

| Element | Casing | Decorator | Example |
|---------|--------|-----------|---------|
| **Page header** | Title case | `=` border (full width) | `=== Bucket: my-bucket ===` |
| **Section title** | Title Case | `-- ... --` | `-- Data Protection --` |
| **Sub-section header** | Sentence case | Trailing `:` | `IAM Policy:` |
| **List table columns** | UPPER CASE | None | `NAME`, `PROVIDER` |
| **Detail table columns** | Title Case | None | `Parameter`, `Value` |
| **Inline context** | Sentence case | None | `Listing objects in bucket: x` |
| **Empty state** | Sentence case | None | `No buckets found.` |
| **Success message** | Sentence case | None | `Bucket 'x' created successfully.` |
| **Warning prefix** | `Warning:` | None | `Warning: some providers failed: ...` |
| **Error prefix** | `Error:` | None | `Error: unsupported provider: xyz` |

---

## 3. Spacing and Layout

### 3.1 Current Patterns

The output uses three distinct layout components, each with its own
spacing convention:

#### Tables (list views)

```
+-------------+----------+----------+
| BUCKET NAME | PROVIDER | LOCATION |
+-------------+----------+----------+
| my-bucket   | GCP      | US       |
+-------------+----------+----------+
```

- Cell padding: 1 space on each side of content (`| content |`)
- Column width: auto-calculated from widest value in each column
- Borders: `+` at intersections, `-` for horizontal, `|` for vertical
- No row separators between data rows (only header separator)

#### Detail views (describe commands)

```
======================================
  Bucket: my-bucket
======================================

-- Overview --
+---------------------+--------------------+
| Parameter           | Value              |
+---------------------+--------------------+
| Provider            | GCP                |
| Location            | US                 |
+---------------------+--------------------+


-- Access Control & Logging --
...
```

- Page header: `=` border, 30 chars longer than title, 2-space indent on title
- Section gap: `\n\n` (one blank line) between sections
- Section title followed by `\n` then table
- Sub-section headers are inline text followed by `:\n`

#### Inline messages (create, delete, config)

```
Bucket 'my-bucket' created successfully in US on provider gcp.
```

- Single line, no decoration
- Resource names quoted with single quotes
- Provider name as-is (lowercase from internal representation)

### 3.2 Inconsistencies Found

| # | Issue | Location | Details |
|---|-------|----------|---------|
| 1 | **Page header padding is content-dependent** | `internal/output/table.go:105` | `FormatHeaderSection` adds 30 characters of padding to the `=` border regardless of title length. Short titles get disproportionately wide borders. |
| 2 | **Inconsistent inter-section spacing** | `internal/output/storage_views.go` | Most sections end with `\n\n`, but the IAM policy section adds an extra `\n` after the conditional-bindings note (line 183), creating triple spacing in some cases. |
| 3 | **Object list has no trailing newline** | `internal/output/storage_views.go:352` | `FormatObjectList` returns the table string without a trailing newline, unlike all other format methods which end with `\n\n`. |
| 4 | **Context line before object list** | `internal/output/storage_views.go:318-321` | Object list has a `Listing objects in bucket: x` prefix. No other list command has a context prefix. Bucket list and instance list render only the table. |
| 5 | **Provider casing in inline messages** | `storage_create_bucket.go:130`, `storage_delete_bucket.go:173` | Success messages show provider as user-typed (lowercase), but the `WARNING` in delete confirmation is uppercase (`strings.ToUpper`). |

### 3.3 Canonical Spacing Rules

| Context | Spacing |
|---------|---------|
| After page header (`===`) | `\n\n` (one blank line) |
| Before section title (`-- --`) | None (start of section) |
| After section title, before table | `\n` (newline, no blank line) |
| After section table | `\n\n` (one blank line) |
| Between inline text and table | `\n` (newline, no blank line) |
| After sub-section header (text with `:`) | `\n` (newline, no blank line) |
| Between unrelated inline messages | `\n` (newline, no blank line) |

---

## 4. Component Patterns

### 4.1 Bordered Table

The primary output component. Used for both list views and key-value
detail sections.

**Structure:**
```
+-----------+-----------+
| Header 1  | Header 2  |      ← top border + header row
+-----------+-----------+      ← header separator
| Cell 1    | Cell 2    |      ← data rows
| Cell 3    | Cell 4    |
+-----------+-----------+      ← bottom border
```

**Implementation:** `internal/output/table.go` — `Table` struct with
`NewTable`, `AddRow`, `String` methods.

**Variants in use:**

| Variant | Headers | Usage | Files |
|---------|---------|-------|-------|
| **List table** | UPPER CASE, multi-column | Bucket list, instance list, object list | `storage_views.go` (BucketListView), `sql_views.go` (InstanceListView), `storage_views.go` (ObjectListView) |
| **Key-value table** | `Parameter` / `Value` (or similar) | Detail overview, data protection, config summary | `storage_views.go` (BucketDetailView.renderOverview), `sql_views.go` (InstanceDetailView.renderOverview) |
| **Entity table** | Domain-specific headers | IAM bindings (`Role`, `Principal(s)`), ACLs (`Entity`, `Role`), lifecycle rules | `storage_views.go` |

**Recommendation:** These are not separate components — they are all
the same `Table` with different header casing conventions. Keep the
single `Table` type. Standardize: list tables use UPPER CASE headers;
detail tables use Title Case.

### 4.2 Page Header

Used at the top of describe/detail views.

**Structure:**
```
======================================
  Bucket: my-bucket
======================================
```

**Implementation:** `internal/output/table.go:102-114` — `FormatHeaderSection`.

**Issue:** The `=` border width is `len(title) + 30`, which is arbitrary.
For a title like `"Bucket: x"`, this produces a 39-char border. For
`"Object: very-long-key-name-here"`, it produces a 60-char border.
Consider a fixed width (e.g., 60 chars) or capping at terminal width.

### 4.3 Section Title

Used to divide sections within a detail view.

**Structure:**
```
-- Overview --
```

**Implementation:** `internal/output/table.go:117-119` — `FormatSectionTitle`.

**Consistent usage:** Always followed by `\n` then a table.

### 4.4 Sub-Section Header

Informal text headers used within the access control section.

**Structure:**
```
Identity and Access Management (IAM) Policy:
  (No IAM bindings found)
```

**Implementation:** Inline in `internal/output/storage_views.go`.
Not extracted into a reusable function.

**Issue:** These follow no consistent format convention. Some include
abbreviation expansions (`IAM`), some include parenthetical notes,
some use dashes. Should be normalized to match section title style
or remain as informal sub-headers with a consistent pattern.

### 4.5 Empty State Messages

Displayed when a query returns no results.

**Variants found:**

| Message | Location |
|---------|----------|
| `No buckets found.` | `storage_list_buckets.go` |
| `No SQL instances found.` | `sql_list_instances.go` |
| `No objects or directories found.` | `storage_views.go` |
| `No providers configured. Use 'synkronus config set'. Supported providers: ...` | `storage_list_buckets.go` |
| `No SQL providers configured. Use 'synkronus config set'. Supported SQL providers: ...` | `sql_list_instances.go` |
| `No configuration values set. Use 'synkronus config set <key> <value>'.` | `config_cmd.go:114` |
| `(No IAM bindings found)` | `storage_views.go` |
| `(Could not retrieve IAM policy - check permissions)` | `storage_views.go` |
| `(Inactive because Uniform Bucket-Level Access is enabled...)` | `storage_views.go` |

**Canonical pattern:**
- Simple empty state: `No <resources> found.`
- Actionable empty state: `No <resources> found. Use 'synkronus <command>' to <action>.`
- Parenthetical note within a detail view: `(<explanation>)`

### 4.6 Confirmation Prompt

Used for destructive operations (delete-bucket).

**Structure:**
```
WARNING: You are about to delete the bucket 'my-bucket' on provider 'GCP'.
This action CANNOT be undone and may result in permanent data loss.
To confirm, please type the name 'my-bucket': _
```

**Implementation:** Warning message constructed in `storage_delete_bucket.go:156`,
prompt handled by `internal/ui/prompt/prompt.go:32-51`.

**Pattern:**
1. `WARNING:` prefix (uppercased)
2. Description of what will happen, with resource name in single quotes
3. Severity statement (`CANNOT be undone`)
4. Confirmation instruction with expected input in single quotes

### 4.7 Success Messages

```
Bucket 'my-bucket' created successfully in US on provider gcp.
Bucket 'my-bucket' deleted successfully from provider gcp.
Configuration set: gcp.project = my-project
Configuration key 'gcp.project' deleted
```

**Inconsistency:** Config success messages don't use `successfully` and
don't quote the key. Storage messages quote resource names but show
provider name in lowercase (as typed). The delete warning shows it
uppercase.

**Canonical pattern:**
`<Resource type> '<name>' <past-tense verb> on provider '<normalized-name>'.`

### 4.8 Warning Messages

```
Warning: some providers failed: <error details>
Warning: some SQL providers failed: <error details>
```

**Output stream:** stderr (correct).

**Pattern:** `Warning: <description>` — sentence case after prefix.

### 4.9 Error Messages

```
Error: unsupported provider: xyz. Supported providers are: [gcp aws]
Error: error describing bucket 'x' on gcp: <api error>
```

**Output stream:** stderr via Cobra's error handling (`root.go:61`).

**Pattern:** `Error: <description>` — the `Error:` prefix is added by
`root.go:61`, not by the error itself.

---

## 5. Iconography

**Not applicable.** No icons, emoji, or Unicode symbols are used in
output. The tool uses only ASCII characters.

**Recommendation:** Keep it ASCII-only for maximum terminal
compatibility. If status indicators are needed in the future, use
text markers:

| Concept | Marker | Example |
|---------|--------|---------|
| Directory | `(DIR)` | Already in use (`internal/output/storage_views.go`) |
| Enabled | `Enabled` | Already in use |
| Disabled | `Disabled` / `Suspended` | Already in use |
| Unknown | `N/A` | Already in use |
| Success | `[OK]` or `✓` (with ASCII fallback) | Not yet used |
| Failure | `[FAIL]` or `✗` (with ASCII fallback) | Not yet used |

---

## 6. Motion and Animation

**Not applicable.** There are no spinners, progress bars, or animated
output. All commands block until complete and print results at once.

**Recommendation:** For operations that take more than ~2 seconds
(ListBuckets with monitoring metrics, multi-provider fanout), consider
adding a simple spinner using a library like
[`briandowns/spinner`](https://github.com/briandowns/spinner). Display
on stderr only (so stdout remains pipe-safe). Disable when not a TTY.

---

## 7. Accessibility

### 7.1 Terminal Accessibility Strengths

- **No color dependency:** All information is conveyed through text alone. Color is not used at all currently, so there are zero color-contrast concerns.
- **Screen reader compatible:** Plain text output works natively with terminal screen readers (JAWS, NVDA, VoiceOver terminal mode).
- **Pipe-safe:** Output goes to stdout; errors/warnings go to stderr. This is the correct Unix convention and works with assistive tooling.

### 7.2 Potential Concerns

| Issue | Severity | Details |
|-------|----------|---------|
| **ASCII table borders may confuse screen readers** | Low | Characters like `+`, `-`, `|` are read literally. Long borders produce noise (e.g., "plus dash dash dash dash..."). An alternative `--output json` or `--output plain` mode would help. |
| **No `--no-color` flag** | Low | Not needed currently (no color exists), but should be implemented alongside any color additions. Honor the `NO_COLOR` environment variable. |
| **Wide tables may wrap in narrow terminals** | Low | No terminal width detection exists. Tables with many columns (object list: 4 columns, instance list: 8 columns) could wrap unpredictably in 80-column terminals. |

---

## 8. Inconsistency Report

Every style conflict found in the codebase, with file references and
recommended resolution.

### 8.1 Casing Inconsistencies

| # | File | Line(s) | Current | Recommended | Rationale |
|---|------|---------|---------|-------------|-----------|
| 1 | `internal/output/storage_views.go` | 28 | List headers: `BUCKET NAME` (with space) | `BUCKET_NAME` or keep `BUCKET NAME` | Decide on one convention. Current uses spaces in multi-word headers; this is consistent across all list tables. **Keep as-is.** |
| 2 | `storage_list_buckets.go` | 156 | `strings.ToUpper(providerName)` in delete warning | Use `strings.ToUpper` consistently | Success messages at lines 130 and 173 show provider as-is (user input, usually lowercase). Warning shows UPPERCASE. Pick one. **Recommend:** Always display provider uppercase in user-facing messages (matches `common.Provider` constants `GCP`, `AWS`). |
| 3 | `internal/output/storage_views.go` | 120 | `strings.ToTitle(bucket.PublicAccessPrevention)` | `strings.ToUpper` or leave as API returns | `ToTitle` capitalizes every word, which is correct for enum-like values like `enforced` → `Enforced`. Acceptable but inconsistent with other enum values displayed as-is. **Keep as-is.** |

### 8.2 Spacing Inconsistencies

| # | File | Line(s) | Current | Recommended |
|---|------|---------|---------|-------------|
| 4 | `internal/output/storage_views.go` | 183 | IAM section ends with `sb.WriteString("\n")` then caller adds `\n\n` via other sections | Produces triple newline when `HasConditions` is true. Remove extra `\n` at line 183. |
| 5 | `internal/output/storage_views.go` | 352 | `FormatObjectList` returns table without trailing newline | Add `\n` for consistency with other format methods. |
| 6 | `internal/output/table.go` | 105 | `FormatHeaderSection` border width = `len(title) + 30` | Use a fixed width (60 chars) or `max(len(title) + 6, 60)` for consistent visual weight. |

### 8.3 Structural Inconsistencies

| # | File | Line(s) | Current | Recommended |
|---|------|---------|---------|-------------|
| 7 | `internal/output/storage_views.go` | 318-321 | Object list has `Listing objects in bucket: x` prefix | Bucket list and instance list have no prefix. Either add a context line to all list commands or remove from object list. **Recommend:** Keep for object list only (it requires bucket context), but format consistently: `Objects in bucket 'x':` (quoted, colon, no `Listing`). |
| 8 | `internal/output/storage_views.go` | 149,193,199 | Sub-section headers are inline strings, not using `FormatSectionTitle` | Extract a `FormatSubSectionTitle(title string) string` that returns `title + ":"`. Use consistently for IAM Policy, ACLs, and similar within-section headers. |
| 9 | `config_cmd.go` | 36 | `Configuration set: gcp.project = my-value` | Use `Config key 'gcp.project' set to 'my-value'.` to match the quoting convention used elsewhere. |
| 10 | `config_cmd.go` | 84 | `Configuration key 'gcp.project' deleted` | Missing period at end. Add `.` for consistency. |
| 11 | `config_cmd.go` | 124-126 | Config list uses `fmt.Printf("  %s = %v\n", ...)` (2-space indent) | All other structured output uses tables. Consider using a key-value table, or keep as-is since it's intentionally simpler. **Keep as-is** — config list is a lightweight command. |

### 8.4 Date Format Inconsistencies

| # | File | Line(s) | Current | Recommended |
|---|------|---------|---------|-------------|
| 12 | `internal/output/storage_views.go` | 32 | Bucket list date: `2006-01-02` (ISO date) | Keep. Compact for table columns. |
| 13 | `internal/output/storage_views.go` | 93-94 | Bucket detail date: `time.RFC1123` (`Mon, 02 Jan 2006 15:04:05 MST`) | Keep. Detailed for describe views. |
| 14 | `internal/output/storage_views.go` | 348 | Object list date: `time.RFC3339` (`2006-01-02T15:04:05Z07:00`) | **Change to `2006-01-02`** for consistency with other list tables, or `2006-01-02 15:04` if time matters. RFC3339 is verbose for a table column. |
| 15 | `internal/output/sql_views.go` | 23-25 | Instance list date: `2006-01-02` (matches bucket list) | Consistent. Keep. |
| 16 | `internal/output/sql_views.go` | 85 | Instance detail date: `time.RFC1123` (matches bucket detail) | Consistent. Keep. |

**Canonical rule:**
- List tables: `2006-01-02` (compact ISO date)
- Detail views: `time.RFC1123` (human-readable with timezone)

### 8.5 Message Pattern Inconsistencies

| # | File | Line(s) | Current | Recommended |
|---|------|---------|---------|-------------|
| 17 | `storage_create_bucket.go` | 130 | `Bucket '%s' created successfully in %s on provider %s.\n` | Good. Keep as canonical success pattern. |
| 18 | `storage_delete_bucket.go` | 173 | `Bucket '%s' deleted successfully from provider %s.\n` | Good. Matches pattern. |
| 19 | `storage_delete_bucket.go` | 163 | `Deletion aborted: Confirmation mismatch or cancelled.` | Change to `Deletion aborted. Confirmation did not match.` — clearer, no colon (which looks like a log prefix). |
| 20 | `config_cmd.go` | 36 | `Configuration set: %s = %s` | Change to `Config '%s' set to '%s'.` for consistency with quoting convention. |

---

## Summary of Recommended Changes

### Priority 1 — Fix Before Adding Features (Small effort each)

1. Normalize provider name casing in user-facing messages to uppercase
2. Fix triple-newline bug in IAM section (`internal/output/storage_views.go:183`)
3. Change object list timestamp from RFC3339 to `2006-01-02`
4. Add trailing newline to `FormatObjectList` output

### Priority 2 — Standardize Patterns (Medium effort)

5. Extract `FormatSubSectionTitle` for within-section headers
6. Standardize success/config message format with consistent quoting
7. Fix `FormatHeaderSection` border width to use fixed or capped width

### Priority 3 — Add When Implementing New Features

8. Add ANSI color (error=red, warning=yellow, success=green, headers=bold)
9. Honor `NO_COLOR` env var and TTY detection
10. ~~Add `--output json`~~ **DONE** — `--output table|json|yaml` implemented in `internal/output/`
11. Add spinner for long-running operations (stderr only)
