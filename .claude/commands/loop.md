---
description: Launch Claude Code pair programming partner in separate pane
allowed-tools: Bash
argument-hint: [initial-prompt]
---

Launch Claude Code in a separate pane as a pair programming partner.

## Initial Prompt

$ARGUMENTS

## Execution

```bash
# Check zellij session
if [ -z "$ZELLIJ" ]; then
  echo "ERROR: Must run inside a zellij session"
  exit 1
fi

# Create pane on the right
zellij action new-pane --direction right --name "pair-programmer"

# Launch Claude Code
INITIAL_PROMPT="$ARGUMENTS"
if [ -n "$INITIAL_PROMPT" ]; then
  zellij action write-chars "claude --print 'Act as pair programming partner. $INITIAL_PROMPT'"
else
  zellij action write-chars "claude --print 'Act as pair programming partner.'"
fi
zellij action write 10

# Return to main
zellij action move-focus left
```

## Usage Patterns

### Test Writer

```bash
/loop Handle test creation
```

### Code Reviewer

```bash
/loop Handle code review
```

### Documentation Writer

```bash
/loop Handle documentation
```

## Completion

Report: "Pair programming partner launched in the right pane."
