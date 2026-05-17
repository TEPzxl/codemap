import type { Graph, SourceSnippet, SymbolsResponse, WarningsResponse } from "@/types/graph";

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

export function fetchGraph(entry: string, depth: number): Promise<Graph> {
  const params = new URLSearchParams({
    entry,
    depth: String(depth),
  });
  return requestJSON<Graph>(`/api/graph?${params.toString()}`);
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
