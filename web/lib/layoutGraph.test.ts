import test from "node:test";
import assert from "node:assert/strict";

import { layoutGraph } from "./layoutGraph";
import type { Edge, Graph, Node as GraphNode } from "@/types/graph";

test("layoutGraph orders callees from the same caller by callsite line", () => {
  const graph = makeGraph(
    [
      node("entry", "main"),
      node("later-label-but-earlier-call", "ZNewCluster"),
      node("earlier-label-but-later-call", "AServerID"),
    ],
    [
      edge("e1", "entry", "later-label-but-earlier-call", "main.go", 12, 2),
      edge("e2", "entry", "earlier-label-but-later-call", "main.go", 28, 2),
    ],
  );

  const positioned = layoutGraph(graph);

  assert(
    positionOf(positioned, "later-label-but-earlier-call").y <
      positionOf(positioned, "earlier-label-but-later-call").y,
  );
});

test("layoutGraph keeps entry direct callees in source order on the same layer", () => {
  const graph = makeGraph(
    [
      node("entry", "main"),
      node("new-cluster", "NewCluster"),
      node("cluster-stop", "Cluster.Stop"),
      node("wait-leader", "Cluster.WaitLeader"),
      node("server-id", "Server.ID"),
      node("cluster-put", "Cluster.Put"),
      node("cluster-get", "Cluster.Get"),
    ],
    [
      edge("e-server-id", "entry", "server-id", "cmd/demo/main.go", 25, 11),
      edge("e-cluster-get", "entry", "cluster-get", "cmd/demo/main.go", 31, 14),
      edge("e-new-cluster", "entry", "new-cluster", "cmd/demo/main.go", 12, 18),
      edge("e-cluster-stop", "entry", "cluster-stop", "cmd/demo/main.go", 16, 16),
      edge("e-cluster-put", "entry", "cluster-put", "cmd/demo/main.go", 27, 14),
      edge("e-wait-leader", "entry", "wait-leader", "cmd/demo/main.go", 21, 22),
    ],
  );

  const positioned = layoutGraph(graph);
  const directCallees = ["new-cluster", "cluster-stop", "wait-leader", "server-id", "cluster-put", "cluster-get"];

  assert.equal(positionOf(positioned, "entry").x, 80);
  for (const id of directCallees) {
    assert.equal(positionOf(positioned, id).x, 440);
  }
  assertYOrder(positioned, directCallees);
});

test("layoutGraph lays out downstream subtree to the right without overriding parent source order", () => {
  const graph = makeGraph(
    [
      node("entry", "main"),
      node("first-parent", "FirstParent"),
      node("second-parent", "SecondParent"),
      node("first-child", "FirstChild"),
      node("second-child", "SecondChild"),
    ],
    [
      edge("e1", "entry", "first-parent", "main.go", 10, 1),
      edge("e2", "entry", "second-parent", "main.go", 20, 1),
      edge("e3", "first-parent", "first-child", "shared.go", 90, 1),
      edge("e4", "second-parent", "second-child", "shared.go", 1, 1),
    ],
  );

  const positioned = layoutGraph(graph);

  assert(positionOf(positioned, "first-child").x > positionOf(positioned, "first-parent").x);
  assert(positionOf(positioned, "second-child").x > positionOf(positioned, "second-parent").x);
  assert(positionOf(positioned, "first-child").y < positionOf(positioned, "second-parent").y);
  assert(positionOf(positioned, "first-child").y < positionOf(positioned, "second-child").y);
});

test("layoutGraph falls back to stable node ordering when callsite is missing", () => {
  const graph = makeGraph(
    [node("entry", "main"), node("b", "B"), node("a", "A")],
    [edge("e1", "entry", "b", "", 0, 0), edge("e2", "entry", "a", "", 0, 0)],
  );

  const positioned = layoutGraph(graph);

  assert(positionOf(positioned, "a").y < positionOf(positioned, "b").y);
});

test("layoutGraph handles nodes with multiple callers without duplicates or infinite layout", () => {
  const graph = makeGraph(
    [node("entry", "main"), node("caller-a", "CallerA"), node("caller-b", "CallerB"), node("shared", "Shared")],
    [
      edge("e1", "entry", "caller-a", "main.go", 10, 1),
      edge("e2", "entry", "caller-b", "main.go", 20, 1),
      edge("e3", "caller-b", "shared", "worker.go", 40, 1),
      edge("e4", "caller-a", "shared", "worker.go", 30, 1),
      edge("e5", "shared", "caller-a", "worker.go", 50, 1),
    ],
  );

  const positioned = layoutGraph(graph);

  assert.equal(positioned.length, 4);
  assert.equal(new Set(positioned.map((item) => item.id)).size, 4);
  assert(positioned.every((item) => Number.isFinite(item.position.x) && Number.isFinite(item.position.y)));
});

function positionOf(positioned: ReturnType<typeof layoutGraph>, id: string): { x: number; y: number } {
  const item = positioned.find((node) => node.id === id);
  assert(item, `missing positioned node ${id}`);
  return item.position;
}

function assertYOrder(positioned: ReturnType<typeof layoutGraph>, ids: string[]): void {
  for (let index = 1; index < ids.length; index += 1) {
    assert(
      positionOf(positioned, ids[index - 1]).y < positionOf(positioned, ids[index]).y,
      `${ids[index - 1]} should be above ${ids[index]}`,
    );
  }
}

function makeGraph(nodes: GraphNode[], edges: Edge[]): Graph {
  return {
    entry: "entry",
    nodes,
    edges,
    warnings: [],
  };
}

function node(id: string, label: string): GraphNode {
  return {
    id,
    label,
    kind: "function",
    package: "example",
    file: `${label}.go`,
    start_line: 1,
    end_line: 2,
    is_external: false,
  };
}

function edge(id: string, from: string, to: string, file: string, line: number, column: number): Edge {
  return {
    id,
    from,
    to,
    kind: "call",
    resolution: "resolved",
    callsite: { file, line, column },
  };
}
