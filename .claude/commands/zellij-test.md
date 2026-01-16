---
description: Run tests in a separate zellij pane
allowed-tools: Bash
argument-hint: [test-pattern]
---

Run tests in a separate pane.

## Steps

1. Check zellij session
2. Create new pane on the right
3. Run tests
4. Return to main pane

## Execution

```bash
# 1. Check zellij session
if [ -z "$ZELLIJ" ]; then
  echo "ERROR: Must run inside a zellij session"
  exit 1
fi

# 2. Create test pane
zellij action new-pane --direction right --name "test-runner"

# 3. Run tests
TEST_PATTERN="$ARGUMENTS"
if [ -n "$TEST_PATTERN" ]; then
  zellij action write-chars "go test -v -run '$TEST_PATTERN' ./..."
else
  zellij action write-chars "go test -v ./..."
fi
zellij action write 10

# 4. Return to main
zellij action move-focus left
```

## Completion

Report: "Tests are running in the right pane. Check results there."
