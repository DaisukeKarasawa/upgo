---
description: Fetch golang/go Changes (CLs) from Gerrit
allowed-tools: Bash
argument-hint: [days] [status] [limit]
---

# /go-changes-fetch

Fetches Change (CL) information from golang/go repository via Gerrit REST API. This is a **primitive command** focused solely on data fetching.

## Command Name

`/go-changes-fetch [days] [status] [limit]`

## One-Line Description

Fetches Change (CL) list from Gerrit and outputs JSON. Does not perform analysis or generate reports.

## Arguments

- `$1` (`days`): Number of days to look back (optional)

  - **Type**: number
  - **Required**: No
  - **Default**: `30`
  - **Description**: Fetches Changes updated within the last N days (using `-age` operator)

- `$2` (`status`): Change status filter (optional)

  - **Type**: string
  - **Required**: No
  - **Default**: `merged`
  - **Constraints**: `open`, `merged`, `abandoned`, or empty (all statuses)
  - **Description**: Filters Changes by status

- `$3` (`limit`): Maximum number of Changes to fetch (optional)
  - **Type**: number
  - **Required**: No
  - **Default**: `50`
  - **Description**: Limits the number of Changes returned

## Output / Side Effects

- **Output**: JSON array of Change objects (printed to stdout)
  - Each object contains: `_number`, `subject`, `owner`, `submitted`, `updated`, `labels`
- **Side Effects**:
  - Makes API calls to Gerrit (default: `https://go-review.googlesource.com`) over the network
  - Requires authentication via `GERRIT_USER` and `GERRIT_HTTP_PASSWORD` environment variables
  - Does not create or update local files
  - Does not perform analysis or generate reports

## Prerequisites (Required State / Permissions / Files)

- **Required Commands**: `curl`, `jq`, `sed`
- **Environment Variables (Required)**:
  - `GERRIT_USER`: Gerrit username
  - `GERRIT_HTTP_PASSWORD`: Gerrit HTTP password (obtain from `https://go-review.googlesource.com/settings/#HTTPCredentials`)
- **Environment Variables (Optional)**:
  - `GERRIT_BASE_URL`: Gerrit server URL (default: `https://go-review.googlesource.com`)
- **Prerequisites**: Network access to Gerrit must be available

## Expected State After Execution

- JSON output containing Change list is available (can be piped to other commands or saved)
- Local files/settings are not modified

## Execution Steps

### 1. Environment Check

Use the environment check pattern from `go-gerrit-reference` skill. See `skills/go-gerrit-reference/REFERENCE.md` for complete helper function and authentication setup.

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

# Check Gerrit environment variables
if [ -z "$GERRIT_USER" ] || [ -z "$GERRIT_HTTP_PASSWORD" ]; then
  echo "ERROR: GERRIT_USER and GERRIT_HTTP_PASSWORD must be set"
  echo "Visit https://go-review.googlesource.com/settings/#HTTPCredentials to get HTTP password"
  exit 1
fi

# Load gerrit_api() helper function (see skills/go-gerrit-reference/REFERENCE.md)
gerrit_api() {
  local endpoint="$1"
  local base_url="${GERRIT_BASE_URL:-https://go-review.googlesource.com}"
  local raw

  raw="$(curl -fsS -u "${GERRIT_USER}:${GERRIT_HTTP_PASSWORD}" "${base_url}/a${endpoint}")" || return $?
  printf '%s\n' "$raw" | sed "1s/^)]}'//"
}
```

### 2. Fetch Change List

```bash
DAYS="${1:-30}"
STATUS="${2:-merged}"
LIMIT="${3:-50}"

# Build query
QUERY="project:go"
if [ -n "$STATUS" ] && [ "$STATUS" != "all" ]; then
  QUERY="${QUERY}+status:${STATUS}"
fi
QUERY="${QUERY}+-age:${DAYS}d"

# Fetch changes
gerrit_api "/changes/?q=${QUERY}&n=${LIMIT}&o=LABELS&o=DETAILED_ACCOUNTS&o=CURRENT_REVISION&o=CURRENT_COMMIT" | jq '[.[] | {_number, subject, owner, submitted, updated, labels}]'
```

## Usage Examples

```bash
# Fetch merged changes from last 30 days (default)
/go-changes-fetch

# Fetch merged changes from last 7 days
/go-changes-fetch 7

# Fetch open changes from last 14 days
/go-changes-fetch 14 open

# Fetch merged changes from last 30 days, limit 20
/go-changes-fetch 30 merged 20

# Pipe output to file
/go-changes-fetch > changes.json

# Extract change numbers and analyze each
# Note: /go-change-analyze requires change-id as argument, not stdin
/go-changes-fetch | jq -r '.[]._number' | while read id; do /go-change-analyze "$id"; done
```

## Related Commands

- `/go-change-analyze`: Analyzes fetched Change data
- `/go-catchup`: Orchestrates fetch + analyze + report workflow

## Notes

- This command focuses on **fetching only**. Use `/go-change-analyze` for analysis.
- Changes are fetched from Gerrit (go-review.googlesource.com), not GitHub
- Output is JSON, suitable for piping to other commands or saving to files
- For Gerrit API helper function and common patterns, see `skills/go-gerrit-reference/REFERENCE.md`
- See `skills/go-pr-fetcher/SKILL.md` for detailed Gerrit API usage patterns
