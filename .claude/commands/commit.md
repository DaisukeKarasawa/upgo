---
description: Create a git commit with gitmoji following TDD principles
allowed-tools: Bash(git status:*), Bash(git diff:*), Bash(git add:*), Bash(git commit:*), Bash(git log:*), Read, Grep, Glob
argument-hint: [message (optional)]
---

# Git Commit with Gitmoji

## Pre-commit Checklist

Before committing, verify:
1. All tests pass: !`make test 2>&1 | tail -5`
2. Current git status: !`git status --short`

## Instructions

1. **Verify tests pass** - Do not commit if tests are failing
2. **Review changes** - Understand what is being committed
3. **Classify the change type**:
   - Structural (refactoring only) or Behavioral (feature/fix)
4. **Select appropriate gitmoji**:
   - `:sparkles:` (✨) New feature
   - `:bug:` (🐛) Bug fix
   - `:recycle:` (♻️) Refactor
   - `:white_check_mark:` (✅) Add/update tests
   - `:memo:` (📝) Documentation
   - `:art:` (🎨) Improve structure/format
   - `:zap:` (⚡) Performance improvement
   - `:fire:` (🔥) Remove code/files
   - `:lipstick:` (💄) UI/style updates
   - `:construction:` (🚧) Work in progress

5. **Create commit** with format: `<gitmoji> <message>`

## User Request

$ARGUMENTS

## Guidelines

- Keep commits small and focused (single logical change)
- Separate structural and behavioral changes
- Write clear, concise commit messages
- Include Co-Authored-By for Claude
