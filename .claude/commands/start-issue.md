---
description: Start working on a GitHub issue
allowed-tools: Bash(gh:*), Read, Write, Edit, Grep, Glob, WebFetch
argument-hint: <issue-number>
---

# Start Working on Issue #$ARGUMENTS

## Instructions

1. Fetch the issue details from GitHub
2. Understand the requirements and context
3. Create a plan for implementation following TDD
4. Create a feature branch if needed

## Issue Information

Fetch issue: !`gh issue view $ARGUMENTS --json title,body,labels,assignees 2>/dev/null || echo "Issue not found or gh CLI not configured"`

## Workflow

1. **Understand**: Read the issue carefully
2. **Plan**: Break down into small, testable tasks
3. **Branch**: Create feature branch (`git checkout -b feature/issue-$ARGUMENTS`)
4. **TDD**: Follow Red-Green-Refactor cycle
5. **Commit**: Use gitmoji conventions
6. **PR**: Create pull request when ready

## TDD Approach

For each task:
1. Write a failing test first
2. Implement minimum code to pass
3. Refactor while keeping tests green
4. Commit (separate structural and behavioral changes)
