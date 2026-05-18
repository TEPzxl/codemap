import test from "node:test";
import assert from "node:assert/strict";

import { summarizeGraph } from "./graphSummary";
import type { Graph } from "@/types/graph";

const graph: Graph = {
  entry: "example/main.main",
  nodes: [
    {
      id: "example/main.main",
      label: "main",
      kind: "function",
      package: "example/main",
      file: "main.go",
      start_line: 1,
      end_line: 8,
      is_external: false,
    },
    {
      id: "example/service.Run",
      label: "Run",
      kind: "function",
      package: "example/service",
      file: "service.go",
      start_line: 10,
      end_line: 20,
      is_external: false,
    },
    {
      id: "fmt.Println",
      label: "Println",
      kind: "external",
      package: "fmt",
      file: "",
      start_line: 0,
      end_line: 0,
      is_external: true,
    },
  ],
  edges: [
    {
      id: "e1",
      from: "example/main.main",
      to: "example/service.Run",
      kind: "call",
      resolution: "resolved",
      callsite: { file: "main.go", line: 4, column: 2 },
    },
    {
      id: "e2",
      from: "example/service.Run",
      to: "example/repo.Interface.Save",
      kind: "call",
      resolution: "interface",
      callsite: { file: "service.go", line: 14, column: 8 },
    },
    {
      id: "e3",
      from: "example/service.Run",
      to: "fmt.Println",
      kind: "call",
      resolution: "external",
      callsite: { file: "service.go", line: 15, column: 8 },
    },
    {
      id: "e4",
      from: "example/service.Run",
      to: "unresolved",
      kind: "call",
      resolution: "unresolved",
      callsite: { file: "service.go", line: 16, column: 8 },
    },
  ],
  warnings: [],
};

test("summarizeGraph counts nodes, edges, packages, and edge resolutions", () => {
  assert.deepEqual(summarizeGraph(graph), {
    nodes: 3,
    edges: 4,
    packages: 3,
    resolvedEdges: 1,
    interfaceEdges: 1,
    externalEdges: 1,
    unresolvedEdges: 1,
  });
});

test("summarizeGraph returns zero counts without a graph", () => {
  assert.deepEqual(summarizeGraph(null), {
    nodes: 0,
    edges: 0,
    packages: 0,
    resolvedEdges: 0,
    interfaceEdges: 0,
    externalEdges: 0,
    unresolvedEdges: 0,
  });
});
