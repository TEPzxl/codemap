"use client";

import { useMemo, useState } from "react";

import { displayPackage, displaySymbolID } from "@/lib/displaySymbol";
import { searchEntrypoints } from "@/lib/entrypointSearch";
import type { Entrypoint } from "@/types/graph";

interface EntrypointsPanelProps {
  entrypoints: Entrypoint[];
  note: string;
  modulePrefix: string;
  loading?: boolean;
  disabled?: boolean;
  onSelectEntrypoint: (id: string) => void;
}

export function EntrypointsPanel({
  entrypoints,
  note,
  modulePrefix,
  loading,
  disabled,
  onSelectEntrypoint,
}: EntrypointsPanelProps) {
  const [query, setQuery] = useState("");
  const normalizedQuery = query.trim();
  const groups = useMemo(() => searchEntrypoints(entrypoints, query, normalizedQuery ? 10 : 6), [entrypoints, normalizedQuery, query]);
  const visibleCount = groups.reduce((count, group) => count + group.items.length, 0);

  return (
    <section className="grid min-w-0 gap-3 rounded-md border border-line bg-white p-3">
      <div>
        <h2 className="text-sm font-semibold text-ink">Entrypoints</h2>
        <p className="mt-1 text-xs leading-5 text-steel">{note || "Heuristic entrypoint candidates."}</p>
      </div>

      {loading ? <p className="text-xs text-steel">Loading entrypoints</p> : null}
      {!loading && entrypoints.length === 0 ? <p className="text-xs text-steel">No entrypoints found</p> : null}

      <label className="grid gap-2 text-sm">
        <span className="font-semibold text-ink">Search entrypoints</span>
        <input
          type="search"
          value={query}
          onChange={(event) => setQuery(event.target.value)}
          placeholder="Search label, package, file, reason"
          disabled={loading}
          className="h-10 min-w-0 rounded-md border border-line bg-white px-3 text-sm text-ink outline-none transition placeholder:text-steel focus:border-moss focus:ring-2 focus:ring-moss/20 disabled:cursor-not-allowed disabled:bg-paper"
        />
      </label>

      {!loading && entrypoints.length > 0 ? (
        <p className="text-xs text-steel">
          Showing {visibleCount} of {entrypoints.length} entrypoints
        </p>
      ) : null}
      {!loading && entrypoints.length > 0 && visibleCount === 0 ? <p className="text-xs text-steel">No matching entrypoints</p> : null}

      <div className="grid gap-3">
        {groups.map((group) => (
          <div key={group.label} className="grid gap-2">
            <h3 className="text-xs font-semibold uppercase text-steel">{group.label}</h3>
            <div className="grid gap-2">
              {group.items.map((entrypoint) => (
                <button
                  key={entrypoint.id}
                  type="button"
                  disabled={disabled}
                  onClick={() => {
                    setQuery(displaySymbolID(entrypoint.id, modulePrefix));
                    onSelectEntrypoint(entrypoint.id);
                  }}
                  className="grid min-w-0 gap-1 rounded-md border border-line bg-paper px-3 py-2 text-left transition hover:border-moss hover:bg-green-50 disabled:cursor-not-allowed disabled:opacity-60"
                >
                  <span className="grid min-w-0 grid-cols-[minmax(0,1fr)_auto] items-center gap-2">
                    <span className="truncate text-sm font-semibold text-ink">{entrypoint.label}</span>
                    <span className="truncate text-[11px] text-steel">{displayPackage(entrypoint.package, modulePrefix)}</span>
                  </span>
                  <span className="truncate font-mono text-[11px] text-steel" title={entrypoint.id}>
                    {displaySymbolID(entrypoint.id, modulePrefix)}
                  </span>
                  <span className="flex flex-wrap gap-1">
                    {entrypoint.reasons.slice(0, 3).map((reason) => (
                      <span key={reason} className="rounded border border-line bg-white px-1.5 py-0.5 font-mono text-[10px] text-steel">
                        {reason}
                      </span>
                    ))}
                  </span>
                </button>
              ))}
            </div>
          </div>
        ))}
      </div>
    </section>
  );
}
