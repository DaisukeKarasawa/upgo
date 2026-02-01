# Slash Commands (Inventory / Unified Format)

Issue: `https://github.com/DaisukeKarasawa/upgo/issues/44`

This directory contains **Slash Commands documentation** (Markdown) for Claude Code plugins.

## Design Principles

See `commands/NAMING.md` for naming conventions, responsibility separation, and compatibility guidelines.

## Command Inventory

### Orchestrator Commands (Workflow)

- `/go-catchup [category]`: Fetches and analyzes golang/go Changes (CLs) from the last 30 days and generates a report
  - Definition: `commands/go-catchup.md`
  - **Status**: Active (orchestrates fetch + analyze + report)

### Primitive Commands (Single Purpose)

- `/go-changes-fetch [days] [status] [limit]`: Fetches Change (CL) list from Gerrit and outputs JSON

  - Definition: `commands/go-changes-fetch.md`
  - **Purpose**: Data fetching only (network access)

- `/go-change-analyze <change-id>`: Analyzes a single Change (CL) and extracts Go philosophy insights
  - Definition: `commands/go-change-analyze.md`
  - **Purpose**: Analysis only (requires Change data)

## Unified Format (Template)

When adding a new command, fill in the following required sections in each command's `.md` file:

- **Command Name** (recommended notation)
  - Example: `/my-command [arg]`
- **One-Line Description**
- **Arguments**
  - Type / Optional / Default / Constraints / Description
- **Output / Side Effects**
  - Output (e.g., chat output, generated files, API calls)
  - Side effects (e.g., file updates, network access, external service writes)
- **Prerequisites** (Required State / Permissions / Files)
  - Required commands, environment variables, authentication, network reachability, etc.
- **Expected State After Execution** (What changes and how)
- **Usage Examples** (at least 1)

## Migration Notes

### Command Naming Convention (Issue #44)

Following the naming convention defined in `commands/NAMING.md`:

- **Pattern**: `domain-subject-action` (e.g., `go-changes-fetch`, `go-change-analyze`)
- **Namespacing**: Use `go-` prefix for Go-related commands to avoid collisions
- **Separation**: Primitive commands (`*-fetch`, `*-analyze`) are separated from orchestrator commands (`*-catchup`)

### Breaking Changes

None. Existing `/go-catchup` command remains available and functional.

### New Commands

- `/go-changes-fetch`: Use when you only need to fetch Change data
- `/go-change-analyze`: Use when you want to analyze a specific Change

These can be used independently or combined for custom workflows.

## Notes

- The `commands/` directory in this repository is intended to be copied to `~/.claude/commands/` following the installation instructions (see the repository's `README.md` for details).
- See `commands/NAMING.md` for guidelines on adding new commands.
