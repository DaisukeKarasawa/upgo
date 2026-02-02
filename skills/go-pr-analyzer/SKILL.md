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

- **Network Access**: Makes API calls to Gerrit server (default: `https://go-review.googlesource.com`) to fetch Change data using anonymous access
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

Use the `gerrit_api()` helper function from `go-gerrit-reference` skill to fetch Change data. See [go-gerrit-reference/REFERENCE.md](../go-gerrit-reference/REFERENCE.md) for complete helper function.

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

#### 2.1. Review Comment Classification

Classify review comments by type:

- **Code Quality**: Readability, maintainability, performance concerns
- **Go Conventions**: Go idioms, naming, error handling patterns
- **Design Decisions**: Approach selection, design pattern discussions
- **Testing**: Test coverage, test quality, edge cases
- **Documentation**: Comments, doc strings, API documentation
- **Must-Fix**: Required changes before approval
- **Suggestion**: Optional improvements
- **Question**: Clarifications or discussions

#### 2.2. Comment Thread Analysis

Map comment threads using `in_reply_to` field:

- **Thread Structure**: Follow `in_reply_to` to build thread trees
- **Issue → Response → Resolution**: Map the flow from initial concern to final resolution
- **Reviewer Engagement**: Track who responded and when
- **Discussion Depth**: Count thread depth and back-and-forth exchanges

#### 2.3. Patch Set Evolution Tracking

Track how code evolved through review:

- **Revision Comparison**: Compare consecutive patch sets to identify changes
- **Review Comment Mapping**: Map specific review comments to code changes in subsequent patch sets
- **Improvement Patterns**: Identify common patterns of improvement (e.g., error handling, naming, test coverage)
- **Iteration Count**: Track number of iterations before approval

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

#### Step 2.1: Review Comment Analysis (Deep Dive)

For detailed review analysis:

1. **Fetch All Comments**:

   ```bash
   # Get all comments with thread structure
   gerrit_api "/changes/${CHANGE_ID}/comments" | jq '.'
   ```

2. **Build Thread Trees**:

   - Parse `in_reply_to` field to build comment thread hierarchy
   - Group comments by file and line number
   - Identify root comments (no `in_reply_to`) vs replies

3. **Classify Comments**:

   - Categorize by type (code quality, Go conventions, design, testing, documentation)
   - Identify must-fix vs suggestions vs questions
   - Extract file:line references for inline comments

4. **Map to Code Changes**:
   - Compare patch sets to see how comments were addressed
   - Track which comments led to code changes
   - Identify patterns in how code improved

#### Step 2.2: Patch Set Evolution Analysis

For tracking code evolution:

1. **Get All Revisions**:

   ```bash
   # Get all revisions for comparison
   gerrit_api "/changes/${CHANGE_ID}/detail?o=ALL_REVISIONS" | jq '.revisions'
   ```

2. **Compare Consecutive Revisions**:

   ```bash
   # Compare two revisions
   REV1="<revision-hash-1>"
   REV2="<revision-hash-2>"
   gerrit_api "/changes/${CHANGE_ID}/revisions/${REV2}/files?base=${REV1}" | jq '.'
   ```

3. **Map Review Comments to Changes**:
   - Identify which review comments correspond to code changes
   - Track the evolution: initial code → review comment → revised code
   - Extract improvement patterns

#### Step 2.3: Review Metrics Extraction

Extract quantitative metrics:

- **Patch Set Count**: Total number of revisions
- **Comment Count**: Total comments (inline + cover messages)
- **Thread Count**: Number of distinct discussion threads
- **Reviewer Count**: Number of unique reviewers
- **Review Duration**: Time from first patch set to approval
- **Label Changes**: Number of Code-Review/Verified label changes

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

### Standard Format

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

### Deep Dive Format (Review Analysis)

For detailed review analysis, use this extended format:

```markdown
# Change #<number> Analysis: <subject>

## Summary

<Change purpose and summary>

## Background

<Why this change was needed>

## Review Process

### Reviewers

- <reviewer-name-1>: <role/contribution>
- <reviewer-name-2>: <role/contribution>

### Review Rounds

<number> patch sets, showing iterative refinement

### Key Review Comments

#### Inline Comments

- `<file>:<line>` - `<reviewer-name>`: "<comment-content>"

  - **Type**: <code-quality|go-conventions|design|testing|documentation>
  - **Priority**: <must-fix|suggestion|question>
  - **Addressed**: <how it was addressed in subsequent patch sets>

- `<file>:<line>` - `<reviewer-name>`: "<comment-content>"
  ...

#### Cover Letter Comments

- `<reviewer-name>`: "<comment-content>"
  - **Discussion**: <thread discussion if applicable>
  - **Resolution**: <how it was resolved>

### Comment Threads

#### Thread 1: <topic>

- **Root**: `<reviewer-name>`: "<initial-concern>"
- **Response**: `<author-name>`: "<response>"
- **Follow-up**: `<reviewer-name>`: "<follow-up>"
- **Resolution**: <final resolution>

### Code Evolution

#### PS1 → PS2

- **Changes**: <what changed>
- **Triggered by**: <which review comments>
- **Improvement**: <how code improved>

#### PS2 → PS3

- **Changes**: <what changed>
- **Triggered by**: <which review comments>
- **Improvement**: <how code improved>

### Review Metrics

- **Patch Sets**: <count>
- **Total Comments**: <count>
- **Threads**: <count>
- **Reviewers**: <count>
- **Review Duration**: <duration>
- **Label Evolution**: <Code-Review/Verified changes>

## Go Philosophy Alignment

<Go design principles this change embodies>
- **Simplicity**: <how simplicity was emphasized>
- **Explicitness**: <how explicitness was emphasized>
- **Orthogonality**: <how orthogonality was emphasized>
- **Practicality**: <how practicality was emphasized>

## Category

<primary-category> (tags: <tag1>, <tag2>)

## Key Learnings

<Practical insights from review culture>
- <learning-1>: <explanation>
- <learning-2>: <explanation>
```
