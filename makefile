SHELL := /bin/bash
MIGRATE := $(shell go env GOPATH)/bin/migrate
BINARY  := bin/server

ifneq (,$(wildcard .env))
    include .env
    export
endif

DB_URL=mysql://$(DB_USER):$(DB_PASS)@tcp($(DB_HOST):$(DB_PORT))/$(DB_NAME)

.PHONY: run build test clean \
        migrate-up migrate-down migrate-version migrate-create \
        docker-up docker-down docker-build docker-logs

# ── Development ───────────────────────────────────────────────────
run:
	go run ./cmd/server/...

build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BINARY) ./cmd/server/...

test:
	go test ./... -v

test-short:
	go test ./internal/... -v

test-v:
	go test ./... -v

test-race:
	go test -race ./...

test-race-v:
	go test -race ./... -v

bench:
	go test -bench=. -benchmem ./...

bench-util:
	go test -bench=. -benchmem ./internal/util/...

bench-middleware:
	go test -bench=. -benchmem ./internal/middleware/...

bench-race:
	go test -bench=. -benchmem -race ./...

bench-race:
	go test -bench=. -benchmem -race ./internal/middleware/...
	
clean:
	rm -rf bin/

# ── Database migrations ───────────────────────────────────────────
migrate-up:
	$(MIGRATE) -path migrations -database "$(DB_URL)" up

migrate-down:
	$(MIGRATE) -path migrations -database "$(DB_URL)" down 1

migrate-version:
	$(MIGRATE) -path migrations -database "$(DB_URL)" version

migrate-create:
	@read -p "Migration name: " name; \
	$(MIGRATE) create -ext sql -dir migrations -seq $$name

migrate-drop:
	$(MIGRATE) -path migrations -database "$(DB_URL)" drop -f

# ── Docker ────────────────────────────────────────────────────────
docker-build:
	docker compose build

docker-up:
	docker compose --env-file .env.docker up -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f app

docker-clean:
	docker compose down -v  # removes volumes too — wipes DB

# ── Helpers ───────────────────────────────────────────────────────
# Print all available commands
help:
	@echo ""
	@echo "Usage: make <command>"
	@echo ""
	@echo "Development:"
	@echo "  run              Start server locally"
	@echo "  build            Build binary to bin/server"
	@echo "  test             Run all tests"
	@echo "  test-short       Run unit tests only (no DB)"
	@echo "  clean            Remove build artifacts"
	@echo ""
	@echo "Migrations:"
	@echo "  migrate-up       Apply all pending migrations"
	@echo "  migrate-down     Roll back one migration"
	@echo "  migrate-version  Show current migration version"
	@echo "  migrate-create   Create a new migration"
	@echo "  migrate-drop     Drop everything (careful!)"
	@echo ""
	@echo "Docker:"
	@echo "  docker-build     Build Docker image"
	@echo "  docker-up        Start all services"
	@echo "  docker-down      Stop all services"
	@echo "  docker-logs      Tail app logs"
	@echo "  docker-clean     Stop + remove volumes"
	@echo ""