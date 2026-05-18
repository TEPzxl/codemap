SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c

API_PATH ?= ./examples/layered-service
API_HOST ?= localhost
API_PORT ?= 18080
WEB_HOST ?= 127.0.0.1
WEB_PORT ?= 3000
WEB_DIR := web
WEB_OUT := $(WEB_DIR)/out
SERVER_STATIC := internal/server/static
BIN_DIR := bin
DIST_DIR := dist
PNPM := corepack pnpm

.PHONY: help install test check smoke test-go test-web web-lint web-typecheck lint-web build-web web-build generate-golden verify-golden build release demo dev-api dev-web dev

help:
	@echo "codemap make targets"
	@echo ""
	@echo "  make install     Install frontend dependencies with pnpm"
	@echo "  make test        Run the full quality gate"
	@echo "  make check       Run Go tests, CLI smoke, frontend lint/typecheck/build"
	@echo "  make smoke       Run CLI smoke tests against fixture projects"
	@echo "  make test-go     Run go test ./..."
	@echo "  make verify-golden  Verify CLI outputs against v0.1 golden files"
	@echo "  make generate-golden  Regenerate v0.1 golden files"
	@echo "  make test-web    Run frontend lint, typecheck, and build"
	@echo "  make web-lint    Run frontend lint"
	@echo "  make web-typecheck  Run frontend TypeScript check"
	@echo "  make web-build   Build Next static export and stage assets for Go embed"
	@echo "  make build       Build Go binary into ./bin/codemap"
	@echo "  make release     Build release binaries into ./dist"
	@echo "  make demo        Run ./bin/codemap serve with the demo fixture"
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

test: check

check: test-go smoke verify-golden web-lint web-typecheck build-web

smoke:
	./scripts/smoke.sh

generate-golden:
	./scripts/generate_golden.sh

verify-golden:
	./scripts/verify_golden.sh

test-go:
	go test ./...

test-web: web-lint web-typecheck build-web

web-lint:
	cd $(WEB_DIR) && $(PNPM) lint

web-typecheck:
	cd $(WEB_DIR) && $(PNPM) typecheck

lint-web: web-lint web-typecheck

build-web:
	cd $(WEB_DIR) && $(PNPM) build

web-build: build-web
	rm -rf $(SERVER_STATIC)
	mkdir -p $(SERVER_STATIC)
	cp -R $(WEB_OUT)/. $(SERVER_STATIC)/

build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/codemap ./cmd/codemap

release: web-build
	rm -rf $(DIST_DIR)
	mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o $(DIST_DIR)/codemap-linux-amd64 ./cmd/codemap
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o $(DIST_DIR)/codemap-darwin-arm64 ./cmd/codemap
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o $(DIST_DIR)/codemap-darwin-amd64 ./cmd/codemap

demo: build
	./$(BIN_DIR)/codemap serve $(API_PATH) --port $(API_PORT)

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
