import type { Edge, Graph, Node as GraphNode } from "@/types/graph";

export interface PositionedNode extends GraphNode {
  position: {
    x: number;
    y: number;
  };
}

const ENTRY_X = 80;
const DEPTH_GAP = 360;
const TOP_Y = 120;
const ROW_GAP = 170;

export function layoutGraph(graph: Graph): PositionedNode[] {
  const nodeByID = new Map(graph.nodes.map((node) => [node.id, node]));
  const outgoing = buildOutgoingEdges(graph.edges, nodeByID);
  const placementByNode = computeSourceOrderPlacement(graph.entry, graph.nodes, outgoing, nodeByID);

  return graph.nodes.map((node) => {
    const placement = placementByNode.get(node.id) ?? { depth: 0, row: 0 };
    return {
      ...node,
      position: {
        x: ENTRY_X + placement.depth * DEPTH_GAP,
        y: TOP_Y + placement.row * ROW_GAP,
      },
    };
  });
}

interface Placement {
  depth: number;
  row: number;
}

function computeSourceOrderPlacement(
  entry: string,
  nodes: GraphNode[],
  outgoing: Map<string, Edge[]>,
  nodeByID: Map<string, GraphNode>,
): Map<string, Placement> {
  const placements = new Map<string, Placement>();
  const nextRowByDepth = new Map<number, number>();
  const queue: string[] = [];

  const enqueue = (nodeID: string, depth: number, preferredRow: number): void => {
    if (!nodeByID.has(nodeID) || placements.has(nodeID)) {
      return;
    }
    const row = Math.max(nextRowByDepth.get(depth) ?? 0, preferredRow);
    placements.set(nodeID, { depth, row });
    nextRowByDepth.set(depth, row + 1);
    queue.push(nodeID);
  };

  const drainQueue = (): void => {
    while (queue.length > 0) {
      const current = queue.shift();
      if (!current) {
        continue;
      }
      const placement = placements.get(current);
      if (!placement) {
        continue;
      }
      for (const edge of outgoing.get(current) ?? []) {
        enqueue(edge.to, placement.depth + 1, placement.row);
      }
    }
  };

  enqueue(entry, 0, 0);
  drainQueue();

  const fallbackDepth = placements.size > 0 ? Math.max(...[...placements.values()].map((placement) => placement.depth)) + 1 : 0;
  const unvisited = nodes.filter((node) => !placements.has(node.id)).sort(compareNodes);
  for (const node of unvisited) {
    enqueue(node.id, fallbackDepth, 0);
    drainQueue();
  }

  return placements;
}

function buildOutgoingEdges(edges: Edge[], nodeByID: Map<string, GraphNode>): Map<string, Edge[]> {
  const outgoing = new Map<string, Edge[]>();
  for (const edge of edges) {
    const list = outgoing.get(edge.from) ?? [];
    list.push(edge);
    outgoing.set(edge.from, list);
  }

  for (const list of outgoing.values()) {
    list.sort(compareEdgesBySourceOrder(nodeByID));
  }
  return outgoing;
}

interface SourceOrderWeight {
  file: string;
  line: number;
  column: number;
  edgeID: string;
}

function compareEdgesBySourceOrder(nodeByID: Map<string, GraphNode>): (a: Edge, b: Edge) => number {
  return (a, b) => {
    const leftValid = hasValidCallsite(a);
    const rightValid = hasValidCallsite(b);
    if (leftValid && rightValid) {
      return compareSourceOrderWeight(sourceOrderWeight(a), sourceOrderWeight(b));
    }
    if (leftValid && !rightValid) {
      return -1;
    }
    if (!leftValid && rightValid) {
      return 1;
    }

    const leftNode = nodeByID.get(a.to);
    const rightNode = nodeByID.get(b.to);
    if (leftNode && rightNode) {
      const order = compareNodes(leftNode, rightNode);
      if (order !== 0) {
        return order;
      }
    }
    if (leftNode && !rightNode) {
      return -1;
    }
    if (!leftNode && rightNode) {
      return 1;
    }
    return a.id.localeCompare(b.id);
  };
}

function sourceOrderWeight(edge: Edge): SourceOrderWeight {
  return {
    file: edge.callsite.file,
    line: edge.callsite.line,
    column: edge.callsite.column,
    edgeID: edge.id,
  };
}

function compareSourceOrderWeight(a: SourceOrderWeight, b: SourceOrderWeight): number {
  if (a.file !== b.file) {
    return a.file.localeCompare(b.file);
  }
  if (a.line !== b.line) {
    return a.line - b.line;
  }
  if (a.column !== b.column) {
    return a.column - b.column;
  }
  return a.edgeID.localeCompare(b.edgeID);
}

function hasValidCallsite(edge: Edge): boolean {
  return edge.callsite.file !== "" && edge.callsite.line > 0 && edge.callsite.column > 0;
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
