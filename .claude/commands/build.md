---
description: Build the project for production
allowed-tools: Bash(make:*), Bash(go build:*), Bash(npm:*), Read
argument-hint: [all|backend|frontend|clean]
---

# Build Project

## Current Build Status

Backend binary: !`ls -la bin/ 2>/dev/null || echo "Not built"`
Frontend dist: !`ls -la web/dist/ 2>/dev/null || echo "Not built"`

## Instructions

Build project based on argument:

- **No argument or "all"**: Build both backend and frontend (`make build`)
- **"backend"**: Build Go binary only
- **"frontend"**: Build frontend only (`cd web && npm run build`)
- **"clean"**: Clean build artifacts (`make clean`)

## Argument

$ARGUMENTS

## Build Output

- Backend: `bin/server`
- Frontend: `web/dist/`

## Production Run

After building, run with:
```bash
make run
```

Access at http://localhost:8081
