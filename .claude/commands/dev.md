---
description: Start development server
allowed-tools: Bash(make:*), Bash(go run:*), Bash(npm:*), Bash(lsof:*), Bash(kill:*)
argument-hint: [all|backend|frontend|stop]
---

# Development Server

## Current Status

Port 8081 (Backend): !`lsof -i :8081 2>/dev/null | head -2 || echo "Not running"`
Port 5173 (Frontend): !`lsof -i :5173 2>/dev/null | head -2 || echo "Not running"`

## Instructions

Start or stop development servers:

- **No argument or "all"**: Start both servers (`make dev`)
- **"backend"**: Start backend only (`make dev-backend`)
- **"frontend"**: Start frontend only (`make dev-frontend`)
- **"stop"**: Stop all development servers

## Argument

$ARGUMENTS

## Access URLs

- **Frontend (dev)**: http://localhost:5173
- **Backend API**: http://localhost:8081
- **API docs**: http://localhost:8081/api/v1/

## Notes

- Frontend uses Vite HMR for hot reloading
- Backend requires manual restart on Go file changes
- API requests from frontend are proxied to backend
