import test from "node:test";
import assert from "node:assert/strict";

import { filterGraphByPackage } from "./packageDrilldown";
import type { Graph } from "@/types/graph";

const graph: Graph = {
  entry: "example/cmd.main",
  nodes: [
    node("example/cmd.main", "example/cmd"),
    node("example/internal/service.Run", "example/internal/service"),
    node("example/internal/service.validate", "example/internal/service"),
    node("example/internal/repo.Save", "example/internal/repo"),
  ],
  edges: [
    edge("e1", "example/cmd.main", "example/internal/service.Run"),
    edge("e2", "example/internal/service.Run", "example/internal/service.validate"),
    edge("e3", "example/internal/service.Run", "example/internal/repo.Save"),
  ],
  warnings: [],
};

test("filterGraphByPackage keeps all matching package nodes and internal edges", () => {
  const filtered = filterGraphByPackage(graph, "example/internal/service");

  assert.equal(filtered.entry, "example/internal/service.Run");
  assert.deepEqual(
    filtered.nodes.map((item) => item.id),
    ["example/internal/service.Run", "example/internal/service.validate"],
  );
  assert.deepEqual(
    filtered.edges.map((item) => item.id),
    ["e2"],
  );
});

function node(id: string, pkg: string): Graph["nodes"][number] {
  return {
    id,
    label: id.split(".").at(-1) ?? id,
    kind: "function",
    package: pkg,
    file: `${pkg}/file.go`,
    start_line: 1,
    end_line: 2,
    is_external: false,
  };
}

function edge(id: string, from: string, to: string): Graph["edges"][number] {
  return {
    id,
    from,
    to,
    kind: "call",
    resolution: "resolved",
    callsite: { file: "file.go", line: 1, column: 1 },
  };
}
