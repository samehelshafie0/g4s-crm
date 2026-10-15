#!/usr/bin/env bash
# Restore the Docker stack from scripts/backup.sh output. This REPLACES the
# current database (and, if given, the uploaded documents).
#   ./scripts/restore.sh backups/g4s_crm-YYYYMMDD-HHMMSS.dump [backups/storage-YYYYMMDD-HHMMSS.tgz]
set -euo pipefail
cd "$(dirname "$0")/.."

db="${1:-}"
files="${2:-}"
if [[ -z "$db" || ! -f "$db" ]] || [[ -n "$files" && ! -f "$files" ]]; then
  echo "usage: $0 <database .dump> [storage .tgz]" >&2
  exit 1
fi

echo "This replaces the g4s_crm database with $db${files:+ and the uploaded documents with $files}."
read -r -p "Type RESTORE to continue: " answer
[[ "$answer" == "RESTORE" ]] || { echo "Cancelled."; exit 1; }

docker compose stop frontend api
docker compose exec -T postgres psql -U g4s_app -d postgres -v ON_ERROR_STOP=1 \
  -c "DROP DATABASE IF EXISTS g4s_crm WITH (FORCE)" -c "CREATE DATABASE g4s_crm OWNER g4s_app"
docker compose exec -T postgres pg_restore -U g4s_app -d g4s_crm --no-owner --exit-on-error < "$db"

if [[ -n "$files" ]]; then
  docker compose run --rm --no-deps -T --user 0 --entrypoint sh api \
    -c 'find /app/storage -mindepth 1 -delete && tar xzf - -C /app/storage && chown -R appuser /app/storage' < "$files"
fi

docker compose start api frontend
echo "Restore complete. Run the API migrations if this backup predates the current release:"
echo "  docker compose exec api ./g4s-crm-api migrate"
