"use client";

import { displayPackage, displaySymbolID } from "@/lib/displaySymbol";
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
  const groups = groupEntrypoints(entrypoints);

  return (
    <section className="grid min-w-0 gap-3 rounded-md border border-line bg-white p-3">
      <div>
        <h2 className="text-sm font-semibold text-ink">Entrypoints</h2>
        <p className="mt-1 text-xs leading-5 text-steel">{note || "Heuristic entrypoint candidates."}</p>
      </div>

      {loading ? <p className="text-xs text-steel">Loading entrypoints</p> : null}
      {!loading && entrypoints.length === 0 ? <p className="text-xs text-steel">No entrypoints found</p> : null}

      <div className="grid max-h-80 gap-3 overflow-y-auto pr-1">
        {groups.map((group) => (
          <div key={group.label} className="grid gap-2">
            <h3 className="text-xs font-semibold uppercase text-steel">{group.label}</h3>
            <div className="grid gap-2">
              {group.items.map((entrypoint) => (
                <button
                  key={entrypoint.id}
                  type="button"
                  disabled={disabled}
                  onClick={() => onSelectEntrypoint(entrypoint.id)}
                  className="grid min-h-16 min-w-0 gap-1 rounded-md border border-line bg-paper px-3 py-2 text-left transition hover:border-moss hover:bg-green-50 disabled:cursor-not-allowed disabled:opacity-60"
                >
                  <span className="truncate text-sm font-semibold text-ink">{entrypoint.label}</span>
                  <span className="truncate font-mono text-[11px] text-steel" title={entrypoint.id}>
                    {displaySymbolID(entrypoint.id, modulePrefix)}
                  </span>
                  <span className="truncate text-[11px] text-steel">{displayPackage(entrypoint.package, modulePrefix)}</span>
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

interface EntrypointGroup {
  label: string;
  items: Entrypoint[];
}

function groupEntrypoints(entrypoints: Entrypoint[]): EntrypointGroup[] {
  const order = ["Main", "Startup", "Handlers", "Goroutines", "Exported"];
  const groups = new Map(order.map((label) => [label, [] as Entrypoint[]]));

  for (const entrypoint of entrypoints) {
    const label = groupLabel(entrypoint);
    groups.get(label)?.push(entrypoint);
  }

  return order
    .map((label) => ({
      label,
      items: groups.get(label) ?? [],
    }))
    .filter((group) => group.items.length > 0);
}

function groupLabel(entrypoint: Entrypoint): string {
  if (entrypoint.reasons.includes("main-function")) {
    return "Main";
  }
  if (entrypoint.reasons.some((reason) => reason === "receiver:Handler" || reason === "name:Handler" || reason === "name:Handle")) {
    return "Handlers";
  }
  if (entrypoint.reasons.some((reason) => reason.startsWith("name:"))) {
    return "Startup";
  }
  if (entrypoint.reasons.includes("contains-goroutine")) {
    return "Goroutines";
  }
  return "Exported";
}
