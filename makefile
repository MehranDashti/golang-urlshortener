SHELL    := /bin/bash
MIGRATE  := $(shell go env GOPATH)/bin/migrate
GOLANGCI := $(shell go env GOPATH)/bin/golangci-lint
BINARY   := bin/server

ifneq (,$(wildcard .env))
    include .env
    export
endif

DB_URL=mysql://$(DB_USER):$(DB_PASS)@tcp($(DB_HOST):$(DB_PORT))/$(DB_NAME)

.PHONY: run run-pprof build build-pprof \
        lint lint-fix fmt fmt-check \
        test test-v test-short \
        test-unit test-integration test-race test-race-v \
        test-race-integration test-all \
        bench bench-util bench-middleware bench-race \
        migrate-up migrate-down migrate-version migrate-create migrate-drop \
        docker-build docker-up docker-down docker-logs docker-clean \
        profile-cpu profile-heap profile-goroutines \
        clean help

# ── Linting ───────────────────────────────────────────────────────

lint:
	$(GOLANGCI) run ./...

lint-fix:
	$(GOLANGCI) run --fix ./...

fmt:
	$(GOLANGCI) fmt ./...

fmt-check:
	$(GOLANGCI) fmt --diff ./...

# ── Development ───────────────────────────────────────────────────

run:
	go run ./cmd/server/...

run-pprof:
	go run -tags pprof ./cmd/server/...

build:
	CGO_ENABLED=0 go build -ldflags="-s -w" \
		-o $(BINARY) ./cmd/server/...

build-pprof:
	CGO_ENABLED=0 go build -tags pprof -ldflags="-s -w" \
		-o bin/server-pprof ./cmd/server/...

clean:
	rm -rf bin/

# ── Testing ───────────────────────────────────────────────────────

test:
	go test ./...

test-v:
	go test ./... -v

test-unit:
	go test ./internal/... -v

test-short:
	go test -short ./internal/... -v

test-integration:
	go test -tags integration ./... -v

test-race:
	go test -race ./...

test-race-v:
	go test -race ./... -v

test-race-integration:
	go test -race -tags integration ./...

test-all:
	go test -race -tags integration ./... -v

fuzz:
	go test -fuzz=FuzzNormaliseURL -fuzztime=30s ./internal/util/...

fuzz-url:
	go test -fuzz=FuzzNormaliseURL -fuzztime=30s ./internal/util/...

fuzz-shortcode:
	go test -fuzz=FuzzGenerateShortCode -fuzztime=30s ./internal/util/...
	
# ── Benchmarks ────────────────────────────────────────────────────

bench:
	go test -bench=. -benchmem ./...

bench-util:
	go test -bench=. -benchmem ./internal/util/...

bench-middleware:
	go test -bench=. -benchmem ./internal/middleware/...

bench-race:
	go test -bench=. -benchmem -race ./internal/middleware/...

# ── Migrations ────────────────────────────────────────────────────

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
	docker compose down -v

# ── Profiling ─────────────────────────────────────────────────────

profile-cpu:
	curl -s "http://localhost:8080/api/v1/debug/pprof/profile?seconds=30" \
		-o cpu.prof && go tool pprof cpu.prof

profile-heap:
	curl -s "http://localhost:8080/api/v1/debug/pprof/heap" \
		-o heap.prof && go tool pprof heap.prof

profile-goroutines:
	curl -s "http://localhost:8080/api/v1/debug/pprof/goroutine?debug=2"

# ── Help ──────────────────────────────────────────────────────────

help:
	@echo ""
	@echo "Usage: make <command>"
	@echo ""
	@echo "Linting:"
	@echo "  lint                 Run golangci-lint (staticcheck, errcheck, revive, ...)"
	@echo "  lint-fix             Run golangci-lint and auto-fix what it can"
	@echo "  fmt                  Format code (gofmt + goimports)"
	@echo "  fmt-check            Show formatting diff without writing"
	@echo ""
	@echo "Development:"
	@echo "  run                  Start server locally"
	@echo "  run-pprof            Start server with pprof endpoints"
	@echo "  build                Build binary to bin/server"
	@echo "  build-pprof          Build binary with pprof support"
	@echo "  clean                Remove build artifacts"
	@echo ""
	@echo "Testing:"
	@echo "  test                 Run all tests"
	@echo "  test-v               Run all tests verbose"
	@echo "  test-unit            Run unit tests only"
	@echo "  test-short           Run tests skipping slow ones"
	@echo "  test-integration     Run integration tests (needs DB)"
	@echo "  test-race            Run tests with race detector"
	@echo "  test-race-integration Run integration + race detector"
	@echo "  test-all             Run everything + race detector"
	@echo ""
	@echo "Benchmarks:"
	@echo "  bench                Run all benchmarks"
	@echo "  bench-util           Benchmark util package"
	@echo "  bench-middleware     Benchmark middleware package"
	@echo "  bench-race           Benchmark middleware with race detector"
	@echo ""
	@echo "Migrations:"
	@echo "  migrate-up           Apply all pending migrations"
	@echo "  migrate-down         Roll back one migration"
	@echo "  migrate-version      Show current migration version"
	@echo "  migrate-create       Create new migration files"
	@echo "  migrate-drop         Drop everything (careful!)"
	@echo ""
	@echo "Docker:"
	@echo "  docker-build         Build Docker image"
	@echo "  docker-up            Start all services"
	@echo "  docker-down          Stop all services"
	@echo "  docker-logs          Tail app logs"
	@echo "  docker-clean         Stop + remove volumes"
	@echo ""
	@echo "Profiling (server must be running):"
	@echo "  profile-cpu          Capture 30s CPU profile"
	@echo "  profile-heap         Capture heap memory profile"
	@echo "  profile-goroutines   Show all goroutine stack traces"
	@echo ""