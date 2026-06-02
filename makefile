SHELL := /bin/bash
MIGRATE := $(shell go env GOPATH)/bin/migrate

ifneq (,$(wildcard .env))
    include .env
    export
endif

DB_URL=mysql://$(DB_USER):$(DB_PASS)@tcp($(DB_HOST):$(DB_PORT))/$(DB_NAME)

.PHONY: run build test migrate-up migrate-down migrate-version migrate-create

run:
	go run cmd/server/main.go

build:
	go build -o bin/server cmd/server/main.go

test:
	go test ./...

migrate-up:
	$(MIGRATE) -path migrations -database "$(DB_URL)" up

migrate-down:
	$(MIGRATE) -path migrations -database "$(DB_URL)" down 1

migrate-version:
	$(MIGRATE) -path migrations -database "$(DB_URL)" version

migrate-create:
	@read -p "Migration name: " name; \
	$(MIGRATE) create -ext sql -dir migrations -seq $$name