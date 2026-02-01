# Gerrit API Reference

Complete reference for Gerrit REST API usage with golang/go repository.

## Prerequisites

### Required Commands

- `curl`: HTTP client for API requests
- `jq`: JSON processor for parsing responses
- `sed`: Text processing for XSSI prefix removal

### Optional Environment Variables

- `GERRIT_BASE_URL`: Gerrit server URL (default: `https://go-review.googlesource.com`)

## Helper Function

All Gerrit API responses start with `)]}'` (XSSI protection). This helper function strips the prefix and uses anonymous access:

```bash
# Helper function to fetch Gerrit API and strip XSSI prefix
# Uses anonymous access (no authentication required)
gerrit_api() {
  local endpoint="$1"
  local base_url="${GERRIT_BASE_URL:-https://go-review.googlesource.com}"
  local raw

  # Capture curl output first, preserving exit status
  # -S flag shows errors even in silent mode for better diagnostics
  # Note: Uses anonymous API (no /a prefix, no authentication)
  raw="$(curl -fsS "${base_url}${endpoint}")" || return $?

  # Strip XSSI prefix if present
  printf '%s\n' "$raw" | sed "1s/^)]}'//"
}
```

## Environment Check

Before using `gerrit_api()`, verify prerequisites:

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
```

## Common Query Patterns

### Fetch Change List

#### Recent changes (default: 30)

```bash
# Get all changes from go project, limit 30, with labels and detailed accounts
gerrit_api "/changes/?q=project:go+status:open&n=30&o=LABELS&o=DETAILED_ACCOUNTS" | jq '.[] | {_number, subject, status, owner, updated, labels}'
```

#### Merged changes only

```bash
# Get merged changes from go project
gerrit_api "/changes/?q=project:go+status:merged&n=30&o=LABELS&o=DETAILED_ACCOUNTS" | jq '.[] | {_number, subject, owner, submitted, labels}'
```

#### Changes from specific period

```bash
# Last 7 days (updated after date)
DATE_7DAYS_AGO=$(date -v-7d +%Y-%m-%d 2>/dev/null || date -d '7 days ago' +%Y-%m-%d)
gerrit_api "/changes/?q=project:go+after:${DATE_7DAYS_AGO}&n=50&o=LABELS&o=DETAILED_ACCOUNTS" | jq '.[] | {_number, subject, status, owner, updated}'
```

#### Changes updated in last month

```bash
# Last 30 days using -age operator
gerrit_api "/changes/?q=project:go+-age:30d&n=50&o=LABELS&o=DETAILED_ACCOUNTS" | jq '.[] | {_number, subject, status, owner, updated, labels}'
```

### Fetch Individual Change Details

#### Change basic info

```bash
# Get change detail with labels, messages, and reviewer updates
# Change ID format: go~master~I<change-id> or just <change-number>
CHANGE_ID="go~master~I<change-id>"  # or use change number: <change-number>
gerrit_api "/changes/${CHANGE_ID}/detail?o=LABELS&o=DETAILED_LABELS&o=MESSAGES&o=REVIEWER_UPDATES&o=CURRENT_REVISION&o=CURRENT_COMMIT&o=CURRENT_FILES" | jq '{_number, subject, status, owner, created, updated, labels, messages}'
```

#### Change comments and discussions

```bash
# Get all published comments across all revisions
gerrit_api "/changes/${CHANGE_ID}/comments" | jq '.'

# Get change messages (cover messages and system messages)
gerrit_api "/changes/${CHANGE_ID}/messages" | jq '.[] | {id, author, date, message, _revision_number}'
```

#### Change diff

```bash
# Get raw patch (unified diff format)
gerrit_api "/changes/${CHANGE_ID}/revisions/current/patch?raw" | cat

# Get structured diff for a specific file
FILE_PATH="src/example.go"  # URL encode if needed
gerrit_api "/changes/${CHANGE_ID}/revisions/current/files/${FILE_PATH}/diff" | jq '.'
```

#### Change commit message

```bash
# Get commit message
gerrit_api "/changes/${CHANGE_ID}/message" | jq '{subject, full_message, footers}'
```

## Change ID Formats

Gerrit supports multiple change ID formats:

- **Full format**: `go~master~I<change-id>` (e.g., `go~master~I8473b95934b5732ac55d26311a706c9c2bde9940`)
- **Change number**: `<number>` (e.g., `3965`)
- **Change-Id only**: `I<change-id>` (if unique)

For the `go` project, you can use:

- Change number: `3965`
- Full format: `go~master~I8473b95934b5732ac55d26311a706c9c2bde9940`

## Error Handling

### Common Errors and Solutions

- **`curl` command not found**: Guide user to install curl
- **Authentication error (401/403)**: This Gerrit server may not allow anonymous API access. Some Gerrit instances require authentication. If you encounter this error, this plugin cannot access that Gerrit instance without authentication. However, `go-review.googlesource.com` currently supports anonymous access.
- **Change not found (404)**: Verify change ID format and that change exists
- **Rate limit (429)**: Anonymous access may have stricter rate limits. Wait and retry with exponential backoff, or reduce the number of requests (limit parameter, fewer concurrent requests)
- **XSSI prefix error**: Ensure `)]}'` is stripped before JSON parsing (handled by `gerrit_api()` function)

### Error Handling Pattern

```bash
# Example: Check for errors after API call
if ! result=$(gerrit_api "/changes/12345/detail"); then
  echo "ERROR: Failed to fetch change details"
  exit 1
fi

# Parse JSON and check for error messages
if echo "$result" | jq -e '.error' >/dev/null 2>&1; then
  error_msg=$(echo "$result" | jq -r '.error')
  echo "ERROR: $error_msg"
  exit 1
fi
```

## Side Effects

- **Network Access**: All API calls require network access to Gerrit server
- **Anonymous Access**: Uses anonymous API access (no authentication required). Some Gerrit instances may not support anonymous access and may return 401/403 errors. Rate limits may be stricter for anonymous access.
- **No Local File Changes**: This reference does not modify local files
