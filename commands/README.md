# Slash Commands (Inventory / Unified Format)

Issue: `https://github.com/DaisukeKarasawa/upgo/issues/43`

This directory contains **Slash Commands documentation** (Markdown) for Claude Code plugins.

## Command Inventory

- `/go-catchup [category]`: Fetches and analyzes golang/go Changes (CLs) from the last 30 days and generates a report
  - Definition: `commands/go-catchup.md`

## Unified Format (Template)

When adding a new command, fill in the following required sections in each command's `.md` file:

- **Command Name** (recommended notation)
  - Example: ``/my-command [arg]``
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

## Notes

- The `commands/` directory in this repository is intended to be copied to `~/.claude/commands/` following the installation instructions (see the repository's `README.md` for details).
