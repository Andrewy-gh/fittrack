.PHONY: help build clean docker-down migrate-create migrate-up migrate-down sqlc swagger run test test-fast test-short vet dev

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
ENV_FILE ?= $(if $(wildcard .setenv.sh),.setenv.sh,$(if $(wildcard setenv.sh),setenv.sh,setenv.sh))

help:
	@echo "Available commands:"
	@echo "  build        - Compile the project"
	@echo "  clean        - Clean build files"
	@echo "  docker-down  - Stop Docker containers"
	@echo "  run          - Run the application"
	@echo "  test-fast    - Quick local check for server code changes (no DB package)"
	@echo "  test-short   - Use before push for DB/integration-impacting server changes"
	@echo "  test         - Full verbose suite (deep local check / CI parity)"
	@echo "  vet          - Run go vet"
	@echo "  migrate-up   - Run database migrations up"
	@echo "  migrate-down - Rollback database migrations"
	@echo "  sqlc         - Generate SQL code"
	@echo "  swagger      - Generate OpenAPI documentation"
	@echo "  dev          - Complete dev setup: load env, start DB, apply migrations, run air"

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
	@if [ -f "$(ENV_FILE)" ] && . ./$(ENV_FILE) && [ -n "$$DATABASE_URL" ]; then \
		echo "DATABASE_URL: $$DATABASE_URL"; \
		goose -dir migrations postgres "$$DATABASE_URL" up; \
	else \
		echo "DATABASE_URL does not exist"; \
		exit 1; \
	fi

migrate-down:
	@if [ -f "$(ENV_FILE)" ] && . ./$(ENV_FILE) && [ -n "$$DATABASE_URL" ]; then \
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

test-fast: ## Run fast test suite without DB integration package
	@pkgs="$$( $(GOCMD) list ./... | grep -Ev '/internal/database$$' || true )"; \
	if [ -z "$$pkgs" ]; then \
		echo "No packages selected for fast tests."; \
		exit 0; \
	fi; \
	$(GOTEST) -short $$pkgs

test-short: ## Start local DB and run short test suite
	@echo "==> Starting database container: $(DB_SERVICE)"
	@if ! command -v docker >/dev/null 2>&1; then \
		echo "ERROR: docker is not installed or not on PATH."; \
		exit 1; \
	fi
	@if [ ! -f "$(ENV_FILE)" ]; then \
		echo "ERROR: expected env file at $$(pwd)/$(ENV_FILE)"; \
		echo "Create .setenv.sh or setenv.sh with DB_USER, DB_PASSWORD, DB_NAME, and DATABASE_URL before running make test-short."; \
		exit 1; \
	fi
	. ./$(ENV_FILE); \
	for req_var in DB_USER DB_PASSWORD DB_NAME DATABASE_URL; do \
		if [ -z "$${!req_var}" ]; then \
			echo "ERROR: $$req_var is not set in $(ENV_FILE)"; \
			exit 1; \
		fi; \
	done; \
	db_port="$${DB_PORT:-5432}"; \
	db_url="postgresql://$${DB_USER}:$${DB_PASSWORD}@127.0.0.1:$${db_port}/$${DB_NAME}"; \
	case "$$db_url" in \
		*sslmode=*) ;; \
		*\?*) db_url="$$db_url&sslmode=disable" ;; \
		*) db_url="$$db_url?sslmode=disable" ;; \
	esac; \
	if ! command -v goose >/dev/null 2>&1; then \
		echo "ERROR: goose is not installed or not on PATH."; \
		exit 1; \
	fi; \
	DB_PORT="$$db_port" docker compose up -d $(DB_SERVICE); \
	start_ts=$$(date +%s); \
	while true; do \
		if docker compose exec -T $(DB_SERVICE) sh -c 'pg_isready -U "$${POSTGRES_USER:-postgres}" -d "$${POSTGRES_DB:-postgres}" >/dev/null 2>&1'; then \
			break; \
		fi; \
		now=$$(date +%s); \
		elapsed=$$(( now - start_ts )); \
		if [ "$$elapsed" -ge 30 ]; then \
			echo "ERROR: Timed out waiting for PostgreSQL readiness"; \
			exit 1; \
		fi; \
		sleep 1; \
	done; \
	echo "==> Applying migrations"; \
	goose -dir migrations postgres "$$db_url" up; \
	echo "==> Running go test -short ./..."; \
	DATABASE_URL="$$db_url" $(GOTEST) -short ./...

vet:
	$(GOCMD) vet -v ./...

dev: ## Complete development setup: load env, start database, wait for ready, apply migrations, run air
	@echo "==> Loading environment from $(ENV_FILE)"
	@if [ ! -f "$(ENV_FILE)" ]; then \
		echo "ERROR: expected env file at $$(pwd)/$(ENV_FILE)"; \
		echo "Create it and export the required variables (e.g., POSTGRES_USER, POSTGRES_DB, POSTGRES_PASSWORD)."; \
		exit 1; \
	fi
	. ./$(ENV_FILE)

	@echo "==> Checking prerequisites: docker, docker compose, goose, and $(AIR_CMD)"
	@if ! command -v docker >/dev/null 2>&1; then \
		echo "ERROR: docker is not installed or not on PATH."; \
		exit 1; \
	fi
	@if ! docker compose version >/dev/null 2>&1; then \
		echo "ERROR: docker compose plugin not available. Install Docker Desktop or the compose plugin."; \
		exit 1; \
	fi
	@if ! command -v goose >/dev/null 2>&1; then \
		echo "ERROR: 'goose' not found."; \
		echo "Install: 'go install github.com/pressly/goose/v3/cmd/goose@latest'"; \
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

	@echo "==> Applying database migrations"
	@if [ -z "$$DATABASE_URL" ]; then \
		echo "ERROR: DATABASE_URL is not set after loading setenv.sh"; \
		exit 1; \
	fi
	goose -dir migrations postgres "$$DATABASE_URL" up

	@echo "==> Starting $(AIR_CMD) for hot-reload development"
	@echo "==> Press Ctrl+C to stop the development server"
	exec $(AIR_CMD)
