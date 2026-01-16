---
description: Create database backup
allowed-tools: Bash(make:*), Bash(curl:*), Bash(cp:*), Bash(ls:*), Read
---

# Database Backup

## Current Backups

!`ls -la backups/ 2>/dev/null | tail -5 || echo "No backups found"`

## Instructions

Create a backup of the SQLite database:

1. Via API (server must be running):
   ```bash
   curl -X POST http://localhost:8081/api/v1/backup
   ```

2. Via Make:
   ```bash
   make backup
   ```

3. Manual copy:
   ```bash
   cp data/upgo.db backups/upgo.db.$(date +%Y%m%d_%H%M%S)
   ```

## Backup Location

Backups are stored in the `backups/` directory.

## Restore

To restore from backup:
```bash
cp backups/<backup-file> data/upgo.db
```
