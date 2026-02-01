---
name: go-pr-analyzer
description: |
  Analyzes golang/go Changes (CLs) and extracts Go design philosophy and insights.
  Use when user says "analyze change", "explain Go philosophy", "why was this change made", etc.
allowed-tools:
  - Bash
---

# Go Change Analyzer

Analyzes golang/go Changes (CLs) to extract Go design philosophy, best practices, and insights.

## Side Effects

- **Network Access**: Makes API calls to Gerrit server (default: `https://go-review.googlesource.com`) to fetch Change data
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

## Quick Start

Use the `gerrit_api()` helper function from `go-gerrit-reference` skill to fetch Change data. See [go-gerrit-reference/REFERENCE.md](../go-gerrit-reference/REFERENCE.md) for complete helper function and authentication setup.

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

- `error-handling`: Error handling improvements
- `performance`: Performance optimization
- `api-design`: API design changes
- `testing`: Test additions/improvements
- `documentation`: Documentation updates
- `tooling`: Toolchain improvements
- `runtime`: Runtime changes
- `compiler`: Compiler improvements

## Analysis Steps

### Step 1: Fetch Change Information

Use the `gerrit_api()` helper function from `go-gerrit-reference` skill:

```bash
# Load helper function from reference (see go-gerrit-reference/REFERENCE.md)
CHANGE_ID="<change-number>"  # e.g., 3965 (numeric) or go~master~I8473b95934b5732ac55d26311a706c9c2bde9940 (full format)

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

For complete API reference, see [go-gerrit-reference/REFERENCE.md](../go-gerrit-reference/REFERENCE.md).

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
