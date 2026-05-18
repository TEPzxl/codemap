import type { Graph, SourceSnippet, SymbolsResponse, WarningsResponse } from "@/types/graph";

export interface GraphRequest {
  entry: string;
  depth: number;
  showExternal?: boolean;
  showUnresolved?: boolean;
  showInterface?: boolean;
  expandInterface?: boolean;
  packagePrefix?: string;
}

async function requestJSON<T>(path: string): Promise<T> {
  const response = await fetch(path, {
    headers: {
      Accept: "application/json",
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

export function graphURL(options: GraphRequest): string {
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
  return `/api/graph?${params.toString()}`;
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

export function fetchWarnings(): Promise<WarningsResponse> {
  return requestJSON<WarningsResponse>("/api/warnings");
}
