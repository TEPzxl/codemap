"use client";

import type { GraphRequest } from "@/lib/api";
import type { GraphSummary } from "@/lib/graphSummary";
import type { GraphDirection, Node as GraphNode } from "@/types/graph";

interface GraphSummaryPanelProps {
  entry: string;
  request: GraphRequest;
  summary: GraphSummary;
  loading?: boolean;
  copyStatus: string | null;
  selectedNode: GraphNode | null;
  onCopyViewURL: () => void;
  onFocusNode: (node: GraphNode, direction: GraphDirection) => void;
  onResetToEntry: () => void;
}

export function GraphSummaryPanel({
  entry,
  request,
  summary,
  loading,
  copyStatus,
  selectedNode,
  onCopyViewURL,
  onFocusNode,
  onResetToEntry,
}: GraphSummaryPanelProps) {
  const activeFilters = activeFilterLabels(request);
  const direction = request.direction ?? "downstream";

  return (
    <section className="grid min-w-0 gap-3 rounded-md border border-line bg-white p-3">
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <h2 className="text-sm font-semibold text-ink">Current graph</h2>
          <p className="mt-1 break-all font-mono text-xs leading-5 text-steel" title={entry}>
            {entry || "No entry loaded"}
          </p>
        </div>
        <button
          type="button"
          onClick={onCopyViewURL}
          className="h-8 shrink-0 rounded-md border border-line bg-white px-3 text-xs font-semibold text-ink transition hover:border-moss hover:text-moss"
        >
          Copy view URL
        </button>
      </div>

      <dl className="grid min-w-0 grid-cols-2 gap-2 text-xs min-[1280px]:grid-cols-4">
        <SummaryItem label="Depth" value={request.depth} />
        <SummaryItem label="Nodes" value={summary.nodes} />
        <SummaryItem label="Edges" value={summary.edges} />
        <SummaryItem label="Packages" value={summary.packages} />
        <SummaryItem label="Resolved" value={summary.resolvedEdges} />
        <SummaryItem label="Interface" value={summary.interfaceEdges} />
        <SummaryItem label="External" value={summary.externalEdges} />
        <SummaryItem label="Unresolved" value={summary.unresolvedEdges} />
      </dl>

      <div className="grid gap-1 text-xs">
        <p className="font-semibold text-ink">Focus</p>
        <p className="font-mono text-steel">{direction}</p>
      </div>

      <div className="grid gap-1 text-xs">
        <p className="font-semibold text-ink">Filters</p>
        <p className="break-words font-mono leading-5 text-steel">{activeFilters.length > 0 ? activeFilters.join(", ") : "default"}</p>
      </div>

      {selectedNode ? (
        <div className="grid gap-2 border-t border-line pt-3">
          <p className="break-all text-xs text-steel" title={selectedNode.id}>
            Focus selected: <span className="font-semibold text-ink">{selectedNode.label}</span>
          </p>
          <div className="grid grid-cols-2 gap-2">
            <FocusButton label="Focus downstream" onClick={() => onFocusNode(selectedNode, "downstream")} />
            <FocusButton label="Focus upstream" onClick={() => onFocusNode(selectedNode, "upstream")} />
            <FocusButton label="Focus neighborhood" onClick={() => onFocusNode(selectedNode, "both")} />
            <FocusButton label="Reset to entry" onClick={onResetToEntry} />
          </div>
        </div>
      ) : null}

      {copyStatus ? <p className="text-xs text-moss">{copyStatus}</p> : null}
      {loading ? <p className="text-xs text-steel">Refreshing graph</p> : null}
    </section>
  );
}

function FocusButton({ label, onClick }: { label: string; onClick: () => void }) {
  return (
    <button
      type="button"
      onClick={onClick}
      className="h-8 rounded-md border border-line bg-white px-2 text-xs font-semibold text-ink transition hover:border-moss hover:text-moss"
    >
      {label}
    </button>
  );
}

function SummaryItem({ label, value }: { label: string; value: number }) {
  return (
    <div className="min-w-0 rounded border border-line bg-paper px-2 py-1">
      <dt className="text-[10px] uppercase text-steel">{label}</dt>
      <dd className="mt-0.5 font-mono text-sm font-semibold text-ink">{value}</dd>
    </div>
  );
}

function activeFilterLabels(request: GraphRequest): string[] {
  const labels: string[] = [];
  if (request.showExternal) {
    labels.push("show_external");
  }
  if (request.showUnresolved) {
    labels.push("show_unresolved");
  }
  if (request.showInterface) {
    labels.push("show_interface");
  }
  if (request.expandInterface) {
    labels.push("expand_interface");
  }
  if (request.packagePrefix) {
    labels.push(`package=${request.packagePrefix}`);
  }
  return labels;
}
