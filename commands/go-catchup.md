---
description: Catch up on recent golang/go Changes (CLs) and learn Go philosophy
allowed-tools: Bash, WebFetch, Read
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
  - Makes API calls to Gerrit (default: `https://go-review.googlesource.com`) over the network
  - Does not create or update local files (paste the output if you want to save it)

## Prerequisites (Required State / Permissions / Files)

- **Required Commands**: `curl`, `jq`, `sed`
- **Environment Variables (Required)**:
  - `GERRIT_USER`
  - `GERRIT_HTTP_PASSWORD` (obtain from `https://go-review.googlesource.com/settings/#HTTPCredentials`)
- **Environment Variables (Optional)**:
  - `GERRIT_BASE_URL` (default: `https://go-review.googlesource.com`)
- **Prerequisites**: Network access to Gerrit must be available

## Expected State After Execution

- A report containing a list and key points of Changes (CLs) from the last 30 days, category trends, detailed analysis of notable Changes, and Go philosophy insights
- Local files/settings are not modified

## Execution Steps

### 1. Environment Check

```bash
# Check curl command
which curl || echo "ERROR: curl command not found. Please install curl."

# Check jq command
which jq || echo "ERROR: jq command not found. Please install jq."

# Check Gerrit environment variables
if [ -z "$GERRIT_USER" ] || [ -z "$GERRIT_HTTP_PASSWORD" ]; then
  echo "ERROR: GERRIT_USER and GERRIT_HTTP_PASSWORD must be set"
  echo "Visit https://go-review.googlesource.com/settings/#HTTPCredentials to get HTTP password"
  exit 1
fi

# Helper function to fetch Gerrit API and strip XSSI prefix
gerrit_api() {
  local endpoint="$1"
  local base_url="${GERRIT_BASE_URL:-https://go-review.googlesource.com}"
  local raw

  # Capture curl output first, preserving exit status
  # -S flag shows errors even in silent mode for better diagnostics
  raw="$(curl -fsS -u "${GERRIT_USER}:${GERRIT_HTTP_PASSWORD}" "${base_url}/a${endpoint}")" || return $?

  # Strip XSSI prefix if present
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
