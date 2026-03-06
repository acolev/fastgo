SHELL := /bin/sh

APP_NAME := fastgo
GO_CACHE := $(CURDIR)/.gocache
GO_MOD_CACHE := $(CURDIR)/.gomodcache
COMPOSE := docker compose
SWAG := $(shell sh -c 'GOBIN=$$(go env GOBIN); if [ -n "$$GOBIN" ]; then printf "%s/swag" "$$GOBIN"; else printf "%s/bin/swag" "$$(go env GOPATH)"; fi')
AIR := $(shell sh -c 'GOBIN=$$(go env GOBIN); if [ -n "$$GOBIN" ]; then printf "%s/air" "$$GOBIN"; else printf "%s/bin/air" "$$(go env GOPATH)"; fi')
GOLANGCI_LINT := $(shell sh -c 'GOBIN=$$(go env GOBIN); if [ -n "$$GOBIN" ]; then printf "%s/golangci-lint" "$$GOBIN"; else printf "%s/bin/golangci-lint" "$$(go env GOPATH)"; fi')
GOLANGCI_LINT_CACHE := $(CURDIR)/.golangci-lint-cache
SWAG_VERSION ?= v1.16.6
GOLANGCI_LINT_VERSION ?= latest
BUILD_DIR := ./tmp
BIN_PATH := $(BUILD_DIR)/$(APP_NAME)

.PHONY: help dev run up infra-up down logs ps test test-race lint lint-fix docs docs-check fmt build docker-build tidy check clean

help:
	@printf "%s\n" \
		"make dev          - run the app with air hot reload" \
		"make run          - run the app directly with go run" \
		"make up           - start full docker compose stack" \
		"make infra-up     - start postgres primary/replica and redis only" \
		"make down         - stop docker compose stack and remove volumes" \
		"make logs         - tail docker compose logs" \
		"make ps           - show docker compose services" \
		"make test         - run go test ./..." \
		"make test-race    - run tests with the race detector" \
		"make lint         - run golangci-lint" \
		"make lint-fix     - run golangci-lint with --fix" \
		"make docs         - generate Swagger docs" \
		"make docs-check   - verify generated docs are up to date" \
		"make fmt          - gofmt all Go files" \
		"make build        - build the API binary into ./tmp" \
		"make docker-build - build the Docker image" \
		"make tidy         - run go mod tidy" \
		"make check        - run fmt, test, lint, docs-check" \
		"make clean        - remove local build artifacts"

dev:
	@test -x "$(AIR)" || go install github.com/air-verse/air@latest
	@$(AIR)

run:
	@GOCACHE="$(GO_CACHE)" GOMODCACHE="$(GO_MOD_CACHE)" go run cmd/api/main.go

up:
	@$(COMPOSE) up -d --build

infra-up:
	@$(COMPOSE) up -d postgres-primary postgres-replica redis

down:
	@$(COMPOSE) down -v --remove-orphans

logs:
	@$(COMPOSE) logs -f --tail=200

ps:
	@$(COMPOSE) ps

test:
	@GOCACHE="$(GO_CACHE)" GOMODCACHE="$(GO_MOD_CACHE)" go test ./...

test-race:
	@GOCACHE="$(GO_CACHE)" GOMODCACHE="$(GO_MOD_CACHE)" go test -race ./...

lint:
	@test -x "$(GOLANGCI_LINT)" || go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	@GOLANGCI_LINT_CACHE="$(GOLANGCI_LINT_CACHE)" GOCACHE="$(GO_CACHE)" GOMODCACHE="$(GO_MOD_CACHE)" "$(GOLANGCI_LINT)" run ./...

lint-fix:
	@test -x "$(GOLANGCI_LINT)" || go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	@GOLANGCI_LINT_CACHE="$(GOLANGCI_LINT_CACHE)" GOCACHE="$(GO_CACHE)" GOMODCACHE="$(GO_MOD_CACHE)" "$(GOLANGCI_LINT)" run --fix ./...

docs:
	@test -x "$(SWAG)" || go install github.com/swaggo/swag/cmd/swag@$(SWAG_VERSION)
	@$(SWAG) init -g cmd/api/main.go -o docs --parseInternal
	@rm -f docs/docs.go

docs-check:
	@$(MAKE) docs
	@git diff --exit-code -- docs/swagger.json docs/swagger.yaml

fmt:
	@gofmt -w $$(find . -path ./tmp -prune -o -path ./.tmp -prune -o -path ./.gocache -prune -o -path ./.gomodcache -prune -o -name '*.go' -print)

build:
	@mkdir -p "$(BUILD_DIR)"
	@GOCACHE="$(GO_CACHE)" GOMODCACHE="$(GO_MOD_CACHE)" CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o "$(BIN_PATH)" ./cmd/api

docker-build:
	@docker build -t $(APP_NAME):local .

tidy:
	@GOCACHE="$(GO_CACHE)" GOMODCACHE="$(GO_MOD_CACHE)" go mod tidy

check: fmt test lint docs-check

clean:
	@rm -rf "$(BUILD_DIR)"
