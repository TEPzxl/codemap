"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";

import { CurrentGraphSearchPanel } from "@/components/CurrentGraphSearchPanel";
import { EntrypointsPanel } from "@/components/EntrypointsPanel";
import { ExportPanel } from "@/components/ExportPanel";
import { GraphSummaryPanel } from "@/components/GraphSummaryPanel";
import { GraphModeToggle, type GraphMode } from "@/components/GraphModeToggle";
import { GraphView } from "@/components/GraphView";
import { PackageGraphView } from "@/components/PackageGraphView";
import { PathSearchPanel } from "@/components/PathSearchPanel";
import { ProjectMetaPanel } from "@/components/ProjectMetaPanel";
import { SourcePanel } from "@/components/SourcePanel";
import { SymbolSearch } from "@/components/SymbolSearch";
import { Toolbar } from "@/components/Toolbar";
import { WarningPanel } from "@/components/WarningPanel";
import {
  fetchGraphExport,
  fetchCallsite,
  fetchEntrypoints,
  fetchGraph,
  fetchMeta,
  fetchPackageGraph,
  fetchPath,
  fetchSource,
  fetchSymbols,
  fetchWarnings,
  rescanProject,
  type GraphRequest,
  type PackageGraphRequest,
} from "@/lib/api";
import { searchCurrentGraphNodes } from "@/lib/currentGraphSearch";
import { displaySymbolID, inferModulePrefix } from "@/lib/displaySymbol";
import { summarizeGraph } from "@/lib/graphSummary";
import { parseViewState, serializeViewState, viewURL } from "@/lib/viewState";
import type {
  Edge as GraphEdge,
  Entrypoint,
  Graph,
  GraphDirection,
  GraphExportFormat,
  Node as GraphNode,
  PackageGraph,
  PackageNode,
  PathResult,
  ProjectMeta,
  SourceView,
  SymbolInfo,
  Warning,
} from "@/types/graph";

export default function Home() {
  const [symbols, setSymbols] = useState<SymbolInfo[]>([]);
  const [entrypoints, setEntrypoints] = useState<Entrypoint[]>([]);
  const [entrypointNote, setEntrypointNote] = useState("");
  const [entry, setEntry] = useState("main.main");
  const [rootEntry, setRootEntry] = useState("main.main");
  const [depth, setDepth] = useState(5);
  const [direction, setDirection] = useState<GraphDirection>("downstream");
  const [packageFilter, setPackageFilter] = useState("");
  const [showExternal, setShowExternal] = useState(false);
  const [showUnresolved, setShowUnresolved] = useState(false);
  const [showInterface, setShowInterface] = useState(false);
  const [expandInterface, setExpandInterface] = useState(false);
  const [graphMode, setGraphMode] = useState<GraphMode>("function");
  const [graph, setGraph] = useState<Graph | null>(null);
  const [packageGraph, setPackageGraph] = useState<PackageGraph | null>(null);
  const [loadedGraphRequest, setLoadedGraphRequest] = useState<GraphRequest | null>(null);
  const [pathFrom, setPathFrom] = useState("main.main");
  const [pathTo, setPathTo] = useState("");
  const [pathMaxDepth, setPathMaxDepth] = useState(8);
  const [pathLimit, setPathLimit] = useState(5);
  const [pathResult, setPathResult] = useState<PathResult | null>(null);
  const [selectedPathIndex, setSelectedPathIndex] = useState(0);
  const [currentGraphQuery, setCurrentGraphQuery] = useState("");
  const [selectedNode, setSelectedNode] = useState<GraphNode | null>(null);
  const [selectedEdgeID, setSelectedEdgeID] = useState<string | null>(null);
  const [selectedPackageID, setSelectedPackageID] = useState<string | null>(null);
  const [warnings, setWarnings] = useState<Warning[]>([]);
  const [meta, setMeta] = useState<ProjectMeta | null>(null);
  const [source, setSource] = useState<SourceView | null>(null);
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  const [symbolsLoading, setSymbolsLoading] = useState(true);
  const [graphLoading, setGraphLoading] = useState(false);
  const [pathLoading, setPathLoading] = useState(false);
  const [sourceLoading, setSourceLoading] = useState(false);
  const [exportLoading, setExportLoading] = useState(false);
  const [rescanLoading, setRescanLoading] = useState(false);
  const [apiError, setAPIError] = useState<string | null>(null);
  const [graphError, setGraphError] = useState<string | null>(null);
  const [pathError, setPathError] = useState<string | null>(null);
  const [sourceError, setSourceError] = useState<string | null>(null);
  const [exportError, setExportError] = useState<string | null>(null);
  const [rescanError, setRescanError] = useState<string | null>(null);
  const [copyStatus, setCopyStatus] = useState<string | null>(null);
  const [exportStatus, setExportStatus] = useState<string | null>(null);
  const graphLoadedRef = useRef(false);
  const graphRequestRef = useRef<GraphRequest>({
    entry: "main.main",
    depth: 5,
    direction: "downstream",
  });
  const initialGraphRequestRef = useRef<GraphRequest | null>(null);
  const requestSeqRef = useRef(0);
  const selectedNodeRef = useRef<GraphNode | null>(null);
  const selectedEdgeRef = useRef<GraphEdge | null>(null);
  const sourceRef = useRef<SourceView | null>(null);

  const modulePrefix = inferModulePrefix(symbols);
  const graphSummary = useMemo(() => summarizeGraph(graph), [graph]);
  const currentGraphResults = useMemo(() => searchCurrentGraphNodes(graph, currentGraphQuery), [currentGraphQuery, graph]);

  const loadProjectIndex = useCallback(async (applyInitialState: boolean) => {
    setSymbolsLoading(true);
    setAPIError(null);
    try {
      const [symbolResponse, entrypointResponse, warningResponse, metaResponse] = await Promise.all([
        fetchSymbols(),
        fetchEntrypoints(),
        fetchWarnings(),
        fetchMeta(),
      ]);
      setSymbols(symbolResponse.symbols);
      setEntrypoints(entrypointResponse.entrypoints);
      setEntrypointNote(entrypointResponse.note);
      setWarnings([...symbolResponse.warnings, ...entrypointResponse.warnings, ...warningResponse.warnings]);
      setMeta(metaResponse);

      if (applyInitialState) {
        const initialState = readInitialGraphState();
        const main = entrypointResponse.entrypoints.find((candidate) => candidate.reasons.includes("main-function"));
        const nextEntry = initialState.entry ?? main?.id ?? symbolResponse.symbols[0]?.id ?? "main.main";
        const nextDepth = initialState.depth ?? 5;
        const nextDirection = initialState.direction ?? "downstream";
        const nextPackageFilter = initialState.packageFilter ?? "";
        const nextShowExternal = initialState.showExternal ?? false;
        const nextShowUnresolved = initialState.showUnresolved ?? false;
        const nextShowInterface = initialState.showInterface ?? false;
        const nextExpandInterface = initialState.expandInterface ?? false;
        setEntry(nextEntry);
        setRootEntry(nextEntry);
        setPathFrom(nextEntry);
        setDepth(nextDepth);
        setDirection(nextDirection);
        setPackageFilter(nextPackageFilter);
        setShowExternal(nextShowExternal);
        setShowUnresolved(nextShowUnresolved);
        setShowInterface(nextShowInterface);
        setExpandInterface(nextExpandInterface);

        if (initialState.entry) {
          initialGraphRequestRef.current = {
            entry: nextEntry,
            depth: nextDepth,
            direction: nextDirection,
            showExternal: nextShowExternal,
            showUnresolved: nextShowUnresolved,
            showInterface: nextShowInterface,
            expandInterface: nextExpandInterface,
            packagePrefix: nextPackageFilter || undefined,
          };
        }
      }
    } catch (error) {
      setAPIError(error instanceof Error ? error.message : "Failed to load project index");
    } finally {
      setSymbolsLoading(false);
    }
  }, []);

  useEffect(() => {
    const timeout = window.setTimeout(() => {
      void loadProjectIndex(true);
    }, 0);
    return () => window.clearTimeout(timeout);
  }, [loadProjectIndex]);

  const buildGraphRequest = useCallback(
    (entryOverride?: string): GraphRequest => ({
      entry: entryOverride ?? entry,
      depth,
      direction,
      showExternal,
      showUnresolved,
      showInterface,
      expandInterface,
      packagePrefix: packageFilter || undefined,
    }),
    [depth, direction, entry, expandInterface, packageFilter, showExternal, showInterface, showUnresolved],
  );

  const buildPackageGraphRequest = useCallback(
    (entryOverride?: string): PackageGraphRequest => ({
      entry: entryOverride ?? entry,
      depth,
      direction,
      showExternal,
      showUnresolved,
      showInterface,
      expandInterface,
      packagePrefix: packageFilter || undefined,
    }),
    [depth, direction, entry, expandInterface, packageFilter, showExternal, showInterface, showUnresolved],
  );

  useEffect(() => {
    selectedNodeRef.current = selectedNode;
  }, [selectedNode]);

  useEffect(() => {
    sourceRef.current = source;
  }, [source]);

  const applyGraph = useCallback((nextGraph: Graph, preserveSelection: boolean) => {
    setGraph(nextGraph);
    setWarnings((current) => mergeWarnings(current, nextGraph.warnings));

    if (preserveSelection) {
      const previous = selectedNodeRef.current;
      const previousEdge = selectedEdgeRef.current;
      const currentSource = sourceRef.current;
      const preserved = previous ? (nextGraph.nodes.find((node) => node.id === previous.id) ?? null) : null;
      const preservedEdge = previousEdge ? (nextGraph.edges.find((edge) => sameEdge(edge, previousEdge)) ?? null) : null;
      setSelectedNode(preserved);
      selectedEdgeRef.current = preservedEdge;
      setSelectedEdgeID(preservedEdge?.id ?? null);
      if (currentSource?.mode === "node" && !preserved) {
        setSource(null);
        setSourceError(null);
      }
      if (currentSource?.mode === "callsite" && !preservedEdge) {
        setSource(null);
        setSourceError(null);
      }
      return;
    }

    setSelectedNode(nextGraph.nodes.find((node) => node.id === nextGraph.entry) ?? nextGraph.nodes[0] ?? null);
    selectedEdgeRef.current = null;
    setSelectedEdgeID(null);
    setSource(null);
    setSourceError(null);
  }, []);

  const loadGraphRequest = useCallback(
    async (request: GraphRequest, options?: { preserveSelection?: boolean }) => {
      const preserveSelection = options?.preserveSelection ?? false;
      const requestID = requestSeqRef.current + 1;
      requestSeqRef.current = requestID;

      setGraphLoading(true);
      setGraphError(null);
      if (!preserveSelection) {
        setSource(null);
        setSourceError(null);
        setSelectedNode(null);
        selectedEdgeRef.current = null;
        setSelectedEdgeID(null);
      }
      try {
        graphRequestRef.current = request;
        writeURLState(request);
        const nextGraph = await fetchGraph(request);
        if (requestSeqRef.current !== requestID) {
          return;
        }
        graphLoadedRef.current = true;
        graphRequestRef.current = request;
        setLoadedGraphRequest(request);
        applyGraph(nextGraph, preserveSelection);
        setPathResult(null);
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

  const loadGraph = useCallback(
    async (options?: { entryOverride?: string; preserveSelection?: boolean }) => {
      const request = buildGraphRequest(options?.entryOverride);
      await loadGraphRequest(request, { preserveSelection: options?.preserveSelection });
    },
    [buildGraphRequest, loadGraphRequest],
  );

  const loadPackageGraphRequest = useCallback(
    async (request: PackageGraphRequest) => {
      const requestID = requestSeqRef.current + 1;
      requestSeqRef.current = requestID;
      const urlRequest: GraphRequest = {
        entry: request.entry ?? entry,
        depth: request.depth,
        direction: request.direction,
        showExternal: request.showExternal,
        showUnresolved: request.showUnresolved,
        showInterface: request.showInterface,
        expandInterface: request.expandInterface,
        packagePrefix: request.packagePrefix,
      };

      setGraphLoading(true);
      setGraphError(null);
      setSource(null);
      setSourceError(null);
      setSelectedNode(null);
      selectedEdgeRef.current = null;
      setSelectedEdgeID(null);
      try {
        graphRequestRef.current = urlRequest;
        writeURLState(urlRequest);
        const nextGraph = await fetchPackageGraph(request);
        if (requestSeqRef.current !== requestID) {
          return;
        }
        graphLoadedRef.current = true;
        graphRequestRef.current = urlRequest;
        setLoadedGraphRequest(urlRequest);
        setPackageGraph(nextGraph);
        setWarnings((current) => mergeWarnings(current, nextGraph.warnings));
        setPathResult(null);
      } catch (error) {
        if (requestSeqRef.current === requestID) {
          setGraphError(error instanceof Error ? error.message : "Failed to load package graph");
        }
      } finally {
        if (requestSeqRef.current === requestID) {
          setGraphLoading(false);
        }
      }
    },
    [entry],
  );

  const loadPackageGraph = useCallback(
    async (entryOverride?: string) => {
      await loadPackageGraphRequest(buildPackageGraphRequest(entryOverride));
    },
    [buildPackageGraphRequest, loadPackageGraphRequest],
  );

  useEffect(() => {
    if (!graphLoadedRef.current) {
      return;
    }
    const timeout = window.setTimeout(() => {
      if (graphMode === "package") {
        void loadPackageGraph();
      } else {
        void loadGraph({ preserveSelection: true });
      }
    }, 350);
    return () => window.clearTimeout(timeout);
  }, [depth, direction, expandInterface, graphMode, loadGraph, loadPackageGraph, packageFilter, showExternal, showInterface, showUnresolved]);

  useEffect(() => {
    if (symbolsLoading || graphLoadedRef.current) {
      return;
    }
    const initialRequest = initialGraphRequestRef.current;
    if (!initialRequest) {
      return;
    }
    initialGraphRequestRef.current = null;
    void loadGraphRequest(initialRequest);
  }, [loadGraphRequest, symbolsLoading]);

  function selectSymbol(symbolID: string) {
    setGraphMode("function");
    setEntry(symbolID);
    setRootEntry(symbolID);
    setPathFrom(symbolID);
    setDirection("downstream");
    void loadGraphRequest({ ...buildGraphRequest(symbolID), direction: "downstream" });
  }

  function manuallyLoadGraph() {
    setGraphMode("function");
    const resolvedEntry = resolveEntryInput(entry, symbols, modulePrefix);
    if (resolvedEntry !== entry) {
      setEntry(resolvedEntry);
    }
    setRootEntry(resolvedEntry);
    setPathFrom(resolvedEntry);
    setDirection("downstream");
    void loadGraphRequest({ ...buildGraphRequest(resolvedEntry), direction: "downstream" });
  }

  async function findPath() {
    setGraphMode("function");
    const resolvedFrom = resolveEntryInput(pathFrom, symbols, modulePrefix);
    const resolvedTo = resolveEntryInput(pathTo, symbols, modulePrefix);
    setPathFrom(resolvedFrom);
    setPathTo(resolvedTo);
    setPathLoading(true);
    setPathError(null);
    try {
      const result = await fetchPath({
        from: resolvedFrom,
        to: resolvedTo,
        maxDepth: pathMaxDepth,
        limit: pathLimit,
        showExternal,
        showUnresolved,
        showInterface,
        expandInterface,
        packagePrefix: packageFilter || undefined,
      });
      setPathResult(result);
      setSelectedPathIndex(0);
      setGraph(result.graph);
      setWarnings((current) => mergeWarnings(current, result.warnings));
      setLoadedGraphRequest({
        entry: result.from,
        depth: pathMaxDepth,
        direction: "downstream",
        showExternal,
        showUnresolved,
        showInterface,
        expandInterface,
        packagePrefix: packageFilter || undefined,
      });
      graphLoadedRef.current = true;
      graphRequestRef.current = {
        entry: result.from,
        depth: pathMaxDepth,
        direction: "downstream",
        showExternal,
        showUnresolved,
        showInterface,
        expandInterface,
        packagePrefix: packageFilter || undefined,
      };
      if (result.paths.length > 0) {
        setSelectedEdgeID(result.paths[0].edges[0] ?? null);
        setSelectedNode(result.graph.nodes.find((node) => node.id === result.to) ?? result.graph.nodes[0] ?? null);
      } else {
        setSelectedEdgeID(null);
        setSelectedNode(null);
      }
      selectedEdgeRef.current = null;
      setSource(null);
      setSourceError(null);
    } catch (error) {
      setPathError(error instanceof Error ? error.message : "Failed to find path");
    } finally {
      setPathLoading(false);
    }
  }

  function selectPath(index: number) {
    setSelectedPathIndex(index);
    const edgeID = pathResult?.paths[index]?.edges[0] ?? null;
    setSelectedEdgeID(edgeID);
  }

  function focusNode(node: GraphNode, nextDirection: GraphDirection) {
    setGraphMode("function");
    setEntry(node.id);
    setDirection(nextDirection);
    void loadGraphRequest({ ...buildGraphRequest(node.id), direction: nextDirection }, { preserveSelection: true });
  }

  function resetToEntry() {
    setGraphMode("function");
    setEntry(rootEntry);
    setDirection("downstream");
    void loadGraphRequest({ ...buildGraphRequest(rootEntry), direction: "downstream" }, { preserveSelection: true });
  }

  async function handleRescan() {
    setRescanLoading(true);
    setRescanError(null);
    try {
      const response = await rescanProject();
      setMeta(response.meta);
      await loadProjectIndex(false);
      if (graphLoadedRef.current) {
        if (graphMode === "package") {
          await loadPackageGraph();
        } else {
          await loadGraph({ preserveSelection: true });
        }
      }
    } catch (error) {
      setRescanError(error instanceof Error ? error.message : "Failed to rescan project");
    } finally {
      setRescanLoading(false);
    }
  }

  async function loadSource(node: GraphNode) {
    setSelectedNode(node);
    selectedEdgeRef.current = null;
    setSelectedEdgeID(null);
    setSourceLoading(true);
    setSourceError(null);
    setSource(null);
    try {
      const snippet = await fetchSource(node.id);
      setSource({ mode: "node", data: snippet });
    } catch (error) {
      setSourceError(error instanceof Error ? error.message : "Failed to load source");
    } finally {
      setSourceLoading(false);
    }
  }

  function jumpToGraphNode(node: GraphNode) {
    setGraphMode("function");
    void loadSource(node);
  }

  async function loadCallsite(edge: GraphEdge) {
    setSelectedNode(null);
    selectedEdgeRef.current = edge;
    setSelectedEdgeID(edge.id);
    setSourceLoading(true);
    setSourceError(null);
    setSource(null);
    try {
      const snippet = await fetchCallsite(edge.id, graphRequestRef.current);
      setSource({ mode: "callsite", data: snippet });
    } catch (error) {
      setSourceError(error instanceof Error ? error.message : "Failed to load callsite");
    } finally {
      setSourceLoading(false);
    }
  }

  async function copyViewURL() {
    if (typeof window === "undefined") {
      return;
    }
    const request = loadedGraphRequest ?? buildGraphRequest();
    const nextURL = `${window.location.origin}${viewURL(window.location.pathname, request)}`;
    try {
      await navigator.clipboard.writeText(nextURL);
      setCopyStatus("Copied current view URL");
    } catch {
      setCopyStatus("Copy failed");
    }
    window.setTimeout(() => setCopyStatus(null), 2000);
  }

  async function copyGraphExport(format: GraphExportFormat) {
    setExportLoading(true);
    setExportError(null);
    setExportStatus(null);
    try {
      const output = await fetchGraphExport(visibleGraphRequest, format);
      await navigator.clipboard.writeText(output);
      setExportStatus(`Copied ${format.toUpperCase()}`);
      window.setTimeout(() => setExportStatus(null), 2000);
    } catch (error) {
      setExportError(error instanceof Error ? error.message : `Failed to copy ${format}`);
    } finally {
      setExportLoading(false);
    }
  }

  async function downloadGraphExport(format: GraphExportFormat) {
    setExportLoading(true);
    setExportError(null);
    setExportStatus(null);
    try {
      const output = await fetchGraphExport(visibleGraphRequest, format);
      downloadTextFile(output, exportFilename(format), exportMimeType(format));
      setExportStatus(`Downloaded ${format.toUpperCase()}`);
      window.setTimeout(() => setExportStatus(null), 2000);
    } catch (error) {
      setExportError(error instanceof Error ? error.message : `Failed to download ${format}`);
    } finally {
      setExportLoading(false);
    }
  }

  function changeGraphMode(nextMode: GraphMode) {
    setGraphMode(nextMode);
    if (!graphLoadedRef.current && nextMode === "package") {
      void loadPackageGraph();
    }
  }

  function selectPackage(node: PackageNode) {
    setSelectedPackageID(node.id);
    setPackageFilter(packageFilterValue(node));
  }

  function openPackage(node: PackageNode) {
    const prefix = packageFilterValue(node);
    setSelectedPackageID(node.id);
    setPackageFilter(prefix);
    setGraphMode("function");
    void loadGraphRequest({ ...buildGraphRequest(), packagePrefix: prefix });
  }

  const visibleGraphRequest = loadedGraphRequest ?? buildGraphRequest();

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

      <section
        className={
          sidebarCollapsed
            ? "grid min-h-0 grid-cols-1"
            : "grid min-h-0 grid-cols-1 lg:grid-cols-[minmax(360px,460px)_minmax(0,1fr)]"
        }
      >
        {!sidebarCollapsed ? (
          <aside className="grid min-h-0 min-w-0 content-start gap-5 overflow-y-auto overflow-x-hidden border-b border-line bg-paper/90 p-4 lg:border-b-0 lg:border-r">
            <ProjectMetaPanel meta={meta} loading={rescanLoading} error={rescanError} onRescan={handleRescan} />
            <GraphModeToggle mode={graphMode} onChange={changeGraphMode} />
            <EntrypointsPanel
              entrypoints={entrypoints}
              note={entrypointNote}
              modulePrefix={modulePrefix}
              loading={symbolsLoading}
              disabled={graphLoading}
              onSelectEntrypoint={selectSymbol}
            />
            <GraphSummaryPanel
              entry={graph?.entry ?? visibleGraphRequest.entry}
              request={visibleGraphRequest}
              summary={graphSummary}
              loading={graphLoading}
              copyStatus={copyStatus}
              selectedNode={selectedNode}
              onCopyViewURL={copyViewURL}
              onFocusNode={focusNode}
              onResetToEntry={resetToEntry}
            />
            <ExportPanel
              loading={exportLoading}
              disabled={graphLoading || !loadedGraphRequest}
              status={exportStatus}
              error={exportError}
              onCopyExport={copyGraphExport}
              onDownloadExport={downloadGraphExport}
              onCopyViewURL={copyViewURL}
            />
            {graphMode === "function" ? (
              <CurrentGraphSearchPanel
                query={currentGraphQuery}
                results={currentGraphResults}
                selectedNodeID={selectedNode?.id}
                modulePrefix={modulePrefix}
                disabled={graphLoading || !graph}
                onQueryChange={setCurrentGraphQuery}
                onSelectNode={jumpToGraphNode}
              />
            ) : null}
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
              onLoadGraph={manuallyLoadGraph}
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
            <PathSearchPanel
              symbols={symbols}
              from={pathFrom}
              to={pathTo}
              maxDepth={pathMaxDepth}
              limit={pathLimit}
              loading={pathLoading}
              error={pathError}
              result={pathResult}
              selectedPathIndex={selectedPathIndex}
              modulePrefix={modulePrefix}
              disabled={symbolsLoading}
              onFromChange={setPathFrom}
              onToChange={setPathTo}
              onMaxDepthChange={setPathMaxDepth}
              onLimitChange={setPathLimit}
              onFindPath={findPath}
              onSelectPath={selectPath}
            />

            {apiError ? <p className="rounded-md border border-signal/30 bg-orange-50 p-3 text-sm text-signal">{apiError}</p> : null}
            <WarningPanel warnings={warnings} />
          </aside>
        ) : null}

        <section className="relative min-h-[520px] min-w-0">
          <button
            type="button"
            onClick={() => setSidebarCollapsed((value) => !value)}
            aria-label={sidebarCollapsed ? "Show control panel" : "Hide control panel"}
            title={sidebarCollapsed ? "Show panel" : "Hide panel"}
            className="absolute left-4 top-4 z-20 grid h-9 w-9 place-items-center rounded-md border border-line bg-white font-mono text-lg font-semibold leading-none text-ink shadow-sm transition hover:border-moss hover:text-moss"
          >
            {sidebarCollapsed ? ">" : "<"}
          </button>
          {graphMode === "package" ? (
            <PackageGraphView
              key={sidebarCollapsed ? "package-graph-collapsed" : "package-graph-expanded"}
              graph={packageGraph}
              selectedPackageID={selectedPackageID}
              loading={graphLoading}
              error={graphError}
              onPackageSelect={selectPackage}
              onPackageOpen={openPackage}
            />
          ) : (
            <GraphView
              key={sidebarCollapsed ? "graph-collapsed" : "graph-expanded"}
              graph={graph}
              selectedNode={selectedNode}
              selectedEdgeID={selectedEdgeID}
              modulePrefix={modulePrefix}
              loading={graphLoading}
              error={graphError}
              onNodeSelect={loadSource}
              onEdgeSelect={loadCallsite}
            />
          )}
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

function sameEdge(left: GraphEdge, right: GraphEdge): boolean {
  return (
    left.id === right.id &&
    left.from === right.from &&
    left.to === right.to &&
    left.resolution === right.resolution &&
    left.candidate === right.candidate &&
    left.callsite.file === right.callsite.file &&
    left.callsite.line === right.callsite.line &&
    left.callsite.column === right.callsite.column
  );
}

function packageFilterValue(node: PackageNode): string {
  return node.full_package || node.package;
}

function resolveEntryInput(value: string, symbols: SymbolInfo[], modulePrefix: string): string {
  const query = value.trim();
  if (!query) {
    return value;
  }

  const matches = symbols.filter((symbol) => {
    return (
      symbol.id === query ||
      displaySymbolID(symbol.id, modulePrefix) === query ||
      symbol.id.endsWith(`.${query}`) ||
      symbol.label === query
    );
  });

  return matches.length === 1 ? matches[0].id : query;
}

interface InitialGraphState {
  entry?: string;
  depth?: number;
  direction?: GraphDirection;
  packageFilter?: string;
  showExternal?: boolean;
  showUnresolved?: boolean;
  showInterface?: boolean;
  expandInterface?: boolean;
}

function readInitialGraphState(): InitialGraphState {
  if (typeof window === "undefined") {
    return {
      showExternal: false,
      showUnresolved: false,
      showInterface: false,
      expandInterface: false,
    };
  }
  return parseViewState(window.location.search);
}

function writeURLState(options: GraphRequest) {
  if (typeof window === "undefined") {
    return;
  }
  window.history.replaceState(null, "", `${window.location.pathname}?${serializeViewState(options).toString()}`);
}

function exportFilename(format: GraphExportFormat): string {
  const extension = format === "mermaid" ? "mmd" : format;
  return `codemap-graph.${extension}`;
}

function exportMimeType(format: GraphExportFormat): string {
  return format === "json" ? "application/json" : "text/plain";
}

function downloadTextFile(content: string, filename: string, mimeType: string) {
  if (typeof window === "undefined") {
    return;
  }
  const blob = new Blob([content], { type: `${mimeType};charset=utf-8` });
  const url = window.URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  link.remove();
  window.URL.revokeObjectURL(url);
}
