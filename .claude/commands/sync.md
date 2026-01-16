---
description: Sync PR data from GitHub
allowed-tools: Bash(curl:*), Bash(make:*), Read
argument-hint: [pr-number (optional)]
---

# Sync PR Data

## Instructions

Sync PR data from GitHub:

- **No argument**: Full sync of all PRs
- **PR number**: Sync specific PR only

## Argument

$ARGUMENTS

## Commands

For full sync:
```bash
curl -X POST http://localhost:8081/api/v1/sync
```

For specific PR:
```bash
curl -X POST http://localhost:8081/api/v1/prs/{id}/sync
```

## Notes

- Requires backend server to be running
- Syncs: PR details, comments, diffs, and generates summaries
- Uses Ollama for Japanese summarization
