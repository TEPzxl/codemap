"use client";

import { Background, BackgroundVariant, Controls, ReactFlow, type Edge as FlowEdge, type Node as FlowNode } from "@xyflow/react";
import "@xyflow/react/dist/style.css";

import { layoutGraph } from "@/lib/layoutGraph";
import { displayPackage, displaySymbolID } from "@/lib/displaySymbol";
import type { Graph, Node as GraphNode } from "@/types/graph";

interface GraphViewProps {
  graph: Graph | null;
  selectedNode: GraphNode | null;
  modulePrefix: string;
  loading?: boolean;
  error?: string | null;
  onNodeSelect: (node: GraphNode) => void;
}

export function GraphView({ graph, selectedNode, modulePrefix, loading, error, onNodeSelect }: GraphViewProps) {
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
    className: nodeClassName(node, selectedNode?.id === node.id),
  }));
  const edges: FlowEdge[] = graph.edges.map((edge) => ({
    id: edge.id,
    source: edge.from,
    target: edge.to,
    type: "smoothstep",
    animated: edge.resolution === "interface" || edge.resolution === "unresolved",
    label: edge.candidate ? "candidate" : edge.resolution,
    style: {
      stroke: edge.resolution === "resolved" ? "#58745f" : "#d95f3d",
      strokeWidth: 1.5,
      strokeDasharray: edge.candidate ? "6 4" : undefined,
    },
    labelStyle: {
      fill: "#52616b",
      fontSize: 11,
      fontWeight: 600,
    },
  }));

  return (
    <div className="relative h-full min-h-[520px]">
      <SelectedNodeOverlay node={selectedNode} modulePrefix={modulePrefix} />
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
    </div>
  );
}

function nodeClassName(node: GraphNode, selected: boolean): string {
  const base = "!rounded-md !border !px-3 !py-2 !shadow-sm";
  const tone = node.is_external ? "!border-signal/40 !bg-orange-50" : "!border-moss/30 !bg-white";
  const selectedTone = selected ? "!border-moss !bg-green-50 !shadow-[0_0_0_3px_rgba(88,116,95,0.25)]" : "";
  return `${base} ${tone} ${selectedTone}`;
}

function SelectedNodeOverlay({ node, modulePrefix }: { node: GraphNode | null; modulePrefix: string }) {
  if (!node) {
    return null;
  }
  return (
    <div className="pointer-events-none absolute left-4 top-4 z-10 max-w-[min(520px,calc(100%-2rem))] rounded-md border border-line bg-white/95 px-3 py-2 shadow-sm backdrop-blur">
      <p className="text-sm font-semibold text-ink">{node.label}</p>
      <p className="mt-1 break-all font-mono text-xs leading-5 text-steel" title={node.id}>
        {displaySymbolID(node.id, modulePrefix)}
      </p>
      <p className="mt-1 break-all text-xs text-steel">{node.file || displayPackage(node.package, modulePrefix)}</p>
    </div>
  );
}

function GraphState({ message, tone = "muted" }: { message: string; tone?: "muted" | "error" }) {
  return (
    <div className="grid h-full place-items-center bg-paper">
      <p className={tone === "error" ? "max-w-md text-sm text-signal" : "text-sm text-steel"}>{message}</p>
    </div>
  );
}
