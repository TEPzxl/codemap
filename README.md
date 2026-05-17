# codemap

`codemap` is a local-first Go static analysis and call graph explorer. It scans a local Go module or workspace, extracts statically resolvable function and method calls, and serves an interactive web UI for inspecting the call chain and source snippets.

## Features

- Scan local Go modules with `go/packages`.
- Ignore `_test.go` files by default.
- Extract function and method symbols with stable IDs.
- Extract resolved, interface, external, and unresolved calls.
- Build depth-limited call graphs from an entry symbol.
- Serve a local HTTP API and static web UI from one Go binary.
- Click graph nodes to load source snippets lazily through `/api/source`.
- Preserve partial results and warnings when analysis is incomplete.

## Tech Stack

- Go, `go/packages`, `go/ast`, `go/types`, `net/http`, `embed`
- Next.js, React, TypeScript, Tailwind CSS, React Flow
- pnpm through Corepack for frontend dependency management

## Install

Prerequisites:

- Go 1.25 or newer
- Node.js with Corepack

Install frontend dependencies:

```bash
make install
```

## Build

Build the static web UI and stage it for Go embed:

```bash
make web-build
```

Build the Go binary:

```bash
make build
```

The binary is written to:

```text
bin/codemap
```

`go build` does not run pnpm. Run `make web-build` before `make build` when you want the binary to include the current frontend.

## Quick Demo

Run the layered-service demo:

```bash
make web-build
make build
./bin/codemap serve ./examples/layered-service --port 8080
```

Open:

```text
http://localhost:8080
```

In the UI:

1. Search or select `main.main`.
2. Click `Load graph`.
3. Inspect the `main -> handler -> service -> repository` call chain.
4. Click `UserService.CreateUser`.
5. Read the source snippet in the source panel.

Demo fixture:

```text
examples/layered-service
```

## CLI Examples

Scan packages:

```bash
go run ./cmd/codemap scan ./examples/simple
```

List symbols:

```bash
go run ./cmd/codemap symbols ./examples/layered-service
```

List calls:

```bash
go run ./cmd/codemap calls ./examples/layered-service
```

Build a graph:

```bash
go run ./cmd/codemap graph ./examples/layered-service --entry main.main --depth 5
```

Serve API and UI:

```bash
go run ./cmd/codemap serve ./examples/layered-service --port 8080
```

## Development

Run Go tests:

```bash
make test-go
```

Run frontend checks:

```bash
make test-web
```

Run the Go API and Next dev server separately:

```bash
make dev-api
make dev-web
```

Or run both together:

```bash
make dev
```

By default, development uses:

```text
Go API: http://localhost:18080
Web UI: http://127.0.0.1:3000
```

Override ports when needed:

```bash
make dev-web WEB_PORT=3001
make dev-api API_PORT=8080
```

## HTTP API

Health:

```bash
curl -s http://localhost:8080/api/health
```

Symbols:

```bash
curl -s http://localhost:8080/api/symbols
```

Graph:

```bash
curl -s "http://localhost:8080/api/graph?entry=main.main&depth=5"
```

Source:

```bash
curl -s "http://localhost:8080/api/source?node_id=<symbol-id>"
```

Warnings:

```bash
curl -s http://localhost:8080/api/warnings
```

## Architecture

```text
Local Go repo
  -> go/packages loader
  -> AST + type info
  -> symbols
  -> calls
  -> graph builder
  -> CLI / HTTP API
  -> Next + React Flow UI
  -> source snippet API
```

The frontend never scans local files and does not use Next API routes for core analysis. The Go server owns package loading, analysis, graph building, source reading, and API responses.

## Demo Screenshot

The screenshot below is captured from the static web UI served by `codemap serve`:

![codemap UI screenshot](docs/demo/codemap-ui.png)

## Current Limits

- Only local Go projects are supported.
- Remote GitHub URL scanning is not supported.
- `_test.go` files are ignored by default.
- Interface calls are handled conservatively and are not guaranteed to resolve to concrete implementations.
- Dynamic calls through function variables may be marked `unresolved`.
- The graph is a static approximation, not a runtime-precise call trace.
- Standard library and third-party calls are hidden from the default graph unless explicitly shown.
- No database, editor plugin, or LLM explanation layer is included in v1.

## Resume Summary

`codemap` is a local-first Go static analysis tool that builds a typed call graph from local repositories and serves an interactive React Flow UI from a single Go binary. It demonstrates practical compiler tooling with `go/packages`, AST/type analysis, graph traversal, local HTTP APIs, and a TypeScript frontend.
