# Slash Commands Design Guide (Naming / Responsibility Separation / Compatibility)

This document provides guidelines for adding Slash Commands under `commands/` to **prevent naming collisions** and maintain **one command = one purpose** principle.  
Issue: `https://github.com/DaisukeKarasawa/upgo/issues/44`

## Principles

- **One command = one purpose** (OK even if workflow is long, as long as the purpose is single)
- **Side effects must be explicit** (network access / local file updates / external service writes, etc.)
- **"Execution" commands should be carefully separated** (separate generate → validate → review → apply to reduce accidents)

## Naming Convention (Collision Prevention)

### Recommended: `domain-subject-action` (shorten as needed)

- **domain**: Major category (e.g., `go`, `pr`, `zellij`, `repo`, `docs`)
- **subject**: Target (e.g., `change`, `changes`, `skill`, `command`)
- **action**: Action (e.g., `fetch`, `analyze`, `report`, `lint`, `validate`, `apply`)

Examples:

- `/go-changes-fetch`
- `/go-change-analyze`
- `/go-changes-report`

### Character Set

- Prioritize compatibility and portability, use **lowercase English letters + hyphens** as the base.
- While namespace expressions like `go:*` are possible, this repository currently uses **`go-` prefix** as a namespace.

## Responsibility Separation Policy (Decomposing Long Commands)

Commands that are "multi-purpose", "produce many outputs", or "have strong side effects" should be considered for separation into the following units:

- **fetch**: Fetch from external sources (network)
- **analyze**: Parse input (near-pure processing)
- **report**: Aggregate and output (final output)
- **generate / validate / review / apply**: Separate generate → validate → review → apply (especially for high-risk operations)

### Orchestrator Commands (Aggregator Commands)

Commands like `/go-catchup` that are "single-purpose workflows" are OK to keep, but must satisfy:

- **One-Line Description clearly focuses on a single purpose**
- **Output / Side Effects lists all side effects**
- If separated primitives exist (e.g., `*-fetch`, `*-analyze`), guide users via **Related Commands**

## Compatibility (Migration Note Guidelines)

When renaming or splitting existing commands, to avoid breaking changes, the following is recommended:

- **Keep old commands as compatibility aliases** (if possible)
- Add the following to the old command's `.md`:
  - **Status**: Deprecated / Compatibility
  - **Replacement**: New command name
  - **Migration**: Show replacement example in 1-3 lines
