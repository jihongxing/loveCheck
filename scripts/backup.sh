#!/bin/bash
# LoveCheck daily backup script
# Add to crontab: 0 3 * * * /opt/lovecheck/scripts/backup.sh >> /var/log/lovecheck-backup.log 2>&1

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
BACKUP_DIR="${BACKUP_DIR:-/opt/backups/lovecheck}"
DATE=$(date +%Y%m%d_%H%M%S)
KEEP_DAYS=30

mkdir -p "$BACKUP_DIR"

echo "[$DATE] Starting backup..."

# PostgreSQL dump
docker compose -f "$PROJECT_DIR/docker-compose.prod.yml" exec -T postgres \
  pg_dump -U lovecheck --format=custom lovecheck \
  > "$BACKUP_DIR/pg_${DATE}.dump"

# Compress
gzip "$BACKUP_DIR/pg_${DATE}.dump"

# Cleanup old backups
find "$BACKUP_DIR" -name "*.gz" -mtime +${KEEP_DAYS} -delete

SIZE=$(du -sh "$BACKUP_DIR/pg_${DATE}.dump.gz" | cut -f1)
echo "[$DATE] Backup completed: pg_${DATE}.dump.gz ($SIZE)"
