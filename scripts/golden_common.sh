#!/usr/bin/env bash
set -euo pipefail

CODEMAP_REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CODEMAP_GOLDEN_DIR="$CODEMAP_REPO_ROOT/docs/golden/v0.1"
CODEMAP_GOLDEN_TMP=""

golden_setup() {
	CODEMAP_GOLDEN_TMP="$(mktemp -d)"
	trap 'rm -rf "$CODEMAP_GOLDEN_TMP"' EXIT

	cat > "$CODEMAP_GOLDEN_TMP/canonicalize_json.go" <<'GO'
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
)

func main() {
	dec := json.NewDecoder(os.Stdin)
	dec.UseNumber()

	var payload any
	if err := dec.Decode(&payload); err != nil {
		fmt.Fprintf(os.Stderr, "decode json: %v\n", err)
		os.Exit(1)
	}

	payload = canonicalize("", payload)

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(payload); err != nil {
		fmt.Fprintf(os.Stderr, "encode json: %v\n", err)
		os.Exit(1)
	}
}

func canonicalize(parent string, value any) any {
	switch v := value.(type) {
	case map[string]any:
		for key, item := range v {
			v[key] = canonicalize(key, item)
		}
		return v
	case []any:
		for i, item := range v {
			v[i] = canonicalize(parent, item)
		}
		if shouldSort(parent) {
			sort.SliceStable(v, func(i, j int) bool {
				return sortKey(parent, v[i]) < sortKey(parent, v[j])
			})
		}
		return v
	default:
		return value
	}
}

func shouldSort(parent string) bool {
	switch parent {
	case "packages", "symbols", "calls", "nodes", "edges", "warnings":
		return true
	default:
		return false
	}
}

func sortKey(parent string, value any) string {
	obj, ok := value.(map[string]any)
	if !ok {
		return jsonString(value)
	}

	switch parent {
	case "packages":
		return joinFields(obj, "id", "pkg", "path", "name", "dir")
	case "symbols", "nodes":
		return joinFields(obj, "id", "package", "receiver", "name", "label", "file", "start_line", "end_line")
	case "calls", "edges":
		callsite, _ := obj["callsite"].(map[string]any)
		return join(
			field(obj, "from"),
			field(obj, "to"),
			field(obj, "resolution"),
			field(obj, "kind"),
			field(callsite, "file"),
			field(callsite, "line"),
			field(callsite, "column"),
			field(obj, "id"),
		)
	case "warnings":
		return joinFields(obj, "package", "file", "line", "message", "error")
	default:
		return jsonString(value)
	}
}

func joinFields(obj map[string]any, keys ...string) string {
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, field(obj, key))
	}
	return join(parts...)
}

func join(parts ...string) string {
	var buf bytes.Buffer
	for i, part := range parts {
		if i > 0 {
			buf.WriteByte('\x00')
		}
		buf.WriteString(part)
	}
	return buf.String()
}

func field(obj map[string]any, key string) string {
	if obj == nil {
		return ""
	}
	value, ok := obj[key]
	if !ok {
		return ""
	}
	switch v := value.(type) {
	case string:
		return v
	case json.Number:
		return v.String()
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	default:
		return jsonString(v)
	}
}

func jsonString(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprintf("%v", value)
	}
	return string(data)
}
GO
}

canonicalize_json() {
	go run "$CODEMAP_GOLDEN_TMP/canonicalize_json.go"
}

run_codemap_json() {
	go run ./cmd/codemap "$@" | canonicalize_json
}

write_golden() {
	local rel_path="$1"
	shift

	mkdir -p "$CODEMAP_GOLDEN_DIR/$(dirname "$rel_path")"
	run_codemap_json "$@" > "$CODEMAP_GOLDEN_DIR/$rel_path"
	echo "wrote docs/golden/v0.1/$rel_path"
}

verify_golden() {
	local rel_path="$1"
	shift

	local actual="$CODEMAP_GOLDEN_TMP/${rel_path//\//_}"
	run_codemap_json "$@" > "$actual"
	diff -u "$CODEMAP_GOLDEN_DIR/$rel_path" "$actual"
	echo "verified docs/golden/v0.1/$rel_path"
}

