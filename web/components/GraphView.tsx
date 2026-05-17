"use client";

import { Background, BackgroundVariant, Controls, ReactFlow, type Edge as FlowEdge, type Node as FlowNode } from "@xyflow/react";
import "@xyflow/react/dist/style.css";

import { layoutGraph } from "@/lib/layoutGraph";
import type { Graph, Node as GraphNode } from "@/types/graph";

interface GraphViewProps {
  graph: Graph | null;
  loading?: boolean;
  error?: string | null;
  onNodeSelect: (node: GraphNode) => void;
}

export function GraphView({ graph, loading, error, onNodeSelect }: GraphViewProps) {
  if (loading) {
    return <GraphState message="Loading graph" />;
  }
  if (error) {
    return <GraphState message={error} tone="error" />;
  }
  if (!graph || graph.nodes.length === 0) {
    return <GraphState message="No graph loaded" />;
  }

  const positioned = layoutGraph(graph);
  const nodeByID = new Map(graph.nodes.map((node) => [node.id, node]));
  const nodes: FlowNode[] = positioned.map((node) => ({
    id: node.id,
    position: node.position,
    data: {
      label: (
        <div className="grid gap-1">
          <span className="text-sm font-semibold text-ink">{node.label}</span>
          <span className="font-mono text-[10px] uppercase tracking-wide text-steel">{node.kind}</span>
        </div>
      ),
    },
    className: node.is_external
      ? "!rounded-md !border !border-signal/40 !bg-orange-50 !px-3 !py-2 !shadow-sm"
      : "!rounded-md !border !border-moss/30 !bg-white !px-3 !py-2 !shadow-sm",
  }));
  const edges: FlowEdge[] = graph.edges.map((edge) => ({
    id: edge.id,
    source: edge.from,
    target: edge.to,
    type: "smoothstep",
    animated: edge.resolution === "interface" || edge.resolution === "unresolved",
    label: edge.resolution,
    style: {
      stroke: edge.resolution === "resolved" ? "#58745f" : "#d95f3d",
      strokeWidth: 1.5,
    },
    labelStyle: {
      fill: "#52616b",
      fontSize: 11,
      fontWeight: 600,
    },
  }));

  return (
    <ReactFlow
      nodes={nodes}
      edges={edges}
      fitView
      minZoom={0.2}
      maxZoom={2}
      onNodeClick={(_, flowNode) => {
        const node = nodeByID.get(flowNode.id);
        if (node) {
          onNodeSelect(node);
        }
      }}
    >
      <Background variant={BackgroundVariant.Dots} gap={18} size={1} color="#d9d3c7" />
      <Controls position="bottom-left" />
    </ReactFlow>
  );
}

function GraphState({ message, tone = "muted" }: { message: string; tone?: "muted" | "error" }) {
  return (
    <div className="grid h-full place-items-center bg-paper">
      <p className={tone === "error" ? "max-w-md text-sm text-signal" : "text-sm text-steel"}>{message}</p>
    </div>
  );
}
