import type { Graph, Node as GraphNode } from "@/types/graph";

export function searchCurrentGraphNodes(graph: Graph | null, query: string, limit = 12): GraphNode[] {
  const normalizedQuery = query.trim().toLowerCase();
  if (!graph || normalizedQuery === "" || limit <= 0) {
    return [];
  }

  return graph.nodes
    .filter((node) => nodeMatchesCurrentGraphQuery(node, normalizedQuery))
    .sort(compareSearchResults(normalizedQuery))
    .slice(0, limit);
}

function nodeMatchesCurrentGraphQuery(node: GraphNode, normalizedQuery: string): boolean {
  return searchableNodeFields(node).some((value) => value.toLowerCase().includes(normalizedQuery));
}

function compareSearchResults(query: string): (left: GraphNode, right: GraphNode) => number {
  return (left, right) => {
    const leftRank = matchRank(left, query);
    const rightRank = matchRank(right, query);
    if (leftRank !== rightRank) {
      return leftRank - rightRank;
    }
    if (left.package !== right.package) {
      return left.package.localeCompare(right.package);
    }
    if (left.label !== right.label) {
      return left.label.localeCompare(right.label);
    }
    return left.id.localeCompare(right.id);
  };
}

function matchRank(node: GraphNode, query: string): number {
  const fields = searchableNodeFields(node).map((field) => field.toLowerCase());
  if (fields.some((field) => field === query)) {
    return 0;
  }
  if (node.label.toLowerCase().startsWith(query)) {
    return 1;
  }
  if (node.id.toLowerCase().includes(query)) {
    return 2;
  }
  if (node.package.toLowerCase().includes(query)) {
    return 3;
  }
  if (node.file.toLowerCase().includes(query)) {
    return 4;
  }
  return 5;
}

function searchableNodeFields(node: GraphNode): string[] {
  return [node.label, node.id, node.package, node.file].filter(Boolean);
}
