---
name: go-pr-fetcher
description: |
  Fetches Change (CL) information from golang/go repository via Gerrit REST API.
  Use when user says "fetch Go changes", "latest CLs from golang/go", "check Go changes", etc.
allowed-tools:
  - Bash
---

# Go Change Fetcher

Fetches Change (CL) information from golang/go repository using Gerrit REST API.

## Side Effects

- **Network Access**: Makes API calls to Gerrit server (default: `https://go-review.googlesource.com`) using anonymous access
- **Anonymous Access**: Uses anonymous API access (no authentication required). Some Gerrit instances may not support anonymous access and may return 401/403 errors. Rate limits may be stricter for anonymous access.
- **No Local File Changes**: Does not create or modify local files

## Prerequisites

### Required Commands

- `curl`: HTTP client for API requests
- `jq`: JSON processor for parsing responses
- `sed`: Text processing for XSSI prefix removal

### Optional Environment Variables

- `GERRIT_BASE_URL`: Gerrit server URL (default: `https://go-review.googlesource.com`)

## Quick Start

Use the `gerrit_api()` helper function from `go-gerrit-reference` skill. See [go-gerrit-reference/REFERENCE.md](../go-gerrit-reference/REFERENCE.md) for complete helper function.

## Fetching Change List

### Get recent changes (default: 30)

```bash
# Get all changes from go project, limit 30, with labels and detailed accounts
gerrit_api "/changes/?q=project:go+status:open&n=30&o=LABELS&o=DETAILED_ACCOUNTS" | jq '.[] | {_number, subject, status, owner, updated, labels}'
```

### Get merged changes only

```bash
# Get merged changes from go project
gerrit_api "/changes/?q=project:go+status:merged&n=30&o=LABELS&o=DETAILED_ACCOUNTS" | jq '.[] | {_number, subject, owner, submitted, labels}'
```

### Get changes from specific period

```bash
# Last 7 days (updated after date)
DATE_7DAYS_AGO=$(date -v-7d +%Y-%m-%d 2>/dev/null || date -d '7 days ago' +%Y-%m-%d)
gerrit_api "/changes/?q=project:go+after:${DATE_7DAYS_AGO}&n=50&o=LABELS&o=DETAILED_ACCOUNTS" | jq '.[] | {_number, subject, status, owner, updated}'
```

### Get changes updated in last month

```bash
# Last 30 days using -age operator
gerrit_api "/changes/?q=project:go+-age:30d&n=50&o=LABELS&o=DETAILED_ACCOUNTS" | jq '.[] | {_number, subject, status, owner, updated, labels}'
```

## Fetching Individual Change Details

For complete API reference including helper function, change ID formats, and error handling, see [go-gerrit-reference/REFERENCE.md](../go-gerrit-reference/REFERENCE.md).

### Change basic info

```bash
CHANGE_ID="<change-number>"  # e.g., 3965 or go~master~I8473b95934b5732ac55d26311a706c9c2bde9940
gerrit_api "/changes/${CHANGE_ID}/detail?o=LABELS&o=DETAILED_LABELS&o=MESSAGES&o=REVIEWER_UPDATES&o=CURRENT_REVISION&o=CURRENT_COMMIT&o=CURRENT_FILES" | jq '{_number, subject, status, owner, created, updated, labels, messages}'
```

### Change comments and discussions

```bash
# Get all published comments across all revisions
gerrit_api "/changes/${CHANGE_ID}/comments" | jq '.'

# Get change messages (cover messages and system messages)
gerrit_api "/changes/${CHANGE_ID}/messages" | jq '.[] | {id, author, date, message, _revision_number}'
```

### Change diff

```bash
# Get raw patch (unified diff format)
gerrit_api "/changes/${CHANGE_ID}/revisions/current/patch?raw" | cat

# Get structured diff for a specific file
FILE_PATH="src/example.go"  # URL encode if needed
gerrit_api "/changes/${CHANGE_ID}/revisions/current/files/${FILE_PATH}/diff" | jq '.'
```

### Deep Dive: Patch Set Evolution

For analyzing how code evolved through review:

```bash
# Get all revisions for a change
gerrit_api "/changes/${CHANGE_ID}/detail?o=ALL_REVISIONS" | jq '.revisions | keys'

# Get list of all revisions
REVISIONS=$(gerrit_api "/changes/${CHANGE_ID}/detail?o=ALL_REVISIONS" | jq -r '.revisions | keys[]')

# Compare two specific revisions
REV1="<revision-hash-1>"
REV2="<revision-hash-2>"
gerrit_api "/changes/${CHANGE_ID}/revisions/${REV2}/files?base=${REV1}" | jq '.'

# Get diff between consecutive revisions
# First, get revision list
REV_LIST=$(gerrit_api "/changes/${CHANGE_ID}/detail?o=ALL_REVISIONS" | jq -r '.revisions | keys | sort')
# Then compare each pair (requires bash loop)
```

### Deep Dive: Review Comments with Thread Structure

For detailed review analysis:

```bash
# Get all comments (includes inline comments with file:line)
gerrit_api "/changes/${CHANGE_ID}/comments" | jq '.'

# Get comments for a specific revision
REV="<revision-hash>"
gerrit_api "/changes/${CHANGE_ID}/revisions/${REV}/comments" | jq '.'

# Extract comment threads (comments with in_reply_to)
gerrit_api "/changes/${CHANGE_ID}/comments" | jq 'to_entries[] | select(.value[].in_reply_to != null)'

# Get comment count per file
gerrit_api "/changes/${CHANGE_ID}/comments" | jq 'to_entries | map({file: .key, count: (.value | length)})'
```

### Deep Dive: Rate Limiting Considerations

When fetching data for multiple Changes or deep dive analysis:

1. **Batch Requests**: Group related API calls when possible
2. **Caching**: Cache Change details if analyzing multiple times
3. **Progressive Loading**: Fetch lite data first, then deep dive data for selected Changes
4. **Error Handling**: Implement retry logic with exponential backoff for rate limit errors (HTTP 429)

```bash
# Example: Fetch lite data for all Changes first
for change_id in "${change_ids[@]}"; do
  # Lite: detail + comments count only
  gerrit_api "/changes/${change_id}/detail?o=MESSAGES&o=CURRENT_REVISION" | jq '{_number, subject, messages: (.messages | length)}'
done

# Then deep dive for selected Changes
for change_id in "${selected_change_ids[@]}"; do
  # Full: all revisions, comments, diff
  gerrit_api "/changes/${change_id}/detail?o=ALL_REVISIONS&o=MESSAGES&o=CURRENT_FILES" | jq '.'
  gerrit_api "/changes/${change_id}/comments" | jq '.'
done
```

### Change commit message

```bash
# Get commit message
gerrit_api "/changes/${CHANGE_ID}/message" | jq '{subject, full_message, footers}'
```

## Output Format

Format fetched Change information as follows:

```markdown
## Change #<number>: <subject>

**Status**: <status> | **Owner**: <owner.name> | **Updated**: <updated>

**Labels**: <labels>

- Code-Review: <value>
- Verified: <value>

### Summary

<commit message summary>

### Changed Files

- <file1>
- <file2>
```

## Change ID Formats

For supported Change ID formats, see [go-gerrit-reference/REFERENCE.md](../go-gerrit-reference/REFERENCE.md).

## Error Handling

For common errors and handling patterns, see [go-gerrit-reference/REFERENCE.md](../go-gerrit-reference/REFERENCE.md).

### Rate Limiting

When performing deep dive analysis on multiple Changes:

- **Anonymous Access Limits**: Anonymous API access may have stricter rate limits
- **429 Too Many Requests**: If you receive HTTP 429, implement exponential backoff:

  ```bash
  # Simple retry with backoff
  retry_with_backoff() {
    local max_attempts=3
    local delay=1
    local attempt=1

    while [ $attempt -le $max_attempts ]; do
      if result=$(gerrit_api "$1" 2>&1); then
        echo "$result"
        return 0
      fi

      if echo "$result" | grep -q "429"; then
        echo "Rate limited, waiting ${delay}s..." >&2
        sleep $delay
        delay=$((delay * 2))
        attempt=$((attempt + 1))
      else
        return 1
      fi
    done

    return 1
  }
  ```

- **Progressive Fetching**: Fetch lite data for all Changes first, then deep dive data for selected Changes only
- **Caching**: Cache Change details to avoid redundant API calls
