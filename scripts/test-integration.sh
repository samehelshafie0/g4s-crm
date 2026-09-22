#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
container="g4s-crm-test-$$"
cleanup() { docker rm -f "$container" >/dev/null 2>&1 || true; }
trap cleanup EXIT
docker run -d --name "$container" -e POSTGRES_USER=crm_test -e POSTGRES_PASSWORD=crm_test_local -e POSTGRES_DB=crm_test -p 127.0.0.1::5432 postgres:16-alpine >/dev/null
for attempt in {1..30}; do
  if docker exec "$container" pg_isready -U crm_test -d crm_test >/dev/null 2>&1; then break; fi
  sleep 1
done
port="$(docker port "$container" 5432/tcp | cut -d: -f2)"
export CRM_TEST_DATABASE_URL="postgres://crm_test:crm_test_local@127.0.0.1:${port}/crm_test?sslmode=disable"
cd crm-api
if [[ -z "${PDF_PYTHON:-}" && -x "$PWD/tmp/pdf-venv/bin/python" ]]; then export PDF_PYTHON="$PWD/tmp/pdf-venv/bin/python"; fi
go test -race ./... -count=1 -timeout 180s
