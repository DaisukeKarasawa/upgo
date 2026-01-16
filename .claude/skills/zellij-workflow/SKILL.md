---
name: zellij-workflow
description: |
  Zellij terminal multiplexer workflow skill for parallel development.
  Use when user says "run in separate pane", "parallel", "zellij", "multi-pane", etc.
allowed-tools:
  - Bash
  - Read
---

# Zellij Workflow for Claude Code

Skill for parallelizing Claude Code work using zellij.

## Important: Check Zellij Session

First, verify you're inside a zellij session:

```bash
echo $ZELLIJ
```

If empty, prompt user to start a session with `zellij`.

## Command Reference

### Create New Pane

```bash
# New pane on the right
zellij action new-pane --direction right --name "pane-name"

# New pane below
zellij action new-pane --direction down --name "pane-name"

# Floating pane (popup)
zellij action new-pane --floating --name "pane-name"
```

### Send Commands

```bash
# Send characters
zellij action write-chars "command string"

# Send Enter key (execute)
zellij action write 10
```

### Move Focus

```bash
zellij action move-focus left
zellij action move-focus right
zellij action move-focus up
zellij action move-focus down
```

## Common Patterns

### Pattern 1: Run Tests in Separate Pane

```bash
# 1. Create test pane on the right
zellij action new-pane --direction right --name "tests"

# 2. Send test command
zellij action write-chars "go test -v ./..."
zellij action write 10

# 3. Return to main pane
zellij action move-focus left
```

### Pattern 2: Run Server in Separate Pane

```bash
# 1. Create server pane below
zellij action new-pane --direction down --name "server"

# 2. Start server
zellij action write-chars "go run cmd/server/main.go"
zellij action write 10

# 3. Return to main pane
zellij action move-focus up
```

### Pattern 3: Launch Claude Code Agent in Separate Pane

```bash
# 1. Create agent pane on the right
zellij action new-pane --direction right --name "agent"

# 2. Launch Claude Code with task
zellij action write-chars "claude --print 'Task instructions here'"
zellij action write 10

# 3. Return to main pane
zellij action move-focus left
```

### Pattern 4: 2x2 Grid for Parallel Tasks

```bash
# Create grid
zellij action new-pane --direction right --name "task-2"
zellij action move-focus left
zellij action new-pane --direction down --name "task-3"
zellij action move-focus right
zellij action new-pane --direction down --name "task-4"

# Return to main
zellij action move-focus up
zellij action move-focus left
```

## Notes

1. **Don't forget Enter**: `write-chars` alone doesn't execute. Always follow with `write 10`.
2. **Return focus**: Return to main pane after setup.
3. **Name panes**: Use `--name` for easy identification.
