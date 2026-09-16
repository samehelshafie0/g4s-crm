#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../crm-api"
# Explicit local-only defaults override any historical .env port settings.
export APP_ENV=development APP_HOST=127.0.0.1 APP_PORT="${CRM_API_PORT:-18080}"
export DB_HOST=127.0.0.1 DB_PORT="${CRM_DB_PORT:-55432}" DB_NAME=g4s_crm_dev DB_USER=g4s_dev DB_PASSWORD=crm-local-development DB_SSLMODE=disable
export JWT_PRIVATE_KEY_PATH=./keys/dev-private.pem JWT_PUBLIC_KEY_PATH=./keys/dev-public.pem
mkdir -p keys
if [[ ! -f "$JWT_PRIVATE_KEY_PATH" ]]; then
  (umask 077; openssl genrsa -out "$JWT_PRIVATE_KEY_PATH" 2048 2>/dev/null)
fi
if [[ ! -f "$JWT_PUBLIC_KEY_PATH" ]]; then
  openssl rsa -in "$JWT_PRIVATE_KEY_PATH" -pubout -out "$JWT_PUBLIC_KEY_PATH" 2>/dev/null
fi
go run ./cmd/server "$@"
