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

import type { PackageEdge, PackageGraph, PackageNode } from "@/types/graph";

interface PackageGraphViewProps {
  graph: PackageGraph | null;
  selectedPackageID: string | null;
  loading?: boolean;
  error?: string | null;
  onPackageSelect: (node: PackageNode) => void;
  onPackageOpen: (node: PackageNode) => void;
}

interface PositionedPackageNode extends PackageNode {
  position: {
    x: number;
    y: number;
  };
}

export function PackageGraphView({
  graph,
  selectedPackageID,
  loading,
  error,
  onPackageSelect,
  onPackageOpen,
}: PackageGraphViewProps) {
  if (loading) {
    return <PackageGraphState message="Loading package graph" />;
  }
  if (error) {
    return <PackageGraphState message={error} tone="error" />;
  }
  if (!graph || graph.nodes.length === 0) {
    return <PackageGraphState message="No package graph loaded" />;
  }

  return (
    <ReactFlowProvider>
      <PackageGraphCanvas
        graph={graph}
        selectedPackageID={selectedPackageID}
        onPackageSelect={onPackageSelect}
        onPackageOpen={onPackageOpen}
      />
    </ReactFlowProvider>
  );
}

function PackageGraphCanvas({
  graph,
  selectedPackageID,
  onPackageSelect,
  onPackageOpen,
}: {
  graph: PackageGraph;
  selectedPackageID: string | null;
  onPackageSelect: (node: PackageNode) => void;
  onPackageOpen: (node: PackageNode) => void;
}) {
  const [nodes, setNodes, onNodesChange] = useNodesState<FlowNode>([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState<FlowEdge>([]);
  const { fitView } = useReactFlow();
  const nodeByID = useMemo(() => new Map(graph.nodes.map((node) => [node.id, node])), [graph]);

  const applyLayout = useCallback(() => {
    const positioned = layoutPackageGraph(graph);
    setNodes(positioned.map((node) => toFlowNode(node, selectedPackageID === node.id)));
    setEdges(graph.edges.map(toFlowEdge));
  }, [graph, selectedPackageID, setEdges, setNodes]);

  useEffect(() => {
    applyLayout();
    const timeout = window.setTimeout(() => {
      void fitView({ padding: 0.2, duration: 250 });
    }, 0);
    return () => window.clearTimeout(timeout);
  }, [applyLayout, fitView]);

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
        onNodeClick={(_, flowNode) => {
          const node = nodeByID.get(flowNode.id);
          if (node) {
            onPackageSelect(node);
          }
        }}
        onNodeDoubleClick={(_, flowNode) => {
          const node = nodeByID.get(flowNode.id);
          if (node) {
            onPackageOpen(node);
          }
        }}
      >
        <Background variant={BackgroundVariant.Dots} gap={18} size={1} color="#d9d3c7" />
        <Controls position="bottom-left" />
        <MiniMap pannable zoomable position="bottom-right" nodeColor="#8fb199" maskColor="rgba(245, 241, 234, 0.72)" />
        <Panel position="top-right">
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

function layoutPackageGraph(graph: PackageGraph): PositionedPackageNode[] {
  const depthByNode = computePackageDepths(graph.nodes, graph.edges);
  const buckets = new Map<number, PackageNode[]>();
  for (const node of graph.nodes) {
    const depth = depthByNode.get(node.id) ?? 0;
    const bucket = buckets.get(depth) ?? [];
    bucket.push(node);
    buckets.set(depth, bucket);
  }
  for (const bucket of buckets.values()) {
    bucket.sort(comparePackageNodes);
  }

  return [...graph.nodes].sort(comparePackageNodes).map((node) => {
    const depth = depthByNode.get(node.id) ?? 0;
    const bucket = buckets.get(depth) ?? [];
    const index = bucket.findIndex((item) => item.id === node.id);
    const centeredIndex = Math.max(index, 0) - (bucket.length - 1) / 2;
    return {
      ...node,
      position: {
        x: depth * 340 + 90,
        y: centeredIndex * 160 + 300,
      },
    };
  });
}

function computePackageDepths(nodes: PackageNode[], edges: PackageEdge[]): Map<string, number> {
  const adjacency = new Map<string, string[]>();
  const indegree = new Map(nodes.map((node) => [node.id, 0]));
  for (const edge of edges) {
    adjacency.set(edge.from, [...(adjacency.get(edge.from) ?? []), edge.to]);
    indegree.set(edge.to, (indegree.get(edge.to) ?? 0) + 1);
  }

  const roots = nodes.filter((node) => (indegree.get(node.id) ?? 0) === 0).sort(comparePackageNodes);
  const queue = roots.length > 0 ? roots.map((node) => node.id) : [...nodes].sort(comparePackageNodes).map((node) => node.id);
  const depths = new Map(queue.map((id) => [id, 0]));
  while (queue.length > 0) {
    const current = queue.shift();
    if (!current) {
      continue;
    }
    const nextDepth = (depths.get(current) ?? 0) + 1;
    for (const target of adjacency.get(current) ?? []) {
      const previous = depths.get(target);
      if (previous === undefined || nextDepth > previous) {
        depths.set(target, nextDepth);
        queue.push(target);
      }
    }
  }
  return depths;
}

function toFlowNode(node: PositionedPackageNode, selected: boolean): FlowNode {
  return {
    id: node.id,
    position: node.position,
    data: {
      label: (
        <div className="grid min-w-[180px] gap-2">
          <span className="break-words font-mono text-sm font-semibold text-ink">{node.package}</span>
          <span className="text-xs text-steel">
            {node.symbols} symbols / {node.calls} calls
          </span>
        </div>
      ),
    },
    className: selected
      ? "!rounded-md !border !border-moss !bg-green-50 !px-3 !py-2 !shadow-[0_0_0_3px_rgba(88,116,95,0.25)]"
      : "!rounded-md !border !border-line !bg-white !px-3 !py-2 !shadow-sm",
  };
}

function toFlowEdge(edge: PackageEdge): FlowEdge {
  return {
    id: edge.id,
    source: edge.from,
    target: edge.to,
    type: "smoothstep",
    label: `${edge.calls} calls`,
    markerEnd: {
      type: MarkerType.ArrowClosed,
      color: "#58745f",
      width: 16,
      height: 16,
    },
    style: {
      stroke: "#58745f",
      strokeWidth: Math.min(4, 1.4 + edge.calls * 0.3),
      opacity: 0.85,
    },
    labelStyle: {
      fill: "#1f2933",
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

function comparePackageNodes(a: PackageNode, b: PackageNode): number {
  return a.id.localeCompare(b.id);
}

function PackageGraphState({ message, tone = "muted" }: { message: string; tone?: "muted" | "error" }) {
  return (
    <div className="grid h-full place-items-center bg-paper">
      <p className={tone === "error" ? "max-w-md text-sm text-signal" : "text-sm text-steel"}>{message}</p>
    </div>
  );
}
