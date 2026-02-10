.PHONY: help build clean docker-down migrate-create migrate-up migrate-down sqlc swagger run test vet dev

GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=api

# Use bash for all shell operations and run entire recipes in one shell.
# Note: GNU Make's SHELL must be a single executable path (no args).
SHELL := bash
.ONESHELL:
.SHELLFLAGS := -eo pipefail -c

# Configurable variables for development
DB_SERVICE ?= postgres
DB_READY_TIMEOUT ?= 90
AIR_CMD ?= air

help:
	@echo "Available commands:"
	@echo "  build        - Compile the project"
	@echo "  clean        - Clean build files"
	@echo "  docker-down  - Stop Docker containers"
	@echo "  run          - Run the application"
	@echo "  test         - Run tests"
	@echo "  vet          - Run go vet"
	@echo "  migrate-up   - Run database migrations up"
	@echo "  migrate-down - Rollback database migrations"
	@echo "  sqlc         - Generate SQL code"
	@echo "  swagger      - Generate OpenAPI documentation"
	@echo "  dev          - Complete dev setup: load env, start DB, wait for ready, run air"

build:
	$(GOBUILD) -o bin/$(BINARY_NAME) ./cmd/$(BINARY_NAME)

clean:
	$(GOCLEAN) -cache -testcache
	rm -rf bin
	rm -rf dist

docker-down:
	docker compose down

migrate-create:
ifeq ($(strip $(NAME)),)
	@echo "No migration name provided"
	@exit 1
else
	goose -dir migrations create -s $(NAME) sql
endif

migrate-up:
	@if . ./setenv.sh && [ -n "$$DATABASE_URL" ]; then \
		echo "DATABASE_URL: $$DATABASE_URL"; \
		goose -dir migrations postgres "$$DATABASE_URL" up; \
	else \
		echo "DATABASE_URL does not exist"; \
		exit 1; \
	fi

migrate-down:
	@if . ./setenv.sh && [ -n "$$DATABASE_URL" ]; then \
		goose -dir migrations postgres "$$DATABASE_URL" down; \
	else \
		echo "DATABASE_URL does not exist"; \
		exit 1; \
	fi

run:
	./bin/$(BINARY_NAME)

sqlc:
	sqlc generate

swagger:
	swag init -g cmd/api/main.go -o docs

test:
	$(GOTEST) -v ./...

vet:
	$(GOCMD) vet -v ./...

dev: ## Complete development setup: load env, start database, wait for ready, run air
	@echo "==> Loading environment from setenv.sh"
	@if [ ! -f setenv.sh ]; then \
		echo "ERROR: setenv.sh not found at $$(pwd)/setenv.sh"; \
		echo "Create it and export the required variables (e.g., POSTGRES_USER, POSTGRES_DB, POSTGRES_PASSWORD)."; \
		exit 1; \
	fi
	. ./setenv.sh

	@echo "==> Checking prerequisites: docker, docker compose, and $(AIR_CMD)"
	@if ! command -v docker >/dev/null 2>&1; then \
		echo "ERROR: docker is not installed or not on PATH."; \
		exit 1; \
	fi
	@if ! docker compose version >/dev/null 2>&1; then \
		echo "ERROR: docker compose plugin not available. Install Docker Desktop or the compose plugin."; \
		exit 1; \
	fi
	@if ! command -v $(AIR_CMD) >/dev/null 2>&1; then \
		echo "ERROR: '$(AIR_CMD)' not found."; \
		echo "Install: 'brew install air' or 'go install github.com/air-verse/air@latest'"; \
		exit 1; \
	fi

	@echo "==> Starting database container: $(DB_SERVICE)"
	docker compose up -d $(DB_SERVICE)

	@echo "==> Waiting for PostgreSQL to accept connections (timeout: $(DB_READY_TIMEOUT)s)"
	if ! docker compose exec -T $(DB_SERVICE) sh -c "command -v pg_isready >/dev/null 2>&1"; then \
		echo "WARNING: 'pg_isready' not found in $(DB_SERVICE) container."; \
		echo "Ensure the Postgres image provides pg_isready (official postgres:16.2 does)."; \
	fi
	start_ts=$$(date +%s); \
	while true; do \
		if docker compose exec -T $(DB_SERVICE) sh -c 'pg_isready -U "$${POSTGRES_USER:-postgres}" -d "$${POSTGRES_DB:-postgres}" >/dev/null 2>&1'; then \
			echo "==> PostgreSQL is ready!"; \
			break; \
		fi; \
		now=$$(date +%s); \
		elapsed=$$(( now - start_ts )); \
		if [ "$$elapsed" -ge "$(DB_READY_TIMEOUT)" ]; then \
			echo "ERROR: Timed out after $(DB_READY_TIMEOUT)s waiting for PostgreSQL."; \
			echo "Tip: Check logs with: docker compose logs $(DB_SERVICE)"; \
			exit 1; \
		fi; \
		echo "Waiting for database... ($${elapsed}s elapsed)"; \
		sleep 2; \
	done

	@echo "==> Starting $(AIR_CMD) for hot-reload development"
	@echo "==> Press Ctrl+C to stop the development server"
	exec $(AIR_CMD)
