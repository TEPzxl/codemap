"use client";

import { useCallback, useEffect, useRef, useState } from "react";

import { clampSourcePanelHeight, sourcePanelDefaultHeight } from "@/lib/sourcePanelSizing";
import type { SourceView } from "@/types/graph";

interface SourcePanelProps {
  source: SourceView | null;
  loading?: boolean;
  error?: string | null;
}

export function SourcePanel({ source, loading, error }: SourcePanelProps) {
  const [height, setHeight] = useState(sourcePanelDefaultHeight);
  const dragStartRef = useRef<{ y: number; height: number } | null>(null);
  const data = source?.data ?? null;
  const title = source?.mode === "callsite" ? "Callsite" : "Node source";
  const location = source ? sourceLocation(source) : "No graph item selected";

  const stopResize = useCallback(() => {
    dragStartRef.current = null;
    document.body.style.cursor = "";
    document.body.style.userSelect = "";
  }, []);

  useEffect(() => {
    function handlePointerMove(event: PointerEvent) {
      const start = dragStartRef.current;
      if (!start) {
        return;
      }
      const nextHeight = start.height - (event.clientY - start.y);
      setHeight(clampSourcePanelHeight(nextHeight, window.innerHeight));
    }

    window.addEventListener("pointermove", handlePointerMove);
    window.addEventListener("pointerup", stopResize);
    window.addEventListener("pointercancel", stopResize);
    return () => {
      window.removeEventListener("pointermove", handlePointerMove);
      window.removeEventListener("pointerup", stopResize);
      window.removeEventListener("pointercancel", stopResize);
    };
  }, [stopResize]);

  useEffect(() => {
    function handleResize() {
      setHeight((current) => clampSourcePanelHeight(current, window.innerHeight));
    }

    window.addEventListener("resize", handleResize);
    return () => window.removeEventListener("resize", handleResize);
  }, []);

  return (
    <section className="relative flex min-h-40 flex-col border-t border-line bg-[#111820] text-stone-100" style={{ height }}>
      <button
        type="button"
        aria-label="Resize source panel"
        title="Drag to resize source panel"
        onPointerDown={(event) => {
          dragStartRef.current = { y: event.clientY, height };
          document.body.style.cursor = "row-resize";
          document.body.style.userSelect = "none";
        }}
        onDoubleClick={() => setHeight(sourcePanelDefaultHeight)}
        className="absolute -top-1 left-0 z-10 h-2 w-full cursor-row-resize bg-transparent transition hover:bg-moss/20"
      />
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

      <div className="min-h-0 flex-1 overflow-auto p-4">
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
