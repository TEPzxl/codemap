"use client";

import type { Warning } from "@/types/graph";

interface WarningPanelProps {
  warnings: Warning[];
}

export function WarningPanel({ warnings }: WarningPanelProps) {
  return (
    <section className="grid gap-3">
      <div className="flex items-center justify-between">
        <h2 className="text-sm font-semibold text-ink">Warnings</h2>
        <span className="rounded border border-line bg-white px-2 py-0.5 font-mono text-xs text-steel">
          {warnings.length}
        </span>
      </div>
      <div className="max-h-36 overflow-auto rounded-md border border-line bg-white">
        {warnings.length === 0 ? (
          <p className="px-3 py-3 text-sm text-steel">No warnings</p>
        ) : (
          <ul className="divide-y divide-line">
            {warnings.map((warning, index) => (
              <li key={`${warning.code}-${index}`} className="px-3 py-2">
                <p className="font-mono text-xs font-semibold text-signal">{warning.code}</p>
                <p className="mt-1 text-sm text-ink">{warning.message}</p>
                {warning.file ? <p className="mt-1 font-mono text-xs text-steel">{warning.file}</p> : null}
              </li>
            ))}
          </ul>
        )}
      </div>
    </section>
  );
}
