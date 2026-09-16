.PHONY: run build test lint tidy migrate migrate-down keys seed clean docker-up docker-down help

# ─── Variables ──────────────────────────────────────────────
BINARY_NAME=g4s-crm-api
MAIN=./cmd/server
MIGRATE=migrate -path ./migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)"

# Load .env if it exists
-include .env
export

# ─── Dev ────────────────────────────────────────────────────
run:
	go run $(MAIN)

build:
	CGO_ENABLED=0 go build -ldflags="-w -s" -o $(BINARY_NAME) $(MAIN)
	@echo "Binary built: ./$(BINARY_NAME)"

test:
	go test ./... -v -timeout 30s

lint:
	golangci-lint run ./...

tidy:
	go mod tidy

# ─── Keys ───────────────────────────────────────────────────
keys:
	@mkdir -p keys
	@openssl genrsa -out keys/private.pem 2048
	@openssl rsa -in keys/private.pem -pubout -out keys/public.pem
	@echo "RSA key pair generated in ./keys/"

# ─── Migrations ─────────────────────────────────────────────
migrate:
	$(MIGRATE) up

migrate-down:
	$(MIGRATE) down 1

migrate-force:
	$(MIGRATE) force $(VERSION)

migrate-status:
	$(MIGRATE) version

# ─── Database Seed ──────────────────────────────────────────
seed:
	go run ./seeds/main.go

# ─── Docker ─────────────────────────────────────────────────
docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-build:
	docker compose build --no-cache

docker-logs:
	docker compose logs -f api

# ─── Clean ──────────────────────────────────────────────────
clean:
	rm -f $(BINARY_NAME)
	go clean -cache

# ─── Help ───────────────────────────────────────────────────
help:
	@echo ""
	@echo "G4S CRM API — Available Targets:"
	@echo ""
	@echo "  make run             Run the API server locally"
	@echo "  make build           Compile production binary"
	@echo "  make test            Run all tests"
	@echo "  make lint            Run linter"
	@echo "  make tidy            Tidy go modules"
	@echo "  make keys            Generate RSA key pair for JWT"
	@echo "  make migrate         Run all up migrations"
	@echo "  make migrate-down    Roll back 1 migration"
	@echo "  make seed            Seed the database"
	@echo "  make docker-up       Start Docker stack"
	@echo "  make docker-down     Stop Docker stack"
	@echo "  make clean           Remove binary and cache"
	@echo ""
