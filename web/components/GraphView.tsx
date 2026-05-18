"use client";

import { useCallback, useEffect, useMemo } from "react";
import {
  Background,
  BackgroundVariant,
  Controls,
  MarkerType,
  MiniMap,
  Panel,
  ReactFlow,
  ReactFlowProvider,
  useEdgesState,
  useNodesState,
  useReactFlow,
  type Edge as FlowEdge,
  type Node as FlowNode,
} from "@xyflow/react";
import "@xyflow/react/dist/style.css";

import { layoutGraph } from "@/lib/layoutGraph";
import type { PositionedNode } from "@/lib/layoutGraph";
import { displayPackage, displaySymbolID } from "@/lib/displaySymbol";
import type { Edge as GraphEdge, Graph, Node as GraphNode } from "@/types/graph";

interface GraphViewProps {
  graph: Graph | null;
  selectedNode: GraphNode | null;
  selectedEdgeID: string | null;
  modulePrefix: string;
  loading?: boolean;
  error?: string | null;
  onNodeSelect: (node: GraphNode) => void;
  onEdgeSelect: (edge: GraphEdge) => void;
}

export function GraphView({
  graph,
  selectedNode,
  selectedEdgeID,
  modulePrefix,
  loading,
  error,
  onNodeSelect,
  onEdgeSelect,
}: GraphViewProps) {
  if (loading) {
    return <GraphState message="Loading graph" />;
  }
  if (error) {
    return <GraphState message={error} tone="error" />;
  }
  if (!graph || graph.nodes.length === 0) {
    return <GraphState message="No graph loaded" />;
  }

  return (
    <ReactFlowProvider>
      <GraphCanvas
        graph={graph}
        selectedNode={selectedNode}
        selectedEdgeID={selectedEdgeID}
        modulePrefix={modulePrefix}
        onNodeSelect={onNodeSelect}
        onEdgeSelect={onEdgeSelect}
      />
    </ReactFlowProvider>
  );
}

function GraphCanvas({
  graph,
  selectedNode,
  selectedEdgeID,
  modulePrefix,
  onNodeSelect,
  onEdgeSelect,
}: {
  graph: Graph;
  selectedNode: GraphNode | null;
  selectedEdgeID: string | null;
  modulePrefix: string;
  onNodeSelect: (node: GraphNode) => void;
  onEdgeSelect: (edge: GraphEdge) => void;
}) {
  const [nodes, setNodes, onNodesChange] = useNodesState<FlowNode>([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState<FlowEdge>([]);
  const { fitView } = useReactFlow();

  const nodeByID = useMemo(() => new Map(graph?.nodes.map((node) => [node.id, node]) ?? []), [graph]);
  const edgeByID = useMemo(() => new Map(graph.edges.map((edge) => [edge.id, edge])), [graph]);
  const connectedEdgeIDs = useMemo(() => {
    if (!graph || !selectedNode) {
      return new Set<string>();
    }
    return new Set(
      graph.edges.filter((edge) => edge.from === selectedNode.id || edge.to === selectedNode.id).map((edge) => edge.id),
    );
  }, [graph, selectedNode]);

  const applyLayout = useCallback(() => {
    if (!graph) {
      setNodes([]);
      setEdges([]);
      return;
    }
    const positioned = layoutGraph(graph);
    setNodes(positioned.map((node) => toFlowNode(node, selectedNode?.id === node.id)));
    setEdges(graph.edges.map((edge) => toFlowEdge(edge, connectedEdgeIDs.has(edge.id), selectedEdgeID === edge.id)));
  }, [connectedEdgeIDs, graph, selectedEdgeID, selectedNode, setEdges, setNodes]);

  useEffect(() => {
    applyLayout();
    const timeout = window.setTimeout(() => {
      void fitView({ padding: 0.2, duration: 250 });
    }, 0);
    return () => window.clearTimeout(timeout);
  }, [applyLayout, fitView]);

  return (
    <div className="relative h-full min-h-[520px]">
      <SelectedNodeOverlay node={selectedNode} modulePrefix={modulePrefix} />
      <ReactFlow
        nodes={nodes}
        edges={edges}
        fitView
        fitViewOptions={{ padding: 0.2 }}
        minZoom={0.2}
        maxZoom={2}
        nodesDraggable
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onPaneClick={() => {
          // Keep selection stable so the source panel remains useful while panning.
        }}
        onNodeClick={(_, flowNode) => {
          const node = nodeByID.get(flowNode.id);
          if (node) {
            onNodeSelect(node);
          }
        }}
        onEdgeClick={(event, flowEdge) => {
          event.stopPropagation();
          const edge = edgeByID.get(flowEdge.id);
          if (edge) {
            onEdgeSelect(edge);
          }
        }}
      >
        <Background variant={BackgroundVariant.Dots} gap={18} size={1} color="#d9d3c7" />
        <Controls position="bottom-left" />
        <MiniMap
          pannable
          zoomable
          position="bottom-right"
          nodeColor={(node) => miniMapNodeColor(node.data.kind)}
          maskColor="rgba(245, 241, 234, 0.72)"
        />
        <Panel position="top-right" className="flex gap-2">
          <button
            type="button"
            onClick={() => {
              applyLayout();
              window.setTimeout(() => {
                void fitView({ padding: 0.2, duration: 250 });
              }, 0);
            }}
            className="rounded-md border border-line bg-white px-3 py-2 text-xs font-semibold text-ink shadow-sm transition hover:border-moss hover:text-moss"
          >
            Reset layout
          </button>
        </Panel>
      </ReactFlow>
    </div>
  );
}

function nodeClassName(node: GraphNode, selected: boolean): string {
  const base = "!rounded-md !border !px-3 !py-2 !shadow-sm !transition-shadow";
  const tone = nodeToneClass(node);
  const selectedTone = selected ? "!border-moss !bg-green-50 !shadow-[0_0_0_3px_rgba(88,116,95,0.25)]" : "";
  return `${base} ${tone} ${selectedTone}`;
}

function nodeToneClass(node: GraphNode): string {
  if (node.is_external || node.kind === "external") {
    return "!border-signal/40 !bg-orange-50";
  }
  if (node.kind === "unresolved") {
    return "!border-signal/50 !bg-red-50";
  }
  if (node.kind === "method") {
    return "!border-moss/40 !bg-white";
  }
  return "!border-blue-300 !bg-blue-50";
}

function toFlowNode(node: PositionedNode, selected: boolean): FlowNode {
  return {
    id: node.id,
    position: node.position,
    data: {
      kind: node.kind,
      label: (
        <div className="grid gap-1">
          <div className="flex items-center gap-2">
            <span className="text-sm font-semibold text-ink">{node.label}</span>
            <span className={kindBadgeClass(node)}>{node.kind}</span>
          </div>
          <span className="font-mono text-[10px] text-steel">{node.file || node.package}</span>
        </div>
      ),
    },
    className: nodeClassName(node, selected),
  };
}

function kindBadgeClass(node: GraphNode): string {
  if (node.kind === "external") {
    return "rounded bg-orange-100 px-1.5 py-0.5 font-mono text-[9px] uppercase tracking-wide text-signal";
  }
  if (node.kind === "unresolved") {
    return "rounded bg-red-100 px-1.5 py-0.5 font-mono text-[9px] uppercase tracking-wide text-signal";
  }
  if (node.kind === "method") {
    return "rounded bg-green-100 px-1.5 py-0.5 font-mono text-[9px] uppercase tracking-wide text-moss";
  }
  return "rounded bg-blue-100 px-1.5 py-0.5 font-mono text-[9px] uppercase tracking-wide text-blue-700";
}

function toFlowEdge(edge: Graph["edges"][number], highlighted: boolean, selected: boolean): FlowEdge {
  const stroke = edgeStroke(edge.resolution);
  const active = highlighted || selected;
  return {
    id: edge.id,
    source: edge.from,
    target: edge.to,
    type: "smoothstep",
    animated: edge.resolution === "interface" || edge.resolution === "unresolved" || active,
    label: edge.candidate ? "candidate" : edge.resolution,
    markerEnd: {
      type: MarkerType.ArrowClosed,
      color: selected ? "#1f2933" : stroke,
      width: 16,
      height: 16,
    },
    style: {
      stroke: selected ? "#1f2933" : stroke,
      strokeWidth: selected ? 3.4 : active ? 3 : 1.6,
      strokeDasharray: edge.candidate || edge.resolution === "interface" ? "6 4" : undefined,
      opacity: active ? 1 : 0.78,
    },
    labelStyle: {
      fill: active ? "#1f2933" : "#52616b",
      fontSize: 11,
      fontWeight: 700,
    },
    labelBgStyle: {
      fill: "#ffffff",
      fillOpacity: 0.92,
    },
    labelBgPadding: [4, 2],
  };
}

function edgeStroke(resolution: string): string {
  switch (resolution) {
    case "resolved":
      return "#58745f";
    case "interface":
      return "#7c5fb3";
    case "external":
      return "#d97706";
    case "unresolved":
      return "#d95f3d";
    default:
      return "#52616b";
  }
}

function miniMapNodeColor(kind: unknown): string {
  switch (kind) {
    case "method":
      return "#8fb199";
    case "external":
      return "#f2b277";
    case "unresolved":
      return "#e68a73";
    default:
      return "#91b7d9";
  }
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
