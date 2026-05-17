"use client";

interface ToolbarProps {
  depth: number;
  onDepthChange: (depth: number) => void;
  onLoadGraph: () => void;
  loading?: boolean;
}

export function Toolbar({ depth, onDepthChange, onLoadGraph, loading }: ToolbarProps) {
  return (
    <div className="grid gap-4">
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
          className="accent-moss"
        />
      </label>

      <button
        type="button"
        onClick={onLoadGraph}
        disabled={loading}
        className="h-10 rounded-md bg-ink px-4 text-sm font-semibold text-white transition hover:bg-moss disabled:cursor-not-allowed disabled:bg-steel"
      >
        {loading ? "Loading graph" : "Load graph"}
      </button>
    </div>
  );
}
