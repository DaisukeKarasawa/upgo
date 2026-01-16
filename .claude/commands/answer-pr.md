---
description: Review and respond to PR comments
allowed-tools: Bash(gh:*), Read, Write, Edit, Grep, Glob
argument-hint: <pr-number>
---

# Answer PR #$ARGUMENTS Comments

## Instructions

1. Fetch the latest review comments on the PR
2. Understand each comment and its context
3. Make necessary code changes
4. Respond to comments appropriately

## PR Information

PR details: !`gh pr view $ARGUMENTS --json title,state,reviews,comments 2>/dev/null || echo "PR not found or gh CLI not configured"`

## Review Comments

!`gh pr view $ARGUMENTS --comments 2>/dev/null | tail -50 || echo "Could not fetch comments"`

## Workflow

1. **Read**: Understand each review comment
2. **Analyze**: Determine what changes are needed
3. **Implement**: Make the requested changes (following TDD)
4. **Test**: Ensure all tests pass
5. **Commit**: Use appropriate gitmoji
6. **Reply**: Summarize changes made in response

## Response Guidelines

- Address each comment specifically
- Explain your changes or reasoning
- Ask for clarification if needed
- Be respectful and constructive
