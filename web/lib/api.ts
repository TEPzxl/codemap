import type {
  CallsiteSnippet,
  Graph,
  GraphDirection,
  ProjectMeta,
  RescanResponse,
  SourceSnippet,
  SymbolsResponse,
  WarningsResponse,
} from "@/types/graph";

export interface GraphRequest {
  entry: string;
  depth: number;
  direction?: GraphDirection;
  showExternal?: boolean;
  showUnresolved?: boolean;
  showInterface?: boolean;
  expandInterface?: boolean;
  packagePrefix?: string;
}

async function requestJSON<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(path, {
    ...init,
    headers: {
      Accept: "application/json",
      ...init?.headers,
    },
  });

  if (!response.ok) {
    let message = `${response.status} ${response.statusText}`;
    try {
      const body = (await response.json()) as { error?: string };
      if (body.error) {
        message = body.error;
      }
    } catch {
      // Keep the HTTP status fallback when the response is not JSON.
    }
    throw new Error(message);
  }

  return (await response.json()) as T;
}

export function fetchSymbols(): Promise<SymbolsResponse> {
  return requestJSON<SymbolsResponse>("/api/symbols");
}

export function fetchMeta(): Promise<ProjectMeta> {
  return requestJSON<ProjectMeta>("/api/meta");
}

export function rescanProject(): Promise<RescanResponse> {
  return requestJSON<RescanResponse>("/api/rescan", { method: "POST" });
}

export function graphURL(options: GraphRequest): string {
  return `/api/graph?${graphParams(options).toString()}`;
}

function graphParams(options: GraphRequest): URLSearchParams {
  const params = new URLSearchParams({
    entry: options.entry,
    depth: String(options.depth),
    direction: options.direction ?? "downstream",
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
  return params;
}

export function fetchGraph(options: GraphRequest): Promise<Graph> {
  return requestJSON<Graph>(graphURL(options));
}

export function fetchSource(nodeId: string): Promise<SourceSnippet> {
  const params = new URLSearchParams({
    node_id: nodeId,
  });
  return requestJSON<SourceSnippet>(`/api/source?${params.toString()}`);
}

export function fetchCallsite(edgeId: string, options: GraphRequest): Promise<CallsiteSnippet> {
  const params = graphParams(options);
  params.set("edge_id", edgeId);
  return requestJSON<CallsiteSnippet>(`/api/callsite?${params.toString()}`);
}

export function fetchWarnings(): Promise<WarningsResponse> {
  return requestJSON<WarningsResponse>("/api/warnings");
}
