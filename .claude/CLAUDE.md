# Upgo - Claude Code Project Instructions

This file contains project-specific instructions for Claude Code.

## Project Overview

Upgo is a Claude Code plugin that automatically fetches and analyzes golang/go Changes (CLs) via Gerrit to help learn Go design philosophy. It provides skills and commands for Claude Code to interact with Gerrit Changes and extract insights about Go's design principles.

## Technology Stack

- **Gerrit REST API**: HTTP API for accessing Gerrit code review system
- **curl**: Command-line tool for HTTP requests
- **Markdown**: Skills and commands are defined in Markdown files
- **Claude Code Plugin**: Plugin manifest format

## Directory Structure

```
upgo/
├── .claude-plugin/       # Plugin manifest
│   └── plugin.json
├── skills/               # User-facing Skills
│   ├── go-pr-fetcher/    # Change fetching skill
│   └── go-pr-analyzer/   # Change analysis skill
├── commands/             # User-facing Commands
│   └── go-catchup.md     # Catchup command
└── .claude/              # Developer tools (internal)
    ├── skills/           # Development skills (zellij, Go development)
    ├── commands/         # Development commands
    └── agents/           # Development agents
```

## Development Workflow

### Plugin Structure

- Skills are defined in `skills/*/SKILL.md`
- Commands are defined in `commands/*.md`
- Plugin metadata is in `.claude-plugin/plugin.json`

### Testing

Use the `plugin-test` skill to validate plugin structure and functionality.

## Commit Conventions

Use [gitmoji](https://gist.github.com/parmentf/035de27d6ed1dce0b36a) in commit messages:

| Gitmoji              | Usage                    |
| -------------------- | ------------------------ |
| `:sparkles:`         | New feature              |
| `:bug:`              | Bug fix                  |
| `:recycle:`          | Refactor                 |
| `:white_check_mark:` | Add/update tests         |
| `:memo:`             | Documentation            |
| `:art:`              | Improve structure/format |
| `:zap:`              | Performance improvement  |
| `:fire:`             | Remove code/files        |
| `:construction:`     | Work in progress         |

Format: `<gitmoji> <commit message>`

Example: `:sparkles: Add new Change analysis feature`

## Important Notes

- Keep commit size small and focused
- Use Japanese for user-facing content, English for code/comments
- Follow the plugin structure defined in README.md
