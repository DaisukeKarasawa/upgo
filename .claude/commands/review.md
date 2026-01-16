---
description: Run CodeRabbit code review on changes
allowed-tools: Bash(coderabbit:*), Bash(cr:*), Bash(git:*), Bash(make:*), Read, Edit, Write, Grep, Glob
argument-hint: [uncommitted|committed|all] [--fix]
---

# CodeRabbit Code Review

## Current Status

Git status: !`git status --short | head -10`
Auth status: !`coderabbit auth status 2>&1 | grep -E "(Authentication|logged)" | head -1`

## Instructions

Run CodeRabbit code review based on arguments:

- **No argument or "uncommitted"**: Review uncommitted changes
- **"committed"**: Review committed changes only
- **"all"**: Review all changes
- **"--fix"**: Automatically fix detected issues

## Arguments

$ARGUMENTS

## Review Commands

```bash
# For AI-optimized output (recommended)
coderabbit --prompt-only --type uncommitted

# For detailed human-readable output
coderabbit --plain --type uncommitted

# Compare against specific branch
coderabbit --prompt-only --base main
```

## Workflow

1. Run CodeRabbit analysis
2. Parse and categorize issues
3. Fix critical issues automatically (if --fix specified)
4. Run tests to verify fixes
5. Report summary

## Note

If not authenticated, run:
```bash
coderabbit auth login
```
