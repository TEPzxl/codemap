"use client";

interface ToolbarProps {
  depth: number;
  onDepthChange: (depth: number) => void;
  onLoadGraph: () => void;
  showExternal: boolean;
  showUnresolved: boolean;
  showInterface: boolean;
  expandInterface: boolean;
  onShowExternalChange: (value: boolean) => void;
  onShowUnresolvedChange: (value: boolean) => void;
  onShowInterfaceChange: (value: boolean) => void;
  onExpandInterfaceChange: (value: boolean) => void;
  loading?: boolean;
}

export function Toolbar({
  depth,
  onDepthChange,
  onLoadGraph,
  showExternal,
  showUnresolved,
  showInterface,
  expandInterface,
  onShowExternalChange,
  onShowUnresolvedChange,
  onShowInterfaceChange,
  onExpandInterfaceChange,
  loading,
}: ToolbarProps) {
  return (
    <div className="grid min-w-0 gap-4">
      <label className="grid gap-2 text-sm">
        <span className="flex items-center justify-between font-semibold text-ink">
          Depth
          <span className="rounded border border-line bg-white px-2 py-0.5 font-mono text-xs text-steel">{depth}</span>
        </span>
        <input
          type="range"
          min={0}
          max={8}
          value={depth}
          onChange={(event) => onDepthChange(Number(event.target.value))}
          className="w-full accent-moss"
        />
      </label>

      <section className="grid gap-2">
        <h2 className="text-sm font-semibold text-ink">Graph filters</h2>
        <FilterToggle label="Show external calls" checked={showExternal} onChange={onShowExternalChange} />
        <FilterToggle label="Show unresolved calls" checked={showUnresolved} onChange={onShowUnresolvedChange} />
        <FilterToggle label="Show interface calls" checked={showInterface} onChange={onShowInterfaceChange} />
        <FilterToggle label="Expand interface candidates" checked={expandInterface} onChange={onExpandInterfaceChange} />
      </section>

      <button
        type="button"
        onClick={onLoadGraph}
        disabled={loading}
        className="h-10 w-full rounded-md bg-ink px-4 text-sm font-semibold text-white transition hover:bg-moss disabled:cursor-not-allowed disabled:bg-steel"
      >
        {loading ? "Loading graph" : "Load graph"}
      </button>
    </div>
  );
}

function FilterToggle({
  label,
  checked,
  onChange,
}: {
  label: string;
  checked: boolean;
  onChange: (value: boolean) => void;
}) {
  return (
    <label className="flex min-h-8 items-center gap-2 text-sm text-ink">
      <input
        type="checkbox"
        checked={checked}
        onChange={(event) => onChange(event.target.checked)}
        className="h-4 w-4 accent-moss"
      />
      <span>{label}</span>
    </label>
  );
}
