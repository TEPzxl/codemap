import type { Graph } from "@/types/graph";

export function filterGraphByPackage(graph: Graph, packagePrefix: string): Graph {
  const nodes = graph.nodes.filter((node) => node.package === packagePrefix || node.package.startsWith(`${packagePrefix}/`));
  const nodeIDs = new Set(nodes.map((node) => node.id));
  const edges = graph.edges.filter((edge) => nodeIDs.has(edge.from) && nodeIDs.has(edge.to));
  const entry = nodeIDs.has(graph.entry) ? graph.entry : (nodes[0]?.id ?? graph.entry);

  return {
    entry,
    nodes,
    edges,
    warnings: graph.warnings,
  };
}
