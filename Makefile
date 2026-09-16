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
	@test -f crm-api/keys/private.pem || (umask 077; openssl genrsa -out crm-api/keys/private.pem 2048)
	@test -f crm-api/keys/public.pem || openssl rsa -in crm-api/keys/private.pem -pubout -out crm-api/keys/public.pem
	@echo "Keys created in crm-api/keys/"
	@echo ""
	@echo "Copying .env.example → .env"
	@cp -n .env.example .env || true
	@echo ""
	@echo "Done! Edit .env then run: make up"

# ─── Database ────────────────────────────────────────────────
migrate:
	docker compose run --rm --no-deps api ./g4s-crm-api migrate

seed:
	docker compose run --rm --no-deps -e ADMIN_EMAIL -e ADMIN_PASSWORD -e ADMIN_FIRST_NAME -e ADMIN_LAST_NAME api ./g4s-crm-api bootstrap-admin

# Local development uses its own compose project and ports.
dev-db:
	docker compose -f compose.dev.yml up -d --wait

dev-migrate:
	./scripts/dev-api.sh migrate

dev-admin:
	./scripts/dev-api.sh bootstrap-admin

dev-api:
	./scripts/dev-api.sh

dev-web:
	cd crm-dashboard && npm run dev

test-integration:
	./scripts/test-integration.sh

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
	@echo ""
	@echo "  CLEANUP:"
	@echo "    make clean          Stop + delete all volumes (wipes DB!)"
	@echo ""
