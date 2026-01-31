---
description: Catch up on recent golang/go Changes (CLs) and learn Go philosophy
allowed-tools: Bash, WebFetch, Read
argument-hint: [category]
---

# Go Change Catchup

Fetches and analyzes Changes (CLs) updated in the last month from golang/go repository to learn Go design philosophy.

## Arguments

- `$1`: Category filter (optional: error-handling, performance, api-design, testing, runtime, compiler)

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
```

## Notes

- Changes are fetched from Gerrit (go-review.googlesource.com), not GitHub
- Use Change numbers (e.g., 3965) or full Change IDs (e.g., go~master~I8473b95934b5732ac55d26311a706c9c2bde9940)
- Analysis includes patch set evolution, which shows how changes were refined through review
