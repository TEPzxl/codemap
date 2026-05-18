"use client";

import { useCallback, useEffect, useRef, useState } from "react";

import { GraphView } from "@/components/GraphView";
import { SourcePanel } from "@/components/SourcePanel";
import { SymbolSearch } from "@/components/SymbolSearch";
import { Toolbar } from "@/components/Toolbar";
import { WarningPanel } from "@/components/WarningPanel";
import { fetchGraph, fetchSource, fetchSymbols, fetchWarnings, type GraphRequest } from "@/lib/api";
import { inferModulePrefix } from "@/lib/displaySymbol";
import type { Graph, Node as GraphNode, SourceSnippet, SymbolInfo, Warning } from "@/types/graph";

export default function Home() {
  const [symbols, setSymbols] = useState<SymbolInfo[]>([]);
  const [entry, setEntry] = useState("main.main");
  const [depth, setDepth] = useState(5);
  const [packageFilter, setPackageFilter] = useState("");
  const [showExternal, setShowExternal] = useState(false);
  const [showUnresolved, setShowUnresolved] = useState(false);
  const [showInterface, setShowInterface] = useState(false);
  const [expandInterface, setExpandInterface] = useState(false);
  const [graph, setGraph] = useState<Graph | null>(null);
  const [selectedNode, setSelectedNode] = useState<GraphNode | null>(null);
  const [warnings, setWarnings] = useState<Warning[]>([]);
  const [source, setSource] = useState<SourceSnippet | null>(null);
  const [symbolsLoading, setSymbolsLoading] = useState(true);
  const [graphLoading, setGraphLoading] = useState(false);
  const [sourceLoading, setSourceLoading] = useState(false);
  const [apiError, setAPIError] = useState<string | null>(null);
  const [graphError, setGraphError] = useState<string | null>(null);
  const [sourceError, setSourceError] = useState<string | null>(null);
  const graphLoadedRef = useRef(false);
  const graphRequestRef = useRef<GraphRequest>({
    entry: "main.main",
    depth: 5,
  });
  const requestSeqRef = useRef(0);
  const selectedNodeRef = useRef<GraphNode | null>(null);

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
        const initialState = readInitialGraphState();
        const main = symbolResponse.symbols.find((symbol) => symbol.id.endsWith(".main"));
        setEntry(initialState.entry ?? main?.id ?? symbolResponse.symbols[0]?.id ?? "main.main");
        setDepth(initialState.depth ?? 5);
        setPackageFilter(initialState.packageFilter ?? "");
        setShowExternal(initialState.showExternal ?? false);
        setShowUnresolved(initialState.showUnresolved ?? false);
        setShowInterface(initialState.showInterface ?? false);
        setExpandInterface(initialState.expandInterface ?? false);
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

  const modulePrefix = inferModulePrefix(symbols);

  useEffect(() => {
    graphRequestRef.current = {
      entry,
      depth,
      showExternal,
      showUnresolved,
      showInterface,
      expandInterface,
      packagePrefix: packageFilter || undefined,
    };
  }, [depth, entry, expandInterface, packageFilter, showExternal, showInterface, showUnresolved]);

  useEffect(() => {
    selectedNodeRef.current = selectedNode;
  }, [selectedNode]);

  const applyGraph = useCallback((nextGraph: Graph, preserveSelection: boolean) => {
    setGraph(nextGraph);
    setWarnings((current) => mergeWarnings(current, nextGraph.warnings));

    if (preserveSelection) {
      const previous = selectedNodeRef.current;
      const preserved = previous ? (nextGraph.nodes.find((node) => node.id === previous.id) ?? null) : null;
      setSelectedNode(preserved);
      if (!preserved) {
        setSource(null);
        setSourceError(null);
      }
      return;
    }

    setSelectedNode(nextGraph.nodes.find((node) => node.id === nextGraph.entry) ?? nextGraph.nodes[0] ?? null);
    setSource(null);
    setSourceError(null);
  }, []);

  const loadGraph = useCallback(
    async (options?: { entryOverride?: string; preserveSelection?: boolean }) => {
      const request = {
        ...graphRequestRef.current,
        entry: options?.entryOverride ?? graphRequestRef.current.entry,
      };
      const preserveSelection = options?.preserveSelection ?? false;
      const requestID = requestSeqRef.current + 1;
      requestSeqRef.current = requestID;

      setGraphLoading(true);
      setGraphError(null);
      if (!preserveSelection) {
        setSource(null);
        setSourceError(null);
        setSelectedNode(null);
      }
      try {
        writeURLState(request);
        const nextGraph = await fetchGraph(request);
        if (requestSeqRef.current !== requestID) {
          return;
        }
        graphLoadedRef.current = true;
        applyGraph(nextGraph, preserveSelection);
      } catch (error) {
        if (requestSeqRef.current === requestID) {
          setGraphError(error instanceof Error ? error.message : "Failed to load graph");
        }
      } finally {
        if (requestSeqRef.current === requestID) {
          setGraphLoading(false);
        }
      }
    },
    [applyGraph],
  );

  useEffect(() => {
    if (!graphLoadedRef.current) {
      return;
    }
    const timeout = window.setTimeout(() => {
      void loadGraph({ preserveSelection: true });
    }, 350);
    return () => window.clearTimeout(timeout);
  }, [depth, expandInterface, loadGraph, showExternal, showInterface, showUnresolved]);

  function selectSymbol(symbolID: string) {
    setEntry(symbolID);
    void loadGraph({ entryOverride: symbolID });
  }

  async function loadSource(node: GraphNode) {
    setSelectedNode(node);
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

      <section className="grid min-h-0 grid-cols-1 lg:grid-cols-[minmax(420px,460px)_minmax(0,1fr)]">
        <aside className="grid min-w-0 content-start gap-5 overflow-hidden border-b border-line bg-paper/90 p-4 lg:border-b-0 lg:border-r">
          <SymbolSearch
            symbols={symbols}
            value={entry}
            onChange={setEntry}
            onSelectSymbol={selectSymbol}
            packageFilter={packageFilter}
            onPackageFilterChange={setPackageFilter}
            modulePrefix={modulePrefix}
            disabled={symbolsLoading}
          />
          <Toolbar
            depth={depth}
            onDepthChange={setDepth}
            onLoadGraph={() => {
              void loadGraph();
            }}
            showExternal={showExternal}
            showUnresolved={showUnresolved}
            showInterface={showInterface}
            expandInterface={expandInterface}
            onShowExternalChange={setShowExternal}
            onShowUnresolvedChange={setShowUnresolved}
            onShowInterfaceChange={setShowInterface}
            onExpandInterfaceChange={setExpandInterface}
            loading={graphLoading}
          />

          {apiError ? <p className="rounded-md border border-signal/30 bg-orange-50 p-3 text-sm text-signal">{apiError}</p> : null}
          <WarningPanel warnings={warnings} />
        </aside>

        <section className="min-w-0 min-h-[520px]">
          <GraphView
            graph={graph}
            selectedNode={selectedNode}
            modulePrefix={modulePrefix}
            loading={graphLoading}
            error={graphError}
            onNodeSelect={loadSource}
          />
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

interface InitialGraphState {
  entry?: string;
  depth?: number;
  packageFilter?: string;
  showExternal?: boolean;
  showUnresolved?: boolean;
  showInterface?: boolean;
  expandInterface?: boolean;
}

function readInitialGraphState(): InitialGraphState {
  if (typeof window === "undefined") {
    return {};
  }
  const params = new URLSearchParams(window.location.search);
  const rawDepth = params.get("depth");
  const depth = rawDepth === null ? undefined : Number(rawDepth);
  return {
    entry: params.get("entry") ?? undefined,
    depth: depth !== undefined && Number.isInteger(depth) && depth >= 0 ? depth : undefined,
    packageFilter: params.get("package") ?? undefined,
    showExternal: params.get("show_external") === "true" || params.get("showExternal") === "true",
    showUnresolved: params.get("show_unresolved") === "true" || params.get("showUnresolved") === "true",
    showInterface: params.get("show_interface") === "true" || params.get("showInterface") === "true",
    expandInterface: params.get("expand_interface") === "true" || params.get("expandInterface") === "true",
  };
}

function writeURLState(options: GraphRequest) {
  if (typeof window === "undefined") {
    return;
  }
  const params = new URLSearchParams({
    entry: options.entry,
    depth: String(options.depth),
  });
  if (options.showExternal) {
    params.set("show_external", "true");
  }
  if (options.showUnresolved) {
    params.set("show_unresolved", "true");
  }
  if (options.showInterface) {
    params.set("show_interface", "true");
  }
  if (options.expandInterface) {
    params.set("expand_interface", "true");
  }
  if (options.packagePrefix) {
    params.set("package", options.packagePrefix);
  }
  window.history.replaceState(null, "", `${window.location.pathname}?${params.toString()}`);
}
