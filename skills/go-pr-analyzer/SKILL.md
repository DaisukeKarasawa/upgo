---
name: go-pr-analyzer
description: |
  Analyzes golang/go Changes (CLs) and extracts Go design philosophy and insights.
  Use when user says "analyze change", "explain Go philosophy", "why was this change made", etc.
allowed-tools:
  - Bash
  - WebFetch
  - Read
---

# Go Change Analyzer

Analyzes golang/go Changes (CLs) to extract Go design philosophy, best practices, and insights.

## Analysis Perspectives

### 1. Change Background & Motivation

Extract from commit message and change messages:

- Why was this change needed?
- What problem does it solve?
- Why was this approach chosen over alternatives?

### 2. Review Discussion Points

Extract from review comments and change messages:

- What points were discussed?
- What was modified before approval (across patch sets)?
- What quality criteria did reviewers emphasize?
- How did labels (Code-Review, Verified) evolve?

### 3. Go Design Philosophy Alignment

Map changes to Go design principles:

- **Simplicity**: Does it reduce complexity?
- **Explicitness**: Does it avoid implicit behavior?
- **Orthogonality**: Is it composable with independent features?
- **Practicality**: Is it based on real use cases?

### 4. Category Classification

Classify Changes into categories:

| Category         | Description                 |
| ---------------- | --------------------------- |
| `error-handling` | Error handling improvements |
| `performance`    | Performance optimization    |
| `api-design`     | API design changes          |
| `testing`        | Test additions/improvements |
| `documentation`  | Documentation updates       |
| `tooling`        | Toolchain improvements      |
| `runtime`        | Runtime changes             |
| `compiler`       | Compiler improvements       |

## Analysis Steps

### Step 1: Fetch Change Information

Use the `gerrit_api` helper function from `go-pr-fetcher` skill:

```bash
# Helper function (same as in go-pr-fetcher)
gerrit_api() {
  local endpoint="$1"
  local base_url="${GERRIT_BASE_URL:-https://go-review.googlesource.com}"
  curl -sf -u "${GERRIT_USER}:${GERRIT_HTTP_PASSWORD}" \
    "${base_url}/a${endpoint}" | sed "1s/^)]}'//"
}

CHANGE_ID="<change-number>"  # e.g., 3965 or go~master~3965

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

### Step 2: Perform Analysis

Analyze fetched information for:

1. **Issue relation**: Identify related issues from `Fixes #XXXX` in commit message footers
2. **Change scope**: Lines added/removed (`insertions`, `deletions`), files affected (`files` field)
3. **Review rounds**: Number of patch sets (`current_revision_number`) before approval
4. **Discussion depth**: Comment count, reply threads (`in_reply_to` field in comments)
5. **Label evolution**: How Code-Review and Verified labels changed across patch sets
6. **Reviewer engagement**: Who reviewed, when, and what they emphasized

### Step 3: Extract Insights

Articulate Go philosophy from analysis:

```markdown
## Go Philosophy Insights

This Change embodies "<Go principle>".

**Specifically:**

- Why <change> aligns with <principle>
- Why <point> was emphasized in review
- How patch set evolution reflects Go's iterative review culture

**Learnings:**

- <practical advice>
```

## Output Format

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
