SHELL := /bin/bash
.DEFAULT_GOAL := help

APP_NAME := delayedNotifier
COMPOSE := docker compose
GO := go
APP_PKG := ./cmd/main
GOLANGCI_LINT := golangci-lint

.PHONY: help
help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "Dev:"
	@echo "  make up              Start services"
	@echo "  make down            Stop everything"
	@echo "  make logs            Follow logs"
	@echo "  make ps              Show containers"
	@echo "  make restart         Restart service"
	@echo ""
	@echo "Migrations:"
	@echo "  make migrate-up      Run migrations (via migrator service)"
	@echo "  make migrate-down    Rollback 1 migration (via migrator service)"
	@echo ""
	@echo "Go:"
	@echo "  make tidy            go mod tidy"
	@echo "  make fmt             gofmt"
	@echo "  make test            go test ./..."
	@echo "  make build           Build service locally into ./bin/"
	@echo ""
	@echo "Lint/format:"
	@echo "  make lint            Run golangci-lint"
	@echo "  make lint-fix        Run golangci-lint with --fix"
	@echo "  make fmt-ci          Run golangci-lint fmt"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-build    Build service image"

.PHONY: up
up:
	$(COMPOSE) up -d --build

.PHONY: down
down:
	$(COMPOSE) down -v

.PHONY: ps
ps:
	$(COMPOSE) ps

.PHONY: logs
logs:
	$(COMPOSE) logs -f --tail=200

.PHONY: restart
restart:
	$(COMPOSE) restart delayedNotifier

.PHONY: migrate-up
migrate-up:
	$(COMPOSE) run --rm migrator up

.PHONY: migrate-down
migrate-down:
	$(COMPOSE) run --rm migrator down

.PHONY: tidy
tidy:
	$(GO) mod tidy

.PHONY: fmt
fmt:
	gofmt -w .

.PHONY: test
test:
	$(GO) test ./...

.PHONY: build
build:
	mkdir -p bin
	CGO_ENABLED=0 $(GO) build -o bin/delayedNotifier $(APP_PKG)

.PHONY: docker-build
docker-build:
	$(COMPOSE) build commentTree

.PHONY: lint
lint:
	$(GOLANGCI_LINT) run ./...

.PHONY: lint-fix
lint-fix:
	$(GOLANGCI_LINT) run --fix ./...

.PHONY: fmt-ci
fmt-ci:
	$(GOLANGCI_LINT) fmt ./...