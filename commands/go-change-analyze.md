---
description: Analyze a single golang/go Change (CL) and extract Go philosophy
allowed-tools: Bash, WebFetch, Read
argument-hint: <change-id>
---

# /go-change-analyze

Analyzes a single Change (CL) from golang/go repository and extracts Go design philosophy insights. This is a **primitive command** focused solely on analysis.

## Command Name

`/go-change-analyze <change-id>`

## One-Line Description

Analyzes a single Change (CL) by fetching its details, comments, and patch, then extracts Go design philosophy and insights.

## Arguments

- `$1` (`change-id`): Change identifier (required)
  - **Type**: string (number or full Change ID)
  - **Required**: Yes
  - **Format**:
    - Change number: `3965`
    - Full format: `go~master~I8473b95934b5732ac55d26311a706c9c2bde9940`
    - Change-Id only: `I8473b95934b5732ac55d26311a706c9c2bde9940` (if unique)
  - **Description**: The Change to analyze

## Output / Side Effects

- **Output**: Displays a Markdown-formatted analysis report in chat containing:
  - Change summary
  - Background and motivation
  - Review discussion points
  - Go philosophy alignment
  - Category classification
  - Key learnings
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

- A detailed analysis report of the specified Change is displayed
- Local files/settings are not modified

## Execution Steps

### 1. Environment Check

```bash
# Check curl command
if ! command -v curl >/dev/null 2>&1; then
  echo "ERROR: curl command not found. Please install curl."
  exit 1
fi

# Check jq command
if ! command -v jq >/dev/null 2>&1; then
  echo "ERROR: jq command not found. Please install jq."
  exit 1
fi

# Check sed command
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

# Check change-id argument
if [ -z "$1" ]; then
  echo "ERROR: Change ID is required"
  echo "Usage: /go-change-analyze <change-id>"
  exit 1
fi

CHANGE_ID="$1"

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

### 2. Fetch Change Details

```bash
# Get change details with labels, messages, and reviewer updates
gerrit_api "/changes/${CHANGE_ID}/detail?o=LABELS&o=DETAILED_LABELS&o=MESSAGES&o=REVIEWER_UPDATES&o=CURRENT_REVISION&o=CURRENT_COMMIT&o=CURRENT_FILES" | jq '.'

# Get all comments across revisions
gerrit_api "/changes/${CHANGE_ID}/comments" | jq '.'

# Get change messages (cover messages and system messages)
gerrit_api "/changes/${CHANGE_ID}/messages" | jq '.[] | {id, author, date, message, _revision_number}'

# Get diff (raw patch)
gerrit_api "/changes/${CHANGE_ID}/revisions/current/patch?raw" | cat

# Get commit message
gerrit_api "/changes/${CHANGE_ID}/message" | jq '{subject, full_message, footers}'
```

### 3. Perform Analysis

Analyze fetched information for:

1. **Change Background & Motivation**: Extract from commit message and change messages
2. **Review Discussion Points**: Extract from review comments and change messages
3. **Go Design Philosophy Alignment**: Map to Go principles (Simplicity, Explicitness, Orthogonality, Practicality)
4. **Category Classification**: Classify into categories (error-handling, performance, api-design, testing, documentation, tooling, runtime, compiler)

### 4. Generate Analysis Report

Create report in the following format:

```markdown
# Change #<number> Analysis: <subject>

## Summary

<Change purpose and summary>

## Background

<Why this change was needed>

## Discussion Points

<Key points discussed in review>
- Patch set evolution: <how change evolved>
- Reviewer feedback: <what reviewers emphasized>
- Label progression: <how Code-Review/Verified evolved>

## Go Philosophy Alignment

<Go design principles this change embodies>

## Category

<category>

## Key Learnings

<Practical insights from this Change>
```

## Usage Examples

```bash
# Analyze Change #3965
/go-change-analyze 3965

# Analyze Change using full Change ID
/go-change-analyze go~master~I8473b95934b5732ac55d26311a706c9c2bde9940
```

## Related Commands

- `/go-changes-fetch`: Fetches Change list
- `/go-catchup`: Orchestrates fetch + analyze + report workflow for multiple Changes

## Notes

- This command focuses on **analyzing a single Change**. Use `/go-changes-fetch` to get a list first.
- Changes are fetched from Gerrit (go-review.googlesource.com), not GitHub
- Analysis includes patch set evolution, which shows how changes were refined through review
- See `skills/go-pr-analyzer/SKILL.md` for detailed analysis perspectives and patterns
