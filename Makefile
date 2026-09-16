.PHONY: up down build logs migrate seed keys clean help

# ─── Full-stack commands ─────────────────────────────────────

up:
	docker compose up -d
	@echo ""
	@echo "  G4S CRM is running at http://localhost"
	@echo ""

down:
	docker compose down

build:
	docker compose build --no-cache

rebuild:
	docker compose up -d --build

logs:
	docker compose logs -f

logs-api:
	docker compose logs -f api

logs-frontend:
	docker compose logs -f frontend

# ─── One-time setup ──────────────────────────────────────────
setup:
	@echo "Generating RSA key pair for JWT..."
	@mkdir -p crm-api/keys
	@openssl genrsa -out crm-api/keys/private.pem 2048
	@openssl rsa -in crm-api/keys/private.pem -pubout -out crm-api/keys/public.pem
	@echo "Keys created in crm-api/keys/"
	@echo ""
	@echo "Copying .env.example → .env"
	@cp -n .env.example .env || true
	@echo ""
	@echo "Done! Edit .env then run: make up"

# ─── Database ────────────────────────────────────────────────
migrate:
	docker compose exec api sh -c \
	  'migrate -path ./migrations -database "postgres://g4s_app:$$DB_PASSWORD@postgres:5432/g4s_crm?sslmode=disable" up'

migrate-down:
	docker compose exec api sh -c \
	  'migrate -path ./migrations -database "postgres://g4s_app:$$DB_PASSWORD@postgres:5432/g4s_crm?sslmode=disable" down 1'

psql:
	docker compose exec postgres psql -U g4s_app -d g4s_crm

# ─── Status ──────────────────────────────────────────────────
status:
	docker compose ps

clean:
	docker compose down -v --remove-orphans

help:
	@echo ""
	@echo "  G4S CRM — Full-Stack Docker Commands"
	@echo ""
	@echo "  FIRST TIME:"
	@echo "    make setup          Generate JWT keys + copy .env.example"
	@echo "    make up             Start all services"
	@echo "    make migrate        Run DB migrations"
	@echo ""
	@echo "  DAILY USE:"
	@echo "    make up             Start stack"
	@echo "    make down           Stop stack"
	@echo "    make rebuild        Rebuild images and restart"
	@echo "    make logs           Follow all logs"
	@echo "    make logs-api       Follow API logs only"
	@echo "    make status         Show running containers"
	@echo ""
	@echo "  DATABASE:"
	@echo "    make psql           Open psql shell"
	@echo "    make migrate        Apply all migrations"
	@echo "    make migrate-down   Roll back 1 migration"
	@echo ""
	@echo "  CLEANUP:"
	@echo "    make clean          Stop + delete all volumes (wipes DB!)"
	@echo ""
