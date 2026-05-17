import type { Edge, Graph, Node as GraphNode } from "@/types/graph";

export interface PositionedNode extends GraphNode {
  position: {
    x: number;
    y: number;
  };
}

export function layoutGraph(graph: Graph): PositionedNode[] {
  const depthByNode = computeDepths(graph.entry, graph.edges);
  const buckets = new Map<number, GraphNode[]>();

  for (const node of graph.nodes) {
    const depth = depthByNode.get(node.id) ?? 0;
    const bucket = buckets.get(depth) ?? [];
    bucket.push(node);
    buckets.set(depth, bucket);
  }

  for (const bucket of buckets.values()) {
    bucket.sort((a, b) => a.label.localeCompare(b.label));
  }

  return graph.nodes.map((node) => {
    const depth = depthByNode.get(node.id) ?? 0;
    const bucket = buckets.get(depth) ?? [];
    const index = bucket.findIndex((item) => item.id === node.id);
    return {
      ...node,
      position: {
        x: depth * 310,
        y: Math.max(index, 0) * 140,
      },
    };
  });
}

function computeDepths(entry: string, edges: Edge[]): Map<string, number> {
  const adjacency = new Map<string, string[]>();
  for (const edge of edges) {
    const targets = adjacency.get(edge.from) ?? [];
    targets.push(edge.to);
    adjacency.set(edge.from, targets);
  }

  const depths = new Map<string, number>([[entry, 0]]);
  const queue = [entry];
  while (queue.length > 0) {
    const current = queue.shift();
    if (!current) {
      continue;
    }
    const nextDepth = (depths.get(current) ?? 0) + 1;
    for (const target of adjacency.get(current) ?? []) {
      const previous = depths.get(target);
      if (previous === undefined || nextDepth < previous) {
        depths.set(target, nextDepth);
        queue.push(target);
      }
    }
  }
  return depths;
}
