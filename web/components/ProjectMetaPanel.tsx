"use client";

import type { ProjectMeta } from "@/types/graph";

interface ProjectMetaPanelProps {
  meta: ProjectMeta | null;
  loading?: boolean;
  error?: string | null;
  onRescan: () => void;
}

export function ProjectMetaPanel({ meta, loading, error, onRescan }: ProjectMetaPanelProps) {
  return (
    <section className="grid min-w-0 gap-3 rounded-md border border-line bg-white p-3">
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <h2 className="text-sm font-semibold text-ink">Project</h2>
          <p className="mt-1 break-all font-mono text-xs leading-5 text-steel" title={meta?.module || meta?.root || ""}>
            {meta?.module || meta?.root || "No metadata loaded"}
          </p>
        </div>
        <button
          type="button"
          onClick={onRescan}
          disabled={loading}
          className="h-8 shrink-0 rounded-md border border-line bg-white px-3 text-xs font-semibold text-ink transition hover:border-moss hover:text-moss disabled:cursor-not-allowed disabled:text-steel"
        >
          {loading ? "Rescanning" : "Rescan"}
        </button>
      </div>

      {meta ? (
        <dl className="grid min-w-0 grid-cols-2 gap-2 text-xs min-[1280px]:grid-cols-4">
          <MetaItem label="Packages" value={meta.packages} />
          <MetaItem label="Symbols" value={meta.symbols} />
          <MetaItem label="Calls" value={meta.calls} />
          <MetaItem label="Warnings" value={meta.warnings} />
        </dl>
      ) : null}

      {meta ? (
        <p className="break-all font-mono text-[11px] leading-5 text-steel">
          {meta.analysis_duration_ms}ms · {formatAnalyzedAt(meta.analyzed_at)} · {meta.version}
        </p>
      ) : null}
      {error ? <p className="text-sm text-signal">{error}</p> : null}
    </section>
  );
}

function MetaItem({ label, value }: { label: string; value: number }) {
  return (
    <div className="min-w-0 rounded border border-line bg-paper px-2 py-1">
      <dt className="text-[10px] uppercase text-steel">{label}</dt>
      <dd className="mt-0.5 font-mono text-sm font-semibold text-ink">{value}</dd>
    </div>
  );
}

function formatAnalyzedAt(value: string): string {
  if (!value) {
    return "not analyzed";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return date.toLocaleString();
}
