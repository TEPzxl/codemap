"use client";

import type { PathResult, SymbolInfo } from "@/types/graph";
import { displaySymbolID } from "@/lib/displaySymbol";

interface PathSearchPanelProps {
  symbols: SymbolInfo[];
  from: string;
  to: string;
  maxDepth: number;
  limit: number;
  loading?: boolean;
  error?: string | null;
  result: PathResult | null;
  selectedPathIndex: number;
  modulePrefix: string;
  disabled?: boolean;
  onFromChange: (value: string) => void;
  onToChange: (value: string) => void;
  onMaxDepthChange: (value: number) => void;
  onLimitChange: (value: number) => void;
  onFindPath: () => void;
  onSelectPath: (index: number) => void;
}

export function PathSearchPanel({
  symbols,
  from,
  to,
  maxDepth,
  limit,
  loading,
  error,
  result,
  selectedPathIndex,
  modulePrefix,
  disabled,
  onFromChange,
  onToChange,
  onMaxDepthChange,
  onLimitChange,
  onFindPath,
  onSelectPath,
}: PathSearchPanelProps) {
  return (
    <section className="grid min-w-0 gap-3 rounded-md border border-line bg-white p-3">
      <div>
        <h2 className="text-sm font-semibold text-ink">Path search</h2>
        <p className="mt-1 text-xs text-steel">Find calls from one symbol to another.</p>
      </div>

      <datalist id="path-symbols">
        {symbols.map((symbol) => (
          <option key={symbol.id} value={symbol.id}>
            {displaySymbolID(symbol.id, modulePrefix)}
          </option>
        ))}
      </datalist>

      <PathInput label="From" value={from} onChange={onFromChange} disabled={disabled} />
      <PathInput label="To" value={to} onChange={onToChange} disabled={disabled} />

      <div className="grid grid-cols-2 gap-2">
        <NumberInput label="Max depth" value={maxDepth} min={0} max={32} onChange={onMaxDepthChange} disabled={disabled} />
        <NumberInput label="Limit" value={limit} min={1} max={20} onChange={onLimitChange} disabled={disabled} />
      </div>

      <button
        type="button"
        onClick={onFindPath}
        disabled={disabled || loading || !from.trim() || !to.trim()}
        className="h-9 rounded-md bg-ink px-3 text-sm font-semibold text-white transition hover:bg-moss disabled:cursor-not-allowed disabled:bg-steel"
      >
        {loading ? "Finding path" : "Find path"}
      </button>

      {error ? <p className="rounded border border-signal/30 bg-orange-50 p-2 text-sm text-signal">{error}</p> : null}
      {result ? <PathResultView result={result} selectedPathIndex={selectedPathIndex} modulePrefix={modulePrefix} onSelectPath={onSelectPath} /> : null}
    </section>
  );
}

function PathInput({
  label,
  value,
  onChange,
  disabled,
}: {
  label: string;
  value: string;
  onChange: (value: string) => void;
  disabled?: boolean;
}) {
  return (
    <label className="grid gap-1 text-sm">
      <span className="font-semibold text-ink">{label}</span>
      <input
        list="path-symbols"
        className="h-9 w-full min-w-0 rounded-md border border-line bg-white px-2 font-mono text-xs text-ink outline-none transition focus:border-moss focus:ring-2 focus:ring-moss/20 disabled:bg-stone-100"
        value={value}
        onChange={(event) => onChange(event.target.value)}
        placeholder="symbol id or shortcut"
        disabled={disabled}
      />
    </label>
  );
}

function NumberInput({
  label,
  value,
  min,
  max,
  onChange,
  disabled,
}: {
  label: string;
  value: number;
  min: number;
  max: number;
  onChange: (value: number) => void;
  disabled?: boolean;
}) {
  return (
    <label className="grid gap-1 text-sm">
      <span className="font-semibold text-ink">{label}</span>
      <input
        type="number"
        min={min}
        max={max}
        className="h-9 w-full rounded-md border border-line bg-white px-2 font-mono text-sm text-ink outline-none transition focus:border-moss focus:ring-2 focus:ring-moss/20 disabled:bg-stone-100"
        value={value}
        onChange={(event) => onChange(Number(event.target.value))}
        disabled={disabled}
      />
    </label>
  );
}

function PathResultView({
  result,
  selectedPathIndex,
  modulePrefix,
  onSelectPath,
}: {
  result: PathResult;
  selectedPathIndex: number;
  modulePrefix: string;
  onSelectPath: (index: number) => void;
}) {
  if (result.paths.length === 0) {
    return <p className="rounded border border-line bg-paper p-2 text-sm text-steel">No path found.</p>;
  }

  return (
    <div className="grid gap-2">
      <p className="text-xs text-steel">
        {result.paths.length} path{result.paths.length === 1 ? "" : "s"} found. The graph is showing the returned path graph.
      </p>
      <ul className="max-h-40 overflow-auto rounded-md border border-line bg-paper">
        {result.paths.map((path, index) => (
          <li key={`${path.nodes.join("->")}-${index}`} className="border-b border-line last:border-b-0">
            <button
              type="button"
              onClick={() => onSelectPath(index)}
              className={`grid w-full gap-1 px-2 py-2 text-left text-xs transition hover:bg-white ${
                selectedPathIndex === index ? "bg-green-50 text-ink" : "text-steel"
              }`}
            >
              <span className="font-semibold text-ink">Path {index + 1}</span>
              <span className="break-all font-mono leading-5">{path.nodes.map((node) => displaySymbolID(node, modulePrefix)).join(" -> ")}</span>
            </button>
          </li>
        ))}
      </ul>
    </div>
  );
}
