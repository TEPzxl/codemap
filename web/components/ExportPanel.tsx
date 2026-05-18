"use client";

import type { GraphExportFormat } from "@/types/graph";

interface ExportPanelProps {
  loading?: boolean;
  disabled?: boolean;
  status: string | null;
  error: string | null;
  onCopyExport: (format: GraphExportFormat) => void;
  onDownloadExport: (format: GraphExportFormat) => void;
  onCopyViewURL: () => void;
}

const formats: Array<{ format: GraphExportFormat; label: string }> = [
  { format: "json", label: "JSON" },
  { format: "mermaid", label: "Mermaid" },
  { format: "dot", label: "DOT" },
];

export function ExportPanel({
  loading,
  disabled,
  status,
  error,
  onCopyExport,
  onDownloadExport,
  onCopyViewURL,
}: ExportPanelProps) {
  const blocked = disabled || loading;

  return (
    <section className="grid min-w-0 gap-3 rounded-md border border-line bg-white p-3">
      <div className="flex items-center justify-between gap-3">
        <h2 className="text-sm font-semibold text-ink">Export</h2>
        <button
          type="button"
          onClick={onCopyViewURL}
          disabled={blocked}
          className="h-8 shrink-0 rounded-md border border-line bg-white px-3 text-xs font-semibold text-ink transition hover:border-moss hover:text-moss disabled:cursor-not-allowed disabled:opacity-60"
        >
          Copy view URL
        </button>
      </div>

      <div className="grid gap-2">
        {formats.map(({ format, label }) => (
          <div key={format} className="grid grid-cols-2 gap-2">
            <button
              type="button"
              onClick={() => onCopyExport(format)}
              disabled={blocked}
              className="h-8 rounded-md border border-line bg-white px-2 text-xs font-semibold text-ink transition hover:border-moss hover:text-moss disabled:cursor-not-allowed disabled:opacity-60"
            >
              Copy {label}
            </button>
            <button
              type="button"
              onClick={() => onDownloadExport(format)}
              disabled={blocked}
              className="h-8 rounded-md border border-line bg-white px-2 text-xs font-semibold text-ink transition hover:border-moss hover:text-moss disabled:cursor-not-allowed disabled:opacity-60"
            >
              Download {label}
            </button>
          </div>
        ))}
      </div>

      {loading ? <p className="text-xs text-steel">Preparing export</p> : null}
      {status ? <p className="text-xs text-moss">{status}</p> : null}
      {error ? <p className="text-xs text-signal">{error}</p> : null}
    </section>
  );
}
