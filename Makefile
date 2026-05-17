SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c

API_PATH ?= ./examples/layered-service
API_HOST ?= localhost
API_PORT ?= 18080
WEB_HOST ?= 127.0.0.1
WEB_PORT ?= 3000
WEB_DIR := web
PNPM := corepack pnpm

.PHONY: help install test test-go test-web lint-web build-web dev-api dev-web dev

help:
	@echo "codemap make targets"
	@echo ""
	@echo "  make install     Install frontend dependencies with pnpm"
	@echo "  make test        Run Go tests and frontend lint/typecheck/build"
	@echo "  make test-go     Run go test ./..."
	@echo "  make test-web    Run frontend lint, typecheck, and build"
	@echo "  make dev-api     Start Go API server"
	@echo "  make dev-web     Start Next dev server"
	@echo "  make dev         Start Go API in background, then Next dev server"
	@echo ""
	@echo "Variables:"
	@echo "  API_PATH=$(API_PATH)"
	@echo "  API_PORT=$(API_PORT)"
	@echo "  WEB_PORT=$(WEB_PORT)"

install:
	cd $(WEB_DIR) && $(PNPM) install

test: test-go test-web

test-go:
	go test ./...

test-web: lint-web build-web

lint-web:
	cd $(WEB_DIR) && $(PNPM) lint
	cd $(WEB_DIR) && $(PNPM) typecheck

build-web:
	cd $(WEB_DIR) && $(PNPM) build

dev-api:
	go run ./cmd/codemap serve $(API_PATH) --port $(API_PORT)

dev-web:
	cd $(WEB_DIR) && CODEMAP_API_BASE_URL=http://$(API_HOST):$(API_PORT) $(PNPM) dev --hostname $(WEB_HOST) --port $(WEB_PORT)

dev:
	@echo "Starting Go API on http://$(API_HOST):$(API_PORT)"
	@go run ./cmd/codemap serve $(API_PATH) --port $(API_PORT) & \
	api_pid=$$!; \
	trap 'kill $$api_pid 2>/dev/null || true' INT TERM EXIT; \
	sleep 1; \
	echo "Starting web UI on http://$(WEB_HOST):$(WEB_PORT)"; \
	cd $(WEB_DIR) && CODEMAP_API_BASE_URL=http://$(API_HOST):$(API_PORT) $(PNPM) dev --hostname $(WEB_HOST) --port $(WEB_PORT)
