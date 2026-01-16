---
description: Analyze task and run in parallel across multiple zellij panes
allowed-tools: Bash, Read, Write
argument-hint: <task-description>
---

Analyze task and split into subtasks for parallel execution across multiple panes.

## Task

$ARGUMENTS

## Steps

### 1. Task Analysis

Analyze user's task and split into 2-4 independent subtasks.

Examples:
- "Implement API and write tests" → Task 1: API implementation, Task 2: Test creation
- "Review code" → Task 1: Security, Task 2: Performance, Task 3: Style

### 2. Pane Layout

Choose layout based on number of subtasks:

**2 tasks (side by side):**
```bash
zellij action new-pane --direction right --name "task-2"
zellij action move-focus left
```

**3 tasks (T-shape):**
```bash
zellij action new-pane --direction right --name "task-2"
zellij action move-focus left
zellij action new-pane --direction down --name "task-3"
zellij action move-focus up
```

**4 tasks (grid):**
```bash
zellij action new-pane --direction right --name "task-2"
zellij action move-focus left
zellij action new-pane --direction down --name "task-3"
zellij action move-focus right
zellij action new-pane --direction down --name "task-4"
zellij action move-focus up
zellij action move-focus left
```

### 3. Launch Agents

Launch Claude Code in each pane with assigned task:

```bash
# Send task to pane 2
zellij action move-focus right
zellij action write-chars "claude --print 'Subtask 2 instructions'"
zellij action write 10
zellij action move-focus left
```

### 4. Execute Main Task

Execute task 1 directly in main pane.

## Completion

Report: "Split task into N subtasks and started parallel execution in each pane."
