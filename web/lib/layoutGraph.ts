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
    const depth = depthByNode.get(node.id) ?? fallbackDepth(node, depthByNode);
    const bucket = buckets.get(depth) ?? [];
    bucket.push(node);
    buckets.set(depth, bucket);
  }

  for (const bucket of buckets.values()) {
    bucket.sort(compareNodes);
  }

  return [...graph.nodes].sort(compareNodes).map((node) => {
    const depth = depthByNode.get(node.id) ?? fallbackDepth(node, depthByNode);
    const bucket = buckets.get(depth) ?? [];
    const index = bucket.findIndex((item) => item.id === node.id);
    const centeredIndex = Math.max(index, 0) - (bucket.length - 1) / 2;
    return {
      ...node,
      position: {
        x: depth * 360 + 80,
        y: centeredIndex * 170 + 300,
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

function fallbackDepth(node: GraphNode, depthByNode: Map<string, number>): number {
  if (node.id === node.package || depthByNode.size === 0) {
    return 0;
  }
  return Math.max(...depthByNode.values(), 0) + 1;
}

function compareNodes(a: GraphNode, b: GraphNode): number {
  if (a.package !== b.package) {
    return a.package.localeCompare(b.package);
  }
  if (a.kind !== b.kind) {
    return a.kind.localeCompare(b.kind);
  }
  if (a.label !== b.label) {
    return a.label.localeCompare(b.label);
  }
  return a.id.localeCompare(b.id);
}
