"use client";

import { displayPackage, displaySymbolID } from "@/lib/displaySymbol";
import type { Node as GraphNode } from "@/types/graph";

interface CurrentGraphSearchPanelProps {
  query: string;
  results: GraphNode[];
  selectedNodeID?: string;
  modulePrefix: string;
  disabled?: boolean;
  onQueryChange: (query: string) => void;
  onSelectNode: (node: GraphNode) => void;
}

export function CurrentGraphSearchPanel({
  query,
  results,
  selectedNodeID,
  modulePrefix,
  disabled,
  onQueryChange,
  onSelectNode,
}: CurrentGraphSearchPanelProps) {
  return (
    <section className="grid min-w-0 gap-3 rounded-md border border-line bg-white p-3">
      <div>
        <h2 className="text-sm font-semibold text-ink">Search current graph</h2>
        <p className="mt-1 text-xs text-steel">Matches label, id, package, and file.</p>
      </div>

      <input
        type="search"
        value={query}
        onChange={(event) => onQueryChange(event.target.value)}
        disabled={disabled}
        placeholder="Search nodes"
        className="h-10 min-w-0 rounded-md border border-line bg-white px-3 text-sm text-ink outline-none transition placeholder:text-steel focus:border-moss disabled:cursor-not-allowed disabled:bg-paper"
      />

      {query.trim() && results.length === 0 ? <p className="text-xs text-steel">No nodes found</p> : null}

      {results.length > 0 ? (
        <div className="grid max-h-72 gap-2 overflow-y-auto pr-1">
          {results.map((node) => (
            <button
              key={node.id}
              type="button"
              onClick={() => onSelectNode(node)}
              disabled={disabled}
              className={
                selectedNodeID === node.id
                  ? "grid min-h-16 min-w-0 gap-1 rounded-md border border-moss bg-green-50 px-3 py-2 text-left"
                  : "grid min-h-16 min-w-0 gap-1 rounded-md border border-line bg-paper px-3 py-2 text-left transition hover:border-moss hover:bg-green-50 disabled:cursor-not-allowed disabled:opacity-60"
              }
            >
              <span className="truncate text-sm font-semibold text-ink">{node.label}</span>
              <span className="truncate font-mono text-[11px] text-steel" title={node.id}>
                {displaySymbolID(node.id, modulePrefix)}
              </span>
              <span className="truncate text-[11px] text-steel">
                {node.file || displayPackage(node.package, modulePrefix)}
              </span>
            </button>
          ))}
        </div>
      ) : null}
    </section>
  );
}
