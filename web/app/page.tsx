"use client";

import { useEffect, useMemo, useState } from "react";

import { GraphView } from "@/components/GraphView";
import { SourcePanel } from "@/components/SourcePanel";
import { SymbolSearch } from "@/components/SymbolSearch";
import { Toolbar } from "@/components/Toolbar";
import { WarningPanel } from "@/components/WarningPanel";
import { fetchGraph, fetchSource, fetchSymbols, fetchWarnings } from "@/lib/api";
import type { Graph, Node as GraphNode, SourceSnippet, SymbolInfo, Warning } from "@/types/graph";

export default function Home() {
  const [symbols, setSymbols] = useState<SymbolInfo[]>([]);
  const [entry, setEntry] = useState("main.main");
  const [depth, setDepth] = useState(5);
  const [graph, setGraph] = useState<Graph | null>(null);
  const [warnings, setWarnings] = useState<Warning[]>([]);
  const [source, setSource] = useState<SourceSnippet | null>(null);
  const [symbolsLoading, setSymbolsLoading] = useState(true);
  const [graphLoading, setGraphLoading] = useState(false);
  const [sourceLoading, setSourceLoading] = useState(false);
  const [apiError, setAPIError] = useState<string | null>(null);
  const [graphError, setGraphError] = useState<string | null>(null);
  const [sourceError, setSourceError] = useState<string | null>(null);

  useEffect(() => {
    let active = true;

    async function loadProjectIndex() {
      setSymbolsLoading(true);
      setAPIError(null);
      try {
        const [symbolResponse, warningResponse] = await Promise.all([fetchSymbols(), fetchWarnings()]);
        if (!active) {
          return;
        }
        setSymbols(symbolResponse.symbols);
        setWarnings([...symbolResponse.warnings, ...warningResponse.warnings]);
        const main = symbolResponse.symbols.find((symbol) => symbol.id.endsWith(".main"));
        setEntry(main?.id ?? symbolResponse.symbols[0]?.id ?? "main.main");
      } catch (error) {
        if (active) {
          setAPIError(error instanceof Error ? error.message : "Failed to load symbols");
        }
      } finally {
        if (active) {
          setSymbolsLoading(false);
        }
      }
    }

    loadProjectIndex();
    return () => {
      active = false;
    };
  }, []);

  const selectedSymbol = useMemo(() => symbols.find((symbol) => symbol.id === entry), [entry, symbols]);

  async function loadGraph() {
    setGraphLoading(true);
    setGraphError(null);
    setSource(null);
    setSourceError(null);
    try {
      const nextGraph = await fetchGraph(entry, depth);
      setGraph(nextGraph);
      setWarnings((current) => mergeWarnings(current, nextGraph.warnings));
    } catch (error) {
      setGraphError(error instanceof Error ? error.message : "Failed to load graph");
    } finally {
      setGraphLoading(false);
    }
  }

  async function loadSource(node: GraphNode) {
    setSourceLoading(true);
    setSourceError(null);
    setSource(null);
    try {
      const snippet = await fetchSource(node.id);
      setSource(snippet);
    } catch (error) {
      setSourceError(error instanceof Error ? error.message : "Failed to load source");
    } finally {
      setSourceLoading(false);
    }
  }

  return (
    <main className="grid min-h-screen grid-rows-[auto_1fr_auto]">
      <header className="border-b border-line bg-paper/95 px-5 py-4 backdrop-blur">
        <div className="flex flex-wrap items-end justify-between gap-3">
          <div>
            <h1 className="text-2xl font-bold text-ink">codemap</h1>
            <p className="mt-1 text-sm text-steel">Go call graph explorer</p>
          </div>
          <div className="rounded-md border border-line bg-white px-3 py-2 font-mono text-xs text-steel">API /api/*</div>
        </div>
      </header>

      <section className="grid min-h-0 grid-cols-1 lg:grid-cols-[340px_1fr]">
        <aside className="grid content-start gap-5 border-b border-line bg-paper/90 p-4 lg:border-b-0 lg:border-r">
          <SymbolSearch symbols={symbols} value={entry} onChange={setEntry} disabled={symbolsLoading} />
          <Toolbar depth={depth} onDepthChange={setDepth} onLoadGraph={loadGraph} loading={graphLoading} />

          <section className="rounded-md border border-line bg-white p-3">
            <h2 className="text-sm font-semibold text-ink">Selected</h2>
            <p className="mt-2 break-all font-mono text-xs text-steel">{entry}</p>
            {selectedSymbol ? <p className="mt-2 text-sm text-steel">{selectedSymbol.file}</p> : null}
          </section>

          {apiError ? <p className="rounded-md border border-signal/30 bg-orange-50 p-3 text-sm text-signal">{apiError}</p> : null}
          <WarningPanel warnings={warnings} />
        </aside>

        <section className="min-h-[520px]">
          <GraphView graph={graph} loading={graphLoading} error={graphError} onNodeSelect={loadSource} />
        </section>
      </section>

      <SourcePanel source={source} loading={sourceLoading} error={sourceError} />
    </main>
  );
}

function mergeWarnings(existing: Warning[], incoming: Warning[]): Warning[] {
  const seen = new Set<string>();
  const merged: Warning[] = [];
  for (const warning of [...existing, ...incoming]) {
    const key = `${warning.code}:${warning.message}:${warning.file ?? ""}`;
    if (!seen.has(key)) {
      seen.add(key);
      merged.push(warning);
    }
  }
  return merged;
}
