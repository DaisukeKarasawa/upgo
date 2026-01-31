---
name: go-pr-fetcher
description: |
  Fetches Change (CL) information from golang/go repository via Gerrit.
  Use when user says "fetch Go changes", "latest CLs from golang/go", "check Go changes", etc.
allowed-tools:
  - Bash
  - WebFetch
---

# Go Change Fetcher

Fetches Change (CL) information from golang/go repository using Gerrit REST API.

## Prerequisites

Environment variables for Gerrit authentication must be set:

- `GERRIT_BASE_URL`: Gerrit server URL (default: `https://go-review.googlesource.com`)
- `GERRIT_USER`: Gerrit username
- `GERRIT_HTTP_PASSWORD`: Gerrit HTTP password (from Settings > HTTP Password)

```bash
echo $GERRIT_BASE_URL
echo $GERRIT_USER
echo $GERRIT_HTTP_PASSWORD
```

If not set, prompt the user to set them. To get HTTP password:

1. Visit [Gerrit HTTP Credentials](https://go-review.googlesource.com/settings/#HTTPCredentials)
2. Generate HTTP password
3. Set it as `GERRIT_HTTP_PASSWORD`

## Helper Function

All Gerrit API responses start with `)]}'` (XSSI protection). Strip it before parsing JSON:

```bash
# Helper function to fetch Gerrit API and strip XSSI prefix
gerrit_api() {
  local endpoint="$1"
  local base_url="${GERRIT_BASE_URL:-https://go-review.googlesource.com}"
  curl -sf -u "${GERRIT_USER}:${GERRIT_HTTP_PASSWORD}" \
    "${base_url}/a${endpoint}" | sed "1s/^)]}'//"
}
```

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

### Change basic info

```bash
# Get change detail with labels, messages, and reviewer updates
# Change ID format: go~master~I<change-id> or just <change-number>
CHANGE_ID="go~master~I<change-id>"  # or use change number: <change-number>
gerrit_api "/changes/${CHANGE_ID}/detail?o=LABELS&o=DETAILED_LABELS&o=MESSAGES&o=REVIEWER_UPDATES" | jq '{_number, subject, status, owner, created, updated, labels, messages}'
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

Gerrit supports multiple change ID formats:

- Full format: `go~master~I<change-id>` (e.g., `go~master~I8473b95934b5732ac55d26311a706c9c2bde9940`)
- Change number: `<number>` (e.g., `3965`)
- Change-Id only: `I<change-id>` (if unique)

For the `go` project, you can use:

- Change number: `3965`
- Full format: `go~master~I8473b95934b5732ac55d26311a706c9c2bde9940`

## Error Handling

- `curl` command not found: Guide user to install curl
- Authentication error (401): Guide user to set `GERRIT_USER` and `GERRIT_HTTP_PASSWORD`
- Change not found (404): Verify change ID format and that change exists
- Rate limit (429): Advise to wait and retry with exponential backoff
- XSSI prefix error: Ensure `)]}'` is stripped before JSON parsing
