export type NodeKind = "function" | "method" | "external" | "unresolved";

export type EdgeResolution = "resolved" | "interface" | "external" | "unresolved";

export interface Graph {
  entry: string;
  nodes: Node[];
  edges: Edge[];
  warnings: Warning[];
}

export interface Node {
  id: string;
  label: string;
  kind: NodeKind;
  package: string;
  receiver?: string;
  file: string;
  start_line: number;
  end_line: number;
  is_external: boolean;
}

export interface Edge {
  id: string;
  from: string;
  to: string;
  kind: string;
  resolution: EdgeResolution;
  callsite: Callsite;
}

export interface Callsite {
  file: string;
  line: number;
  column: number;
}

export interface Warning {
  code: string;
  message: string;
  file?: string;
}

export interface SymbolInfo {
  id: string;
  name: string;
  label: string;
  kind: "function" | "method";
  package: string;
  receiver?: string;
  file: string;
  start_line: number;
  end_line: number;
}

export interface SymbolsResponse {
  symbols: SymbolInfo[];
  warnings: Warning[];
}

export interface WarningsResponse {
  warnings: Warning[];
}

export interface SourceSnippet {
  node_id: string;
  file: string;
  start_line: number;
  end_line: number;
  source: string;
  language: string;
}
