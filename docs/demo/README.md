# codemap v0.2 Demo

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

1. In `Entry symbol`, search for `main.main` or select the `main` symbol from `cmd/api/main.go`.
2. Click `Load graph`.
3. The default graph shows the local business chain from `main` through handler, service, and repository.
4. Click a node such as `UserService.CreateUser` to view its source in the bottom panel.
5. Click an edge to view the callsite line that produced that edge.
6. Use `Depth` to limit traversal.
7. Use graph filters to show external calls, unresolved calls, interface calls, or expanded interface candidates.

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

## Screenshot

![codemap UI](codemap-ui.png)

## Current Limits

- Only local Go projects are supported.
- `_test.go` files are ignored by default.
- Interface candidate expansion is static and conservative.
- Dynamic calls through function variables may be marked `unresolved`.
- Standard library and third-party calls are hidden unless filters enable them.
- The graph is a static approximation, not a runtime trace.
