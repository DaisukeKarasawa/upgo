---
description: Catch up on recent golang/go Changes (CLs) and learn Go philosophy
allowed-tools: Bash
argument-hint: [category]
---

# /go-catchup

Fetches and analyzes Changes (CLs) updated in the last month from golang/go repository to learn Go design philosophy.

## Command Name

`/go-catchup [category]`

## One-Line Description

Fetches merged Changes (CLs) updated in the last 30 days from Gerrit, analyzes review discussions, and reports key insights and Go philosophy.

## Arguments

- `$1` (`category`): Category filter (optional)
  - **Type**: string
  - **Required**: No (optional)
  - **Default**: None (all categories)
  - **Constraints**: Recommended category strings (not strictly enforced):
    - `error-handling`, `performance`, `api-design`, `testing`, `runtime`, `compiler`
  - **Description**: Prioritizes extraction and summarization of Changes from the specified category perspective (e.g., specifying `compiler` focuses on compiler-related Changes).

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

- A report containing a list and key points of Changes (CLs) from the last 30 days, category trends, detailed analysis of notable Changes, and Go philosophy insights
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

1. Get change details with messages and comments
2. Review change messages and inline comments
3. Analyze patch set evolution
4. Extract Go philosophy insights

Use `go-pr-analyzer` skill for detailed analysis patterns. See `skills/go-pr-analyzer/SKILL.md` for analysis perspectives and output format.

### 4. Generate Report

Create report in the following format:

```markdown
# Go Change Catchup Report

**Period**: Last 30 days
**Count**: <count> Changes

## Summary

### By Category

- error-handling: X Changes
- performance: Y Changes
- ...

### Notable Changes

1. Change #XXXX: <subject> - <one-line description>
2. Change #YYYY: <subject> - <one-line description>

---

## Detailed Analysis

### Change #XXXX: <subject>

**Summary**: <summary>

**Go Philosophy**: <learnings>

**Patch Sets**: <number> patch sets, showing iterative refinement

---

## Key Takeaways

<Overall Go philosophy and best practices learned>
```

## Usage Examples

```bash
# Catch up on all Changes from the last month
/go-catchup

# Filter by error-handling category
/go-catchup error-handling

# Filter by performance category
/go-catchup performance

# Filter by compiler category
/go-catchup compiler
```

## Related Commands

This command orchestrates the following primitive commands:

- `/go-changes-fetch`: Fetches Change list from Gerrit
- `/go-change-analyze`: Analyzes individual Changes

For more granular control, use the primitive commands directly.

## Notes

- Changes are fetched from Gerrit (go-review.googlesource.com), not GitHub
- Use Change numbers (e.g., 3965) or full Change IDs (e.g., go~master~I8473b95934b5732ac55d26311a706c9c2bde9940)
- Analysis includes patch set evolution, which shows how changes were refined through review
- This is an **orchestrator command** that combines fetch + analyze + report. See `commands/NAMING.md` for design principles.
- For Gerrit API helper function and common patterns, see `skills/go-gerrit-reference/REFERENCE.md`
