#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/golden_common.sh"

cd "$CODEMAP_REPO_ROOT"
golden_setup

write_golden simple.graph.json graph ./examples/simple --entry main.main --depth 5
write_golden layered-service.graph.json graph ./examples/layered-service --entry main.main --depth 5
write_golden interface-call.graph.json graph ./examples/interface-call --entry main.main --depth 5 --show-interface
write_golden simple.symbols.json symbols ./examples/simple
write_golden layered-service.calls.json calls ./examples/layered-service

