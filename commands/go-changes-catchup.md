---
description: Catch up on recent golang/go Changes (CLs) and learn Go philosophy
allowed-tools: Bash
argument-hint: [category]
---

# /go-changes-catchup

Fetches and analyzes Changes (CLs) updated in the last month from golang/go repository to learn Go design philosophy.

## Command Name

`/go-changes-catchup [category] [--days=N] [--status=STATUS] [--limit=N] [--review=lite|full] [--deep-dives=N] [--deep-criteria=CRITERIA] [--format=md|json]`

## One-Line Description

Fetches Changes (CLs) updated in the last 30 days from Gerrit (default: merged), analyzes review discussions with detailed comment analysis, and generates a comprehensive report covering all Changes with deep dives into Go design philosophy and review culture.

## Arguments

- `$1` (`category`): Category filter (optional)

  - **Type**: string
  - **Required**: No (optional)
  - **Default**: None (all categories)
  - **Constraints**: Recommended category strings (not strictly enforced):
    - `error-handling`, `performance`, `api-design`, `testing`, `documentation`, `tooling`, `runtime`, `compiler`, `standard-library`, `language-spec`
  - **Description**: Prioritizes extraction and summarization of Changes from the specified category perspective (e.g., specifying `compiler` focuses on compiler-related Changes). When specified, deep dives are prioritized from this category.

- `--days=N`: Number of days to look back (optional)

  - **Type**: number
  - **Required**: No
  - **Default**: `30`
  - **Description**: Fetches Changes updated within the last N days (using `-age` operator)

- `--status=STATUS`: Change status filter (optional)

  - **Type**: string
  - **Required**: No
  - **Default**: `merged`
  - **Constraints**: `open`, `merged`, `abandoned`, or empty (all statuses)
  - **Description**: Filters Changes by status

- `--limit=N`: Maximum number of Changes to fetch (optional)

  - **Type**: number
  - **Required**: No
  - **Default**: `50`
  - **Description**: Limits the number of Changes returned

- `--review=lite|full`: Review analysis depth for all Changes (optional)

  - **Type**: string
  - **Required**: No
  - **Default**: `lite`
  - **Constraints**: `lite` (metrics + representative points) or `full` (detailed inline comments with file:line)
  - **Description**: Controls review analysis depth. Note: Deep dive Changes always include full review analysis regardless of this setting.

- `--deep-dives=N`: Number of Changes to analyze in depth (optional)

  - **Type**: number
  - **Required**: No
  - **Default**: `8`
  - **Description**: Number of Changes selected for deep dive analysis (includes file:line comments, thread analysis, patch set evolution). When category is specified, deep dives are prioritized from that category.

- `--deep-criteria=CRITERIA`: Selection criteria for deep dives (optional)

  - **Type**: string
  - **Required**: No
  - **Default**: `mixed`
  - **Constraints**: `review` (comment count), `ps` (patch set count), `labels` (label evolution), `mixed` (scoring based on multiple factors)
  - **Description**: How to select Changes for deep dive. `mixed` considers comment count, patch set count, label evolution, important areas (runtime/compiler/stdlib), and design discussion presence.

- `--format=md|json`: Output format (optional)
  - **Type**: string
  - **Required**: No
  - **Default**: `md`
  - **Constraints**: `md` (Markdown) or `json` (structured JSON)
  - **Description**: Output format. JSON format is useful for saving or further processing.

## Output / Side Effects

- **Output**: Displays a Markdown-formatted "Go Change Catchup Report" in chat
- **Side Effects**:
  - Makes API calls to Gerrit (default: `https://go-review.googlesource.com`) over the network using anonymous access
  - Does not create or update local files (paste the output if you want to save it)

## Prerequisites (Required State / Permissions / Files)

- **Required Commands**: `curl`, `jq`, `sed`
- **Environment Variables (Optional)**:
  - `GERRIT_BASE_URL`: Gerrit server URL (default: `https://go-review.googlesource.com`)
- **Prerequisites**: Network access to Gerrit must be available

## Expected State After Execution

- A comprehensive report containing:
  - **Full Index**: All fetched Changes (up to limit) organized by category, each with 1-2 line summary and review-lite metrics
  - **Category Digest**: Trends and common patterns within each category
  - **Deep Dives**: Selected Changes with detailed review analysis (file:line comments, thread discussions, patch set evolution)
  - **Review Patterns & Best Practices**: Cross-cutting patterns from review comments and Go coding culture insights
- Local files/settings are not modified

## Execution Steps

### 1. Environment Check

Use the environment check pattern from `go-gerrit-reference` skill. See `skills/go-gerrit-reference/REFERENCE.md` for complete helper function.

```bash
# Check required commands
if ! command -v curl >/dev/null 2>&1; then
  echo "ERROR: curl command not found. Please install curl."
  exit 1
fi

if ! command -v jq >/dev/null 2>&1; then
  echo "ERROR: jq command not found. Please install jq."
  exit 1
fi

if ! command -v sed >/dev/null 2>&1; then
  echo "ERROR: sed command not found. Please install sed."
  exit 1
fi

# Load gerrit_api() helper function (see skills/go-gerrit-reference/REFERENCE.md)
gerrit_api() {
  local endpoint="$1"
  local base_url="${GERRIT_BASE_URL:-https://go-review.googlesource.com}"
  local raw

  # Uses anonymous access (no authentication required)
  raw="$(curl -fsS "${base_url}${endpoint}")" || return $?
  printf '%s\n' "$raw" | sed "1s/^)]}'//"
}
```

### 2. Fetch Change List

```bash
# Fetch merged changes updated in the last month using -age operator
gerrit_api "/changes/?q=project:go+status:merged+-age:30d&n=50&o=LABELS&o=DETAILED_ACCOUNTS&o=CURRENT_REVISION&o=CURRENT_COMMIT" | jq '.[] | {_number, subject, owner, submitted, updated, labels}'
```

### 3. Analyze Each Change

For each fetched Change:

1. **Lite Analysis (All Changes)**:

   - Get change details with messages and comments
   - Extract review metrics: patch set count, comment count, label evolution
   - Identify primary category and tags
   - Extract 1-2 representative discussion points

2. **Deep Dive Analysis (Selected Changes)**:
   - Thread analysis: Follow `in_reply_to` to map comment threads (issue → response → fix)
   - Inline comment details: Extract file:line references and comment content
   - Patch set evolution: Compare revisions to identify how code improved through review
   - Design discussion extraction: Identify design decisions and alternatives discussed
   - Go philosophy connection: Map changes to Go principles (simplicity, explicitness, orthogonality, practicality)

Use `go-pr-analyzer` skill for detailed analysis patterns. See `skills/go-pr-analyzer/SKILL.md` for analysis perspectives and output format.

### 3.1. Select Deep Dive Candidates

Score each Change based on:

- Comment count (higher = more discussion)
- Patch set count (higher = more iteration)
- Label evolution (Code-Review/Verified changes indicate discussion)
- Important areas (runtime, compiler, standard library)
- Design discussion presence (messages contain design-related keywords)

Select top N Changes (default: 8) based on `--deep-criteria` setting.

### 4. Generate Report

Create report in the following format:

```markdown
# Go Change Catchup Report

**Period**: Last <days> days
**Total Changes**: <count> Changes
**Deep Dives**: <deep-dive-count> Changes

## Summary

### By Category

- error-handling: X Changes
- performance: Y Changes
- api-design: Y Changes
- testing: Z Changes
- documentation: W Changes
- tooling: V Changes
- runtime: U Changes
- compiler: T Changes
- standard-library: S Changes
- language-spec: R Changes
- other: Q Changes

### Review Activity Overview

- Average patch sets per Change: <avg-ps>
- Average comments per Change: <avg-comments>
- Most active reviewers: <reviewer-list>
- Changes with design discussions: <design-discussion-count>

---

## Detailed Changes by Category

### error-handling (X Changes)

#### Change #XXXX: <subject>

- **Summary**: <1-2 line description of what changed and why>
- **Owner**: <owner.name> | **Merged**: <submitted-date>
- **Category**: error-handling (tags: <tag1>, <tag2>)
- **Review Metrics**: <ps-count> patch sets, <comment-count> comments
- **Key Discussion Points**: <1-2 representative points from review>

#### Change #YYYY: <subject>

...

### performance (Y Changes)

...

### [Other categories...]

---

## Notable Deep Dives

### Change #XXXX: <subject>

**Summary**: <detailed summary>

**Background**: <why this change was needed>

**Review Process**:

- **Reviewers**: <reviewer-list>
- **Review Rounds**: <patch-set-count> patch sets
- **Key Review Comments**:
  - `<file>:<line>` - `<reviewer-name>`: "<comment-content>" → <how-addressed>
  - `<file>:<line>` - `<reviewer-name>`: "<comment-content>" → <how-addressed>
- **Code Evolution**:
  - PS1 → PS2: <what changed based on review>
  - PS2 → PS3: <what changed based on review>
- **Thread Discussions**:
  - Issue: "<initial-concern>" → Response: "<author-response>" → Resolution: "<final-state>"

**Go Philosophy Alignment**: <how this change embodies Go principles>

**Key Learnings**: <practical insights from review culture>

---

### Change #YYYY: <subject>

...

---

## Review Patterns & Best Practices

### Common Review Feedback Patterns

Patterns observed across multiple Changes:

- **Naming Conventions**: <description> - Examples: Change #XXXX, #YYYY
- **Error Handling**: <description> - Examples: Change #ZZZZ, #AAAA
- **API Design**: <description> - Examples: Change #BBBB, #CCCC
- **Testing**: <description> - Examples: Change #DDDD, #EEEE
- **Performance Considerations**: <description> - Examples: Change #FFFF, #GGGG

### Go Code Review Best Practices

Insights from review culture:

- <best-practice-1>: <explanation with examples>
- <best-practice-2>: <explanation with examples>
- <best-practice-3>: <explanation with examples>

### Go Philosophy in Practice

How review discussions reflect Go design principles:

- **Simplicity**: <examples where simplicity was emphasized>
- **Explicitness**: <examples where explicitness was emphasized>
- **Orthogonality**: <examples where orthogonality was emphasized>
- **Practicality**: <examples where practicality was emphasized>

---

## Key Takeaways

<Overall Go design philosophy and best practices learned from review culture>
```

## Usage Examples

```bash
# Catch up on all Changes from the last month (default: 30 days, merged, limit 50)
/go-changes-catchup

# Filter by error-handling category
/go-changes-catchup error-handling

# Filter by compiler category with custom deep dive count
/go-changes-catchup compiler --deep-dives=10

# Fetch changes from last 7 days with full review analysis
/go-changes-catchup --days=7 --review=full

# Custom limit and deep dive criteria
/go-changes-catchup --limit=30 --deep-dives=5 --deep-criteria=review

# Output as JSON for further processing
/go-changes-catchup --format=json > catchup-report.json

# Filter by runtime category with specific review depth
/go-changes-catchup runtime --review=lite --deep-dives=6
```

## Output Examples

### Full Index Example (Category Section)

```markdown
### error-handling (5 Changes)

#### Change #12345: improve error wrapping in runtime

- **Summary**: Adds context to error messages in runtime package to improve debugging experience
- **Owner**: alice@example.com | **Merged**: 2026-01-15
- **Category**: error-handling (tags: runtime, debugging)
- **Review Metrics**: 3 patch sets, 12 comments
- **Key Discussion Points**: Discussion on error message format consistency, performance impact of error wrapping

#### Change #12346: standardize error handling in net/http

- **Summary**: Unifies error handling patterns across HTTP client code
- **Owner**: bob@example.com | **Merged**: 2026-01-18
- **Category**: error-handling (tags: standard-library, api-design)
- **Review Metrics**: 2 patch sets, 8 comments
- **Key Discussion Points**: Backward compatibility concerns, API design consistency
```

### Deep Dive Example

```markdown
### Change #12345: improve error wrapping in runtime

**Summary**: This change adds context information to error messages in the runtime package to improve debugging experience. The implementation wraps errors with additional context while maintaining backward compatibility.

**Background**: Users reported difficulty debugging runtime errors due to lack of context. This change addresses the issue by adding structured error wrapping.

**Review Process**:

- **Reviewers**: @rsc, @ianlancetaylor, @bradfitz
- **Review Rounds**: 3 patch sets
- **Key Review Comments**:
  - `src/runtime/error.go:42` - @rsc: "Consider using fmt.Errorf with %w instead of custom wrapper" → Updated to use standard error wrapping
  - `src/runtime/error.go:58` - @ianlancetaylor: "Performance concern: error wrapping adds allocation" → Added benchmarks, confirmed acceptable overhead
  - `src/runtime/error_test.go:15` - @bradfitz: "Add test case for error chain unwrapping" → Added comprehensive test coverage
- **Code Evolution**:
  - PS1 → PS2: Switched from custom wrapper to fmt.Errorf with %w per review feedback
  - PS2 → PS3: Added performance benchmarks and optimized hot path based on review discussion
- **Thread Discussions**:
  - Issue: "Should we use errors.Is/errors.As for error checking?" → Response: "Yes, but need to ensure compatibility" → Resolution: "Added compatibility layer, documented migration path"

**Go Philosophy Alignment**: This change embodies **explicitness** (clear error context) and **practicality** (addresses real debugging needs). The review discussion emphasized maintaining **simplicity** by using standard library patterns (fmt.Errorf with %w) rather than custom solutions.

**Key Learnings**:

- Go's standard error wrapping (`fmt.Errorf` with `%w`) is preferred over custom wrappers for consistency
- Performance considerations are important even for error paths, but should be measured rather than assumed
- Backward compatibility is crucial - changes should be additive rather than breaking
```

### Review Patterns Example

```markdown
## Review Patterns & Best Practices

### Common Review Feedback Patterns

- **Error Handling**: Multiple Changes emphasized using `fmt.Errorf` with `%w` for error wrapping rather than custom solutions. Examples: Change #12345, #12350
- **Naming Conventions**: Reviewers consistently asked for more descriptive names, especially for exported functions. Examples: Change #12346, #12352
- **API Design**: Discussion often focused on backward compatibility and clear migration paths. Examples: Change #12347, #12351
- **Testing**: Reviewers emphasized comprehensive test coverage, especially edge cases. Examples: Change #12348, #12353

### Go Code Review Best Practices

- **Prefer Standard Library**: When standard library provides functionality (e.g., error wrapping), use it rather than custom implementations
- **Measure Performance**: Don't assume performance impact - add benchmarks to verify
- **Document Migration**: For API changes, provide clear migration paths and examples
- **Test Edge Cases**: Go's testing culture values comprehensive edge case coverage
```

## Related Commands

This command orchestrates the following primitive commands:

- `/go-changes-fetch`: Fetches Change list from Gerrit
- `/go-change-analyze`: Analyzes individual Changes

For more granular control, use the primitive commands directly.

## Notes

- Changes are fetched from Gerrit (go-review.googlesource.com), not GitHub
- Use Change numbers (e.g., 3965) or full Change IDs (e.g., go~master~I8473b95934b5732ac55d26311a706c9c2bde9940)
- **All fetched Changes are included in the Full Index** - no Changes are omitted
- Analysis includes patch set evolution, which shows how changes were refined through review
- Deep dive analysis includes detailed review comments with file:line references, thread discussions, and code evolution
- This is an **orchestrator command** that combines fetch + analyze + report. See `commands/NAMING.md` for design principles.
- For Gerrit API helper function and common patterns, see `skills/go-gerrit-reference/REFERENCE.md`
- **Rate Limiting**: When analyzing many Changes, the command may take time due to API rate limits. If deep dive analysis fails for some Changes, the report will continue with available data and reduce deep dive count automatically.
- **Review Analysis**: Review-lite includes patch set count, comment count, and 1-2 representative discussion points. Deep dives include full inline comment analysis with file:line references, thread mapping, and patch set evolution details.
