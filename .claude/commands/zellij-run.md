---
description: Run a command in a separate zellij pane
allowed-tools: Bash
argument-hint: <command>
---

Run any command in a separate pane.

## Arguments

$ARGUMENTS

## Execution

```bash
# Check zellij session
if [ -z "$ZELLIJ" ]; then
  echo "ERROR: Must run inside a zellij session"
  exit 1
fi

# Check arguments
if [ -z "$ARGUMENTS" ]; then
  echo "ERROR: Please specify a command to run"
  exit 1
fi

# Create pane below
zellij action new-pane --direction down --name "runner"

# Run command
zellij action write-chars "$ARGUMENTS"
zellij action write 10

# Return to main
zellij action move-focus up
```

## Completion

Report: "Command is running in the bottom pane."
