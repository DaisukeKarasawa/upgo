---
name: go-pr-analyzer
description: |
  Analyzes golang/go PRs and extracts Go design philosophy and insights.
  Use when user says "analyze PR", "explain Go philosophy", "why was this change made", etc.
allowed-tools:
  - Bash
  - WebFetch
  - Read
---

# Go PR Analyzer

Analyzes golang/go PRs to extract Go design philosophy, best practices, and insights.

## Analysis Perspectives

### 1. Change Background & Motivation

Extract from PR description and comments:

- Why was this change needed?
- What problem does it solve?
- Why was this approach chosen over alternatives?

### 2. Review Discussion Points

Extract from review comments:

- What points were discussed?
- What was modified before approval?
- What quality criteria did reviewers emphasize?

### 3. Go Design Philosophy Alignment

Map changes to Go design principles:

- **Simplicity**: Does it reduce complexity?
- **Explicitness**: Does it avoid implicit behavior?
- **Orthogonality**: Is it composable with independent features?
- **Practicality**: Is it based on real use cases?

### 4. Category Classification

Classify PRs into categories:

| Category | Description |
|----------|-------------|
| `error-handling` | Error handling improvements |
| `performance` | Performance optimization |
| `api-design` | API design changes |
| `testing` | Test additions/improvements |
| `documentation` | Documentation updates |
| `tooling` | Toolchain improvements |
| `runtime` | Runtime changes |
| `compiler` | Compiler improvements |

## Analysis Steps

### Step 1: Fetch PR Information

```bash
# Get PR details
gh pr view <PR_NUMBER> --repo golang/go --json number,title,body,state,author,labels,comments,reviews

# Get comments and discussions
gh pr view <PR_NUMBER> --repo golang/go --comments

# Get diff
gh pr diff <PR_NUMBER> --repo golang/go
```

### Step 2: Perform Analysis

Analyze fetched information for:

1. **Issue relation**: Identify related issues from `Fixes #XXXX`
2. **Change scope**: Lines added/removed, files affected
3. **Review rounds**: Number of revisions before approval
4. **Discussion depth**: Comment count, reply threads

### Step 3: Extract Insights

Articulate Go philosophy from analysis:

```markdown
## Go Philosophy Insights

This PR embodies "<Go principle>".

**Specifically:**
- Why <change> aligns with <principle>
- Why <point> was emphasized in review

**Learnings:**
- <practical advice>
```

## Output Format

```markdown
# PR #<number> Analysis: <title>

## Summary
<PR purpose and change summary>

## Background
<Why this change was needed>

## Discussion Points
<Key points discussed in review>

## Go Philosophy Alignment
<Go design principles this change embodies>

## Category
<category>

## Key Learnings
<Practical insights from this PR>
```
