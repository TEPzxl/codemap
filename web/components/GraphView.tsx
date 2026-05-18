"use client";

import { useCallback, useEffect, useMemo } from "react";
import {
  Background,
  BackgroundVariant,
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

import { ResetLayoutButton } from "@/components/ResetLayoutButton";
import { layoutGraph } from "@/lib/layoutGraph";
import type { PositionedNode } from "@/lib/layoutGraph";
import type { Edge as GraphEdge, Graph, Node as GraphNode } from "@/types/graph";

interface GraphViewProps {
  graph: Graph | null;
  selectedNode: GraphNode | null;
  selectedEdgeID: string | null;
  loading?: boolean;
  error?: string | null;
  onNodeSelect: (node: GraphNode) => void;
  onEdgeSelect: (edge: GraphEdge) => void;
}

export function GraphView({
  graph,
  selectedNode,
  selectedEdgeID,
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
  onNodeSelect,
  onEdgeSelect,
}: {
  graph: Graph;
  selectedNode: GraphNode | null;
  selectedEdgeID: string | null;
  onNodeSelect: (node: GraphNode) => void;
  onEdgeSelect: (edge: GraphEdge) => void;
}) {
  const [nodes, setNodes, onNodesChange] = useNodesState<FlowNode>([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState<FlowEdge>([]);
  const { fitView, setCenter } = useReactFlow();

  const nodeByID = useMemo(() => new Map(graph?.nodes.map((node) => [node.id, node]) ?? []), [graph]);
  const edgeByID = useMemo(() => new Map(graph.edges.map((edge) => [edge.id, edge])), [graph]);
  const neighborhood = useMemo(() => {
    const empty = {
      upstream: new Set<string>(),
      downstream: new Set<string>(),
      relatedNodes: new Set<string>(),
      relatedEdges: new Set<string>(),
    };
    if (!graph || !selectedNode) {
      return empty;
    }
    const upstream = new Set<string>();
    const downstream = new Set<string>();
    const relatedNodes = new Set<string>([selectedNode.id]);
    const relatedEdges = new Set<string>();
    for (const edge of graph.edges) {
      if (edge.to === selectedNode.id) {
        upstream.add(edge.from);
        relatedNodes.add(edge.from);
        relatedEdges.add(edge.id);
      }
      if (edge.from === selectedNode.id) {
        downstream.add(edge.to);
        relatedNodes.add(edge.to);
        relatedEdges.add(edge.id);
      }
    }
    return { upstream, downstream, relatedNodes, relatedEdges };
  }, [graph, selectedNode]);

  const applyLayout = useCallback(() => {
    if (!graph) {
      setNodes([]);
      setEdges([]);
      return;
    }
    const positioned = layoutGraph(graph);
    setNodes(positioned.map((node) => toFlowNode(node, nodeSelectionState(node, selectedNode, neighborhood))));
    setEdges(graph.edges.map((edge) => toFlowEdge(edge, neighborhood.relatedEdges.has(edge.id), selectedEdgeID === edge.id, Boolean(selectedNode))));
  }, [graph, neighborhood, selectedEdgeID, selectedNode, setEdges, setNodes]);

  useEffect(() => {
    applyLayout();
  }, [applyLayout]);

  useEffect(() => {
    const timeout = window.setTimeout(() => {
      void fitView({ padding: 0.2, duration: 250 });
    }, 0);
    return () => window.clearTimeout(timeout);
  }, [fitView, graph]);

  useEffect(() => {
    if (!selectedNode) {
      return;
    }
    const positioned = layoutGraph(graph).find((node) => node.id === selectedNode.id);
    if (!positioned) {
      return;
    }
    const timeout = window.setTimeout(() => {
      void setCenter(positioned.position.x + 150, positioned.position.y + 45, { zoom: 1.05, duration: 260 });
    }, 0);
    return () => window.clearTimeout(timeout);
  }, [graph, selectedNode, setCenter]);

  return (
    <div className="relative h-full min-h-[520px]">
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
        <MiniMap
          pannable
          zoomable
          position="bottom-right"
          nodeColor={(node) => miniMapNodeColor(node.data.kind)}
          maskColor="rgba(245, 241, 234, 0.72)"
        />
        <Panel position="top-right">
          <ResetLayoutButton
            onClick={() => {
              applyLayout();
              window.setTimeout(() => {
                void fitView({ padding: 0.2, duration: 250 });
              }, 0);
            }}
          />
        </Panel>
      </ReactFlow>
    </div>
  );
}

type NodeSelectionState = "selected" | "upstream" | "downstream" | "dimmed" | "default";

function nodeSelectionState(node: GraphNode, selectedNode: GraphNode | null, neighborhood: { upstream: Set<string>; downstream: Set<string>; relatedNodes: Set<string> }): NodeSelectionState {
  if (!selectedNode) {
    return "default";
  }
  if (node.id === selectedNode.id) {
    return "selected";
  }
  if (neighborhood.upstream.has(node.id)) {
    return "upstream";
  }
  if (neighborhood.downstream.has(node.id)) {
    return "downstream";
  }
  if (!neighborhood.relatedNodes.has(node.id)) {
    return "dimmed";
  }
  return "default";
}

function nodeClassName(node: GraphNode, selection: NodeSelectionState): string {
  const base = "!rounded-md !border !px-3 !py-2 !shadow-sm !transition-all";
  const tone = nodeToneClass(node);
  const selectionTone = nodeSelectionClass(selection);
  return `${base} ${tone} ${selectionTone}`;
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

function nodeSelectionClass(selection: NodeSelectionState): string {
  switch (selection) {
    case "selected":
      return "!border-ink !bg-green-50 !opacity-100 !shadow-[0_0_0_3px_rgba(88,116,95,0.28)]";
    case "upstream":
      return "!border-moss !bg-white !opacity-100 !shadow-[0_0_0_2px_rgba(88,116,95,0.16)]";
    case "downstream":
      return "!border-blue-500 !bg-white !opacity-100 !shadow-[0_0_0_2px_rgba(80,126,164,0.16)]";
    case "dimmed":
      return "!opacity-35";
    default:
      return "!opacity-100";
  }
}

function toFlowNode(node: PositionedNode, selection: NodeSelectionState): FlowNode {
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
    className: nodeClassName(node, selection),
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

function toFlowEdge(edge: Graph["edges"][number], highlighted: boolean, selected: boolean, hasSelectedNode: boolean): FlowEdge {
  const stroke = edgeStroke(edge.resolution);
  const active = highlighted || selected;
  const dimmed = hasSelectedNode && !active;
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
      opacity: dimmed ? 0.22 : active ? 1 : 0.78,
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

function GraphState({ message, tone = "muted" }: { message: string; tone?: "muted" | "error" }) {
  return (
    <div className="grid h-full place-items-center bg-paper">
      <p className={tone === "error" ? "max-w-md text-sm text-signal" : "text-sm text-steel"}>{message}</p>
    </div>
  );
}
