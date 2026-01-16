---
description: Catch up on recent golang/go PRs and learn Go philosophy
allowed-tools: Bash, WebFetch, Read
argument-hint: [category]
---

# Go PR Catchup

Fetches and analyzes PRs updated in the last month from golang/go repository to learn Go design philosophy.

## Arguments

- `$1`: Category filter (optional: error-handling, performance, api-design, testing, runtime, compiler)

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
# Calculate date 1 month ago
ONE_MONTH_AGO=$(date -v-1m +%Y-%m-%d 2>/dev/null || date -d "1 month ago" +%Y-%m-%d)

# Fetch PRs updated in the last month
gh pr list --repo golang/go --state merged --search "updated:>=$ONE_MONTH_AGO" --json number,title,author,mergedAt,labels,updatedAt
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
# Catch up on all PRs from the last month
/go-catchup

# Filter by error-handling category
/go-catchup error-handling

# Filter by performance category
/go-catchup performance
```
