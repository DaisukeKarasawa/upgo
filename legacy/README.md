# Legacy Code

This directory contains the legacy Web UI implementation of Upgo.

## Contents

- `web/` - React frontend (Vite + TypeScript + Tailwind CSS)
- `cmd/server/` - HTTP server entry point
- `internal/api/` - REST API handlers
- `internal/service/` - Business logic services
- `internal/scheduler/` - Background job scheduler
- `internal/tracker/` - PR tracking service
- `internal/logger/` - Logging utilities

## Status

This code has been moved here as part of the migration to Claude Code Skills format (Issue #21).

The new system uses:
- `cmd/skillgen/` - CLI tool for skill generation
- `internal/skillgen/` - Skill generation service
- `.claude/skills/` - Generated skill files

## Running Legacy Server

If you need to run the legacy Web UI:

```bash
cd legacy
go run cmd/server/main.go
```

Note: The legacy server requires the frontend build:
```bash
cd legacy/web
npm install
npm run build
```

## Migration Date

Migrated: 2026-01-17
