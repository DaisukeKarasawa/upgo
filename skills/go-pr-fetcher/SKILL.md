---
name: go-pr-fetcher
description: |
  Fetches PR information from golang/go repository.
  Use when user says "fetch Go PRs", "latest PRs from golang/go", "check Go changes", etc.
allowed-tools:
  - Bash
  - WebFetch
---

# Go PR Fetcher

Fetches PR information from golang/go repository using GitHub API.

## Prerequisites

Environment variable `GITHUB_TOKEN` must be set.

```bash
echo $GITHUB_TOKEN
```

If not set, prompt the user to set it.

## Fetching PR List

### Get recent PRs (default: 30)

```bash
gh pr list --repo golang/go --state all --limit 30 --json number,title,state,author,createdAt,updatedAt,labels
```

### Get merged PRs only

```bash
gh pr list --repo golang/go --state merged --limit 30 --json number,title,author,mergedAt,labels
```

### Get PRs from specific period

```bash
# Last 7 days
gh pr list --repo golang/go --state all --search "updated:>=$(date -v-7d +%Y-%m-%d 2>/dev/null || date -d '7 days ago' +%Y-%m-%d)" --limit 50 --json number,title,state,author,updatedAt
```

## Fetching Individual PR Details

### PR basic info

```bash
gh pr view <PR_NUMBER> --repo golang/go --json number,title,body,state,author,labels,comments,reviews
```

### PR comments and discussions

```bash
gh pr view <PR_NUMBER> --repo golang/go --comments
```

### PR diff

```bash
gh pr diff <PR_NUMBER> --repo golang/go
```

## Output Format

Format fetched PR information as follows:

```markdown
## PR #<number>: <title>

**State**: <state> | **Author**: <author> | **Updated**: <updatedAt>

**Labels**: <labels>

### Summary
<body summary>

### Changed Files
- <file1>
- <file2>
```

## Error Handling

- `gh` command not found: Guide user to install GitHub CLI
- Authentication error: Guide user to run `gh auth login`
- Rate limit: Advise to wait and retry
