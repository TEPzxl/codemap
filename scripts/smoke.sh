#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

cat > "$tmpdir/jsoncheck.go" <<'GO'
package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: jsoncheck <mode> <file>")
		os.Exit(2)
	}

	data, err := os.ReadFile(os.Args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "read json: %v\n", err)
		os.Exit(1)
	}

	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		fmt.Fprintf(os.Stderr, "invalid json: %v\n%s\n", err, string(data))
		os.Exit(1)
	}

	mode := os.Args[1]
	switch mode {
	case "scan":
		requireArray(payload, "packages")
	case "symbols":
		requireArray(payload, "symbols")
	case "calls":
		requireArray(payload, "calls")
	case "entrypoints":
		requireArray(payload, "entrypoints")
	case "graph":
		requireString(payload, "entry")
		requireArray(payload, "nodes")
		requireArray(payload, "edges")
	case "packages":
		requireArray(payload, "nodes")
		requireArray(payload, "edges")
	default:
		fmt.Fprintf(os.Stderr, "unknown mode: %s\n", mode)
		os.Exit(2)
	}
}

func requireArray(payload map[string]any, key string) {
	value, ok := payload[key]
	if !ok {
		fmt.Fprintf(os.Stderr, "missing %q\n", key)
		os.Exit(1)
	}
	items, ok := value.([]any)
	if !ok || len(items) == 0 {
		fmt.Fprintf(os.Stderr, "%q must be a non-empty array\n", key)
		os.Exit(1)
	}
}

func requireString(payload map[string]any, key string) {
	value, ok := payload[key].(string)
	if !ok || value == "" {
		fmt.Fprintf(os.Stderr, "%q must be a non-empty string\n", key)
		os.Exit(1)
	}
}
GO

run_json() {
	local mode="$1"
	local name="$2"
	shift 2

	local output="$tmpdir/$name.json"
	go run ./cmd/codemap "$@" > "$output"
	go run "$tmpdir/jsoncheck.go" "$mode" "$output"
}

run_json scan simple-scan scan ./examples/simple
run_json scan layered-service-scan scan ./examples/layered-service
run_json scan interface-call-scan scan ./examples/interface-call
run_json symbols simple-symbols symbols ./examples/simple
run_json calls simple-calls calls ./examples/simple
run_json entrypoints layered-service-entrypoints entrypoints ./examples/layered-service
run_json graph layered-service-graph graph ./examples/layered-service --entry main.main --depth 5
run_json packages layered-service-packages packages ./examples/layered-service --entry main.main --depth 5
