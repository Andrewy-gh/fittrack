.PHONY: help build clean docker-up docker-down migrate-create migrate-up migrate-down sqlc swagger run test vet

GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=api

help:
	@echo "Available commands:"
	@echo "  build        - Compile the project"
	@echo "  clean        - Clean build files"
	@echo "  docker-up    - Start Docker containers"
	@echo "  docker-down  - Stop Docker containers"
	@echo "  run          - Run the application"
	@echo "  test         - Run tests"
	@echo "  migrate-up   - Run database migrations up"
	@echo "  migrate-down - Rollback database migrations"
	@echo "  sqlc         - Generate SQL code"
	@echo "  swagger      - Generate OpenAPI documentation"

build:
	$(GOBUILD) -o bin/$(BINARY_NAME) ./cmd/$(BINARY_NAME)

clean:
	$(GOCLEAN) -cache -testcache
	rm -rf bin
	rm -rf dist

docker-up:
	@if . ./setenv.sh && [ -n "$$DATABASE_URL" ]; then \
		docker compose up -d; \
	else \
		echo "DATABASE_URL does not exist"; \
		exit 1; \
	fi

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