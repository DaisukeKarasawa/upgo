---
description: Catch up on recent golang/go PRs and learn Go philosophy
allowed-tools: Bash, WebFetch, Read
argument-hint: [count] [category]
---

# Go PR Catchup

Fetches and analyzes recent PRs from golang/go repository to learn Go design philosophy.

## Arguments

- `$1`: Number of PRs to fetch (default: 10)
- `$2`: Category filter (optional: error-handling, performance, api-design, testing, runtime, compiler)

## Execution Steps

### 1. Environment Check

```bash
# Check GitHub CLI
which gh || echo "ERROR: gh command not found. Please install GitHub CLI."

# Check authentication
gh auth status
```

### 2. Fetch PR List

```bash
# Fetch recent merged PRs
LIMIT="${1:-10}"
gh pr list --repo golang/go --state merged --limit $LIMIT --json number,title,author,mergedAt,labels
```

### 3. Analyze Each PR

For each fetched PR:

1. Get PR details
2. Review comments and discussions
3. Analyze changes
4. Extract Go philosophy insights

### 4. Generate Report

Create report in the following format:

```markdown
# Go PR Catchup Report

**Period**: <oldest_date> - <newest_date>
**Count**: <count> PRs

## Summary

### By Category
- error-handling: X PRs
- performance: Y PRs
- ...

### Notable PRs
1. PR #XXXX: <title> - <one-line description>
2. PR #YYYY: <title> - <one-line description>

---

## Detailed Analysis

### PR #XXXX: <title>

**Summary**: <summary>

**Go Philosophy**: <learnings>

---

## Key Takeaways

<Overall Go philosophy and best practices learned>
```

## Usage Examples

```bash
# Catch up on 10 recent PRs
/go-catchup

# Catch up on 20 PRs
/go-catchup 20

# Filter by error-handling category
/go-catchup 10 error-handling
```
