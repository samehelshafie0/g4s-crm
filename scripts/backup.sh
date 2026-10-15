#!/usr/bin/env bash
# Back up the running Docker stack: the PostgreSQL database and the uploaded
# documents. Run from cron on the server, e.g. daily at 02:00:
#   0 2 * * * cd /opt/g4s-crm && ./scripts/backup.sh >> backups/backup.log 2>&1
# BACKUP_DIR (default backups/) and BACKUP_KEEP_DAYS (default 14) can be set.
set -euo pipefail
cd "$(dirname "$0")/.."

dir="${BACKUP_DIR:-backups}"
keep="${BACKUP_KEEP_DAYS:-14}"
stamp="$(date +%Y%m%d-%H%M%S)"
umask 077
mkdir -p "$dir"

db="$dir/g4s_crm-$stamp.dump"
files="$dir/storage-$stamp.tgz"

# Write to a temporary name first so a failed run never looks like a backup.
docker compose exec -T postgres pg_dump -U g4s_app -d g4s_crm -Fc > "$db.partial"
docker compose exec -T postgres pg_restore --list < "$db.partial" > /dev/null
mv "$db.partial" "$db"

docker compose exec -T api tar czf - -C /app/storage . > "$files.partial"
tar tzf "$files.partial" > /dev/null
mv "$files.partial" "$files"

find "$dir" -maxdepth 1 \( -name 'g4s_crm-*.dump' -o -name 'storage-*.tgz' \) -mtime +"$keep" -delete

echo "$(date '+%F %T') backup ok: $db ($(du -h "$db" | cut -f1)), $files ($(du -h "$files" | cut -f1))"
