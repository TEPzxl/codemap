# codemap v0.3 Demo

This demo uses `examples/layered-service`, a small Go service with `cmd/api`, handler, service, and repository packages.

## Build

From the repository root:

```bash
make install
make web-build
make build
```

## Run

```bash
./bin/codemap serve ./examples/layered-service --port 8080
```

Open:

```text
http://localhost:8080
```

## Walkthrough

1. In `Entrypoints`, select the `main` function, or search `main.main` in `Entry symbol`.
2. Click `Load graph`.
3. The default function graph shows the local business chain from `main` through handler, service, and repository.
4. Use `Depth` and `Graph filters` to change traversal and visibility.
5. Click a node such as `UserService.CreateUser` to view its source in the bottom panel.
6. Click an edge to view the callsite line that produced that edge.
7. Check `Current graph` for node, edge, package, and resolution counts.
8. Select a node and use `Focus downstream`, `Focus upstream`, or `Focus neighborhood`.
9. Use `Path search` to find a static call path between two symbols.
10. Switch `Graph view` to `Package graph` to inspect package-level call aggregation.
11. Click `Rescan` after changing local source files.

## Interface Candidate Expansion

Run the interface fixture:

```bash
./bin/codemap serve ./examples/interface-call --port 8080
```

Open `http://localhost:8080`, select `main.main`, enable `Expand interface candidates`, and load the graph. Candidate implementation edges are shown only when this option is enabled; default behavior remains conservative.

The same behavior is available from the CLI:

```bash
./bin/codemap graph ./examples/interface-call --entry main.main --depth 5 --expand-interface
```

## Metadata And Rescan

The left panel shows package, symbol, call, warning, analysis duration, and version metadata. Click `Rescan` after changing local source files; successful rescan replaces the cached index atomically. If rescan fails, the previous graph remains available.

API equivalents:

```bash
curl -s http://localhost:8080/api/meta | python -m json.tool
curl -s -X POST http://localhost:8080/api/rescan | python -m json.tool
```

## Real Project Baseline

真实项目 demo 和 v0.3 baseline 记录见：[real-projects.md](real-projects.md)。

## Current Limits

- Only local Go projects are supported.
- `_test.go` files are ignored by default.
- Interface candidate expansion is static and conservative.
- Dynamic calls through function variables may be marked `unresolved`.
- Standard library and third-party calls are hidden unless filters enable them.
- The graph is a static approximation, not a runtime trace.
- Path search uses the static call graph and does not prove a runtime path.
- Package graph is aggregated from function and method calls, not import declarations.
