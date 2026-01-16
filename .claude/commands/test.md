---
description: Run tests with various options
allowed-tools: Bash(make:*), Bash(go test:*), Bash(npm test:*), Read
argument-hint: [all|backend|frontend|bench|perf]
---

# Run Tests

## Current Test Status

Backend tests: !`make test 2>&1 | tail -10`

## Instructions

Run tests based on the argument:

- **No argument or "all"**: Run all tests (`make test-all`)
- **"backend"**: Run Go tests only (`make test`)
- **"frontend"**: Run frontend tests (`cd web && npm test`)
- **"bench"**: Run benchmarks (`make bench`)
- **"perf"**: Run performance tests (`make perf-test`)

## Argument

$ARGUMENTS

## TDD Reminder

Remember the TDD cycle:
1. **Red**: Write failing test first
2. **Green**: Minimum code to pass
3. **Refactor**: Improve while green

Report test results clearly, highlighting any failures.
