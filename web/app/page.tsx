"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";

import { CurrentGraphSearchPanel } from "@/components/CurrentGraphSearchPanel";
import { EntrypointsPanel } from "@/components/EntrypointsPanel";
import { GraphSummaryPanel } from "@/components/GraphSummaryPanel";
import { GraphModeToggle } from "@/components/GraphModeToggle";
import { GraphView } from "@/components/GraphView";
import { PackageGraphView } from "@/components/PackageGraphView";
import { PathSearchPanel } from "@/components/PathSearchPanel";
import { ProjectMetaPanel } from "@/components/ProjectMetaPanel";
import { SourcePanel } from "@/components/SourcePanel";
import { SymbolSearch } from "@/components/SymbolSearch";
import { Toolbar } from "@/components/Toolbar";
import { WarningPanel } from "@/components/WarningPanel";
import {
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
import { entrypointTargetMode, type GraphMode } from "@/lib/entrypointSelection";
import { summarizeGraph } from "@/lib/graphSummary";
import { filterGraphByPackage } from "@/lib/packageDrilldown";
import { parseViewState, serializeViewState } from "@/lib/viewState";
import type {
  Edge as GraphEdge,
  Entrypoint,
  Graph,
  GraphDirection,
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
  const [graphFullscreen, setGraphFullscreen] = useState(false);
  const [symbolsLoading, setSymbolsLoading] = useState(true);
  const [graphLoading, setGraphLoading] = useState(false);
  const [pathLoading, setPathLoading] = useState(false);
  const [sourceLoading, setSourceLoading] = useState(false);
  const [rescanLoading, setRescanLoading] = useState(false);
  const [apiError, setAPIError] = useState<string | null>(null);
  const [graphError, setGraphError] = useState<string | null>(null);
  const [pathError, setPathError] = useState<string | null>(null);
  const [sourceError, setSourceError] = useState<string | null>(null);
  const [rescanError, setRescanError] = useState<string | null>(null);
  const graphLoadedRef = useRef(false);
  const graphRequestRef = useRef<GraphRequest>({
    entry: "main.main",
    depth: 5,
    direction: "downstream",
  });
  const initialGraphRequestRef = useRef<GraphRequest | null>(null);
  const skipNextAutoRefreshRef = useRef(false);
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

  useEffect(() => {
    if (!graphFullscreen) {
      return;
    }
    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === "Escape") {
        setGraphFullscreen(false);
      }
    }
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [graphFullscreen]);

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
    if (skipNextAutoRefreshRef.current) {
      skipNextAutoRefreshRef.current = false;
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
    const nextMode = entrypointTargetMode(graphMode);
    setGraphMode(nextMode);
    setEntry(symbolID);
    setRootEntry(symbolID);
    setPathFrom(symbolID);
    setDirection("downstream");
    if (nextMode === "package") {
      void loadPackageGraphRequest({ ...buildPackageGraphRequest(symbolID), direction: "downstream" });
      return;
    }
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
    if (!graph) {
      void loadGraphRequest({ ...buildGraphRequest(), packagePrefix: prefix });
      return;
    }

    const filteredGraph = filterGraphByPackage(graph, prefix);
    if (filteredGraph.nodes.length === 0) {
      void loadGraphRequest({ ...buildGraphRequest(), packagePrefix: prefix });
      return;
    }

    skipNextAutoRefreshRef.current = true;
    const nextRequest = {
      ...(loadedGraphRequest ?? buildGraphRequest()),
      entry: filteredGraph.entry,
      packagePrefix: prefix,
    };
    graphRequestRef.current = nextRequest;
    setLoadedGraphRequest(nextRequest);
    writeURLState(nextRequest);
    applyGraph(filteredGraph, false);
    setPathResult(null);
  }

  const visibleGraphRequest = loadedGraphRequest ?? buildGraphRequest();
  const graphSurface = (
    <section className={graphFullscreen ? "fixed inset-0 z-50 min-h-0 min-w-0 overflow-hidden bg-paper" : "relative min-h-0 min-w-0 overflow-hidden"}>
      {!graphFullscreen ? (
        <button
          type="button"
          onClick={() => setSidebarCollapsed((value) => !value)}
          aria-label={sidebarCollapsed ? "Show control panel" : "Hide control panel"}
          title={sidebarCollapsed ? "Show panel" : "Hide panel"}
          className="absolute left-4 top-4 z-20 grid h-9 w-9 place-items-center rounded-md border border-line bg-white font-mono text-lg font-semibold leading-none text-ink shadow-sm transition hover:border-moss hover:text-moss"
        >
          {sidebarCollapsed ? ">" : "<"}
        </button>
      ) : null}
      <FullscreenGraphButton fullscreen={graphFullscreen} onClick={() => setGraphFullscreen((value) => !value)} />
      {graphMode === "package" ? (
        <PackageGraphView
          key={`${graphFullscreen ? "fullscreen" : "windowed"}-${sidebarCollapsed ? "package-graph-collapsed" : "package-graph-expanded"}`}
          graph={packageGraph}
          selectedPackageID={selectedPackageID}
          loading={graphLoading}
          error={graphError}
          onPackageSelect={selectPackage}
          onPackageOpen={openPackage}
        />
      ) : (
        <GraphView
          key={`${graphFullscreen ? "fullscreen" : "windowed"}-${sidebarCollapsed ? "graph-collapsed" : "graph-expanded"}`}
          graph={graph}
          selectedNode={selectedNode}
          selectedEdgeID={selectedEdgeID}
          loading={graphLoading}
          error={graphError}
          onNodeSelect={loadSource}
          onEdgeSelect={loadCallsite}
        />
      )}
    </section>
  );

  if (graphFullscreen) {
    return graphSurface;
  }

  return (
    <main className="grid h-screen min-h-0 grid-rows-[auto_minmax(0,1fr)_auto] overflow-hidden">
      <header className="border-b border-line bg-paper/95 px-4 py-2.5 backdrop-blur">
        <div className="flex flex-wrap items-center justify-between gap-2">
          <div>
            <h1 className="text-xl font-bold text-ink">codemap</h1>
            <p className="text-xs text-steel">Go call graph explorer</p>
          </div>
          <HeaderSelectionSummary node={selectedNode} modulePrefix={modulePrefix} />
        </div>
      </header>

      <section
        className={
          sidebarCollapsed
            ? "grid min-h-0 grid-cols-1 overflow-hidden"
            : "grid min-h-0 grid-cols-1 overflow-hidden lg:grid-cols-[minmax(360px,460px)_minmax(0,1fr)]"
        }
      >
        {!sidebarCollapsed ? (
          <aside className="grid h-full min-h-0 min-w-0 content-start gap-5 overflow-y-auto overflow-x-hidden border-b border-line bg-paper/90 p-4 lg:border-b-0 lg:border-r">
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
              selectedNode={selectedNode}
              onFocusNode={focusNode}
              onResetToEntry={resetToEntry}
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

        {graphSurface}
      </section>

      <SourcePanel source={source} loading={sourceLoading} error={sourceError} />
    </main>
  );
}

function FullscreenGraphButton({ fullscreen, onClick }: { fullscreen: boolean; onClick: () => void }) {
  return (
    <button
      type="button"
      onClick={onClick}
      aria-label={fullscreen ? "Exit graph fullscreen" : "Show graph fullscreen"}
      title={fullscreen ? "Exit fullscreen" : "Fullscreen graph"}
      className="absolute right-16 top-4 z-20 grid h-9 w-9 place-items-center rounded-md border border-line bg-white font-mono text-lg font-semibold leading-none text-ink shadow-sm transition hover:border-moss hover:text-moss"
    >
      {fullscreen ? <ExitFullscreenIcon /> : <EnterFullscreenIcon />}
    </button>
  );
}

function EnterFullscreenIcon() {
  return (
    <span aria-hidden="true" className="relative block h-4 w-4">
      <span className="absolute left-0 top-0 h-1.5 w-1.5 border-l-2 border-t-2 border-current" />
      <span className="absolute right-0 top-0 h-1.5 w-1.5 border-r-2 border-t-2 border-current" />
      <span className="absolute bottom-0 left-0 h-1.5 w-1.5 border-b-2 border-l-2 border-current" />
      <span className="absolute bottom-0 right-0 h-1.5 w-1.5 border-b-2 border-r-2 border-current" />
    </span>
  );
}

function ExitFullscreenIcon() {
  return (
    <span aria-hidden="true" className="relative block h-4 w-4">
      <span className="absolute left-1/2 top-0 h-4 w-0.5 -translate-x-1/2 rotate-45 rounded bg-current" />
      <span className="absolute left-1/2 top-0 h-4 w-0.5 -translate-x-1/2 -rotate-45 rounded bg-current" />
    </span>
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

function HeaderSelectionSummary({ node, modulePrefix }: { node: GraphNode | null; modulePrefix: string }) {
  if (!node) {
    return <div className="rounded-md border border-line bg-white px-2.5 py-1.5 font-mono text-xs text-steel">API /api/*</div>;
  }

  return (
    <div className="grid max-w-[min(520px,100%)] min-w-0 gap-0.5 rounded-md border border-line bg-white px-2.5 py-1.5 text-right shadow-sm">
      <p className="truncate text-xs font-semibold text-ink">{node.label}</p>
      <p className="truncate font-mono text-[11px] text-steel" title={node.id}>
        {displaySymbolID(node.id, modulePrefix)}
      </p>
      <p className="truncate text-[11px] text-steel" title={node.file || node.package}>
        {node.file || node.package}
      </p>
    </div>
  );
}
