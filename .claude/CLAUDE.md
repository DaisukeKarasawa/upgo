# Upgo - Claude Code Project Instructions

This file contains project-specific instructions for Claude Code.

## Project Overview

Upgo is a GitHub repository monitoring system for Go projects. It monitors PRs, collects diffs and comments, and provides Japanese summaries using a local LLM (Ollama).

## Technology Stack

### Backend (Go)
- **Framework**: Gin v1.11
- **Database**: SQLite with `mattn/go-sqlite3`
- **Config**: Viper
- **Logging**: Uber Zap
- **GitHub API**: `google/go-github/v60`

### Frontend (React + TypeScript)
- **Build Tool**: Vite
- **Styling**: Tailwind CSS
- **Language**: TypeScript

### External Services
- **Ollama**: Local LLM for Japanese summarization
- **GitHub API**: PR monitoring

## Directory Structure

```
upgo/
├── cmd/server/        # Application entry point
├── internal/          # Internal packages
├── web/              # React frontend
│   └── src/          # Frontend source
├── data/             # SQLite database
├── logs/             # Log files
└── backups/          # Database backups
```

## Development Workflow

### TDD Approach (Mandatory)

This project follows Kent Beck's TDD methodology:

1. **Red**: Write a failing test first
2. **Green**: Write minimum code to pass the test
3. **Refactor**: Improve code structure while keeping tests green

### Tidy First Principle

Separate changes into two types:
- **Structural Changes**: Refactoring without behavior changes
- **Behavioral Changes**: Adding/modifying functionality

Never mix these in the same commit.

## Useful Commands

```bash
# Development
make dev              # Start both backend and frontend dev servers
make dev-backend      # Start only backend
make dev-frontend     # Start only frontend

# Testing
make test             # Run unit tests
make bench            # Run benchmarks
make perf-test        # Run performance tests
make test-all         # Run all tests

# Building
make build            # Build both backend and frontend
make clean            # Clean build artifacts

# Operations
make backup           # Create database backup
```

## API Endpoints

- Backend: `http://localhost:8081`
- Frontend (dev): `http://localhost:5173`
- API prefix: `/api/v1/`

## Commit Conventions

Use [gitmoji](https://gist.github.com/parmentf/035de27d6ed1dce0b36a) in commit messages:

| Gitmoji | Usage |
|---------|-------|
| `:sparkles:` | New feature |
| `:bug:` | Bug fix |
| `:recycle:` | Refactor |
| `:white_check_mark:` | Add/update tests |
| `:memo:` | Documentation |
| `:art:` | Improve structure/format |
| `:zap:` | Performance improvement |
| `:fire:` | Remove code/files |
| `:construction:` | Work in progress |

Format: `<gitmoji> <commit message>`

Example: `:sparkles: Add user authentication feature`

## Testing Guidelines

### Go Tests
- Place tests in the same package as the code
- Use table-driven tests where appropriate
- Name test files with `_test.go` suffix

### Frontend Tests
- Run with `cd web && npm test`
- Use React Testing Library patterns

## Environment Variables

- `GITHUB_TOKEN`: GitHub personal access token
- `UPGO_ENV`: Environment (`development`/`production`)

## Important Notes

- Always run tests before committing
- Keep commit size small and focused
- Update AGENTS.md when adding new conventions
- Use Japanese for user-facing content, English for code/comments
