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

- **Network Access**: Makes API calls to Gerrit server (default: `https://go-review.googlesource.com`)
- **Authentication Required**: Requires `GERRIT_USER` and `GERRIT_HTTP_PASSWORD` environment variables
- **No Local File Changes**: Does not create or modify local files

## Prerequisites

### Required Commands

- `curl`: HTTP client for API requests
- `jq`: JSON processor for parsing responses
- `sed`: Text processing for XSSI prefix removal

### Required Environment Variables

- `GERRIT_USER`: Gerrit username
- `GERRIT_HTTP_PASSWORD`: Gerrit HTTP password (obtain from [Gerrit HTTP Credentials](https://go-review.googlesource.com/settings/#HTTPCredentials))

### Optional Environment Variables

- `GERRIT_BASE_URL`: Gerrit server URL (default: `https://go-review.googlesource.com`)

If environment variables are not set, prompt the user to set them.

## Quick Start

Use the `gerrit_api()` helper function from `go-gerrit-reference` skill. See [go-gerrit-reference/REFERENCE.md](../go-gerrit-reference/REFERENCE.md) for complete helper function and authentication setup.

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
