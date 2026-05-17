"use client";

import type { SourceSnippet } from "@/types/graph";

interface SourcePanelProps {
  source: SourceSnippet | null;
  loading?: boolean;
  error?: string | null;
}

export function SourcePanel({ source, loading, error }: SourcePanelProps) {
  return (
    <section className="min-h-48 border-t border-line bg-[#111820] text-stone-100">
      <div className="flex items-center justify-between border-b border-white/10 px-4 py-3">
        <div>
          <h2 className="text-sm font-semibold">Source</h2>
          <p className="mt-1 font-mono text-xs text-stone-400">
            {source ? `${source.file}:${source.start_line}-${source.end_line}` : "No node selected"}
          </p>
        </div>
        {source ? (
          <span className="rounded border border-white/10 px-2 py-1 font-mono text-xs text-stone-300">
            {source.language}
          </span>
        ) : null}
      </div>

      <div className="max-h-72 overflow-auto p-4">
        {loading ? <p className="text-sm text-stone-300">Loading source</p> : null}
        {error ? <p className="text-sm text-red-300">{error}</p> : null}
        {!loading && !error && !source ? (
          <p className="text-sm text-stone-400">Select a graph node to inspect its source snippet.</p>
        ) : null}
        {source && !loading && !error ? (
          <pre className="whitespace-pre-wrap font-mono text-sm leading-6">
            <code>{source.source}</code>
          </pre>
        ) : null}
      </div>
    </section>
  );
}
