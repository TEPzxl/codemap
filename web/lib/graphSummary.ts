import type { EdgeResolution, Graph } from "@/types/graph";

export interface GraphSummary {
  nodes: number;
  edges: number;
  packages: number;
  resolvedEdges: number;
  interfaceEdges: number;
  externalEdges: number;
  unresolvedEdges: number;
}

const emptySummary: GraphSummary = {
  nodes: 0,
  edges: 0,
  packages: 0,
  resolvedEdges: 0,
  interfaceEdges: 0,
  externalEdges: 0,
  unresolvedEdges: 0,
};

export function summarizeGraph(graph: Graph | null): GraphSummary {
  if (!graph) {
    return emptySummary;
  }

  const packages = new Set<string>();
  for (const node of graph.nodes) {
    if (node.package) {
      packages.add(node.package);
    }
  }

  const resolutionCounts = graph.edges.reduce<Record<EdgeResolution, number>>(
    (counts, edge) => {
      counts[edge.resolution] += 1;
      return counts;
    },
    {
      resolved: 0,
      interface: 0,
      external: 0,
      unresolved: 0,
    },
  );

  return {
    nodes: graph.nodes.length,
    edges: graph.edges.length,
    packages: packages.size,
    resolvedEdges: resolutionCounts.resolved,
    interfaceEdges: resolutionCounts.interface,
    externalEdges: resolutionCounts.external,
    unresolvedEdges: resolutionCounts.unresolved,
  };
}
