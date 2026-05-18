"use client";

import type { SourceView } from "@/types/graph";

interface SourcePanelProps {
  source: SourceView | null;
  loading?: boolean;
  error?: string | null;
}

export function SourcePanel({ source, loading, error }: SourcePanelProps) {
  const data = source?.data ?? null;
  const title = source?.mode === "callsite" ? "Callsite" : "Node source";
  const location = source ? sourceLocation(source) : "No graph item selected";

  return (
    <section className="min-h-48 border-t border-line bg-[#111820] text-stone-100">
      <div className="flex items-center justify-between border-b border-white/10 px-4 py-3">
        <div>
          <h2 className="text-sm font-semibold">{title}</h2>
          <p className="mt-1 font-mono text-xs text-stone-400">{location}</p>
        </div>
        {data ? (
          <div className="flex items-center gap-2">
            {source?.mode === "callsite" ? (
              <span className="rounded border border-white/10 px-2 py-1 font-mono text-xs text-stone-300">
                {source.data.line}:{source.data.column}
              </span>
            ) : null}
            <span className="rounded border border-white/10 px-2 py-1 font-mono text-xs text-stone-300">
              {data.language}
            </span>
          </div>
        ) : null}
      </div>

      <div className="max-h-72 overflow-auto p-4">
        {loading ? <p className="text-sm text-stone-300">Loading source</p> : null}
        {error ? <p className="text-sm text-red-300">{error}</p> : null}
        {!loading && !error && !source ? (
          <p className="text-sm text-stone-400">Select a graph node or edge to inspect its source snippet.</p>
        ) : null}
        {source && !loading && !error ? (
          <LineNumberedSource source={source} />
        ) : null}
      </div>
    </section>
  );
}

function sourceLocation(source: SourceView): string {
  if (source.mode === "callsite") {
    return `${source.data.file}:${source.data.line}:${source.data.column}`;
  }
  return `${source.data.file}:${source.data.start_line}-${source.data.end_line}`;
}

function LineNumberedSource({ source }: { source: SourceView }) {
  const lines = source.data.source.split("\n");
  return (
    <pre className="min-w-max font-mono text-sm leading-6">
      <code>
        {lines.map((line, index) => {
          const lineNumber = source.data.start_line + index;
          const highlighted = source.mode === "callsite" && lineNumber === source.data.highlight_line;
          return (
            <span
              key={lineNumber}
              className={`grid grid-cols-[3.5rem_minmax(0,1fr)] px-2 ${highlighted ? "bg-[#26323b] text-white" : ""}`}
            >
              <span className="select-none pr-4 text-right text-stone-500">{lineNumber}</span>
              <span className="whitespace-pre">{line || " "}</span>
            </span>
          );
        })}
      </code>
    </pre>
  );
}
