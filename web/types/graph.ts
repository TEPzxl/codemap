export type NodeKind = "function" | "method" | "external" | "unresolved";

export type EdgeResolution = "resolved" | "interface" | "external" | "unresolved";

export type GraphDirection = "downstream" | "upstream" | "both";

export type GraphExportFormat = "json" | "mermaid" | "dot";

export interface Graph {
  entry: string;
  nodes: Node[];
  edges: Edge[];
  warnings: Warning[];
}

export interface PackageGraph {
  nodes: PackageNode[];
  edges: PackageEdge[];
  warnings: Warning[];
}

export interface PackageNode {
  id: string;
  package: string;
  full_package?: string;
  symbols: number;
  calls: number;
}

export interface PackageEdge {
  id: string;
  from: string;
  to: string;
  calls: number;
}

export interface Entrypoint {
  id: string;
  label: string;
  package: string;
  file: string;
  kind: "function" | "method";
  reasons: string[];
}

export interface EntrypointsResponse {
  entrypoints: Entrypoint[];
  warnings: Warning[];
  note: string;
}

export interface PathResult {
  from: string;
  to: string;
  paths: SymbolPath[];
  graph: Graph;
  warnings: Warning[];
}

export interface SymbolPath {
  nodes: string[];
  edges: string[];
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
  candidate?: boolean;
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

export interface ProjectMeta {
  root: string;
  module: string;
  packages: number;
  symbols: number;
  calls: number;
  warnings: number;
  analyzed_at: string;
  analysis_duration_ms: number;
  version: string;
}

export interface RescanResponse {
  meta: ProjectMeta;
}

export interface SourceSnippet {
  node_id: string;
  file: string;
  start_line: number;
  end_line: number;
  source: string;
  language: string;
}

export interface CallsiteSnippet {
  edge_id: string;
  file: string;
  line: number;
  column: number;
  start_line: number;
  end_line: number;
  source: string;
  highlight_line: number;
  language: string;
}

export type SourceView =
  | {
      mode: "node";
      data: SourceSnippet;
    }
  | {
      mode: "callsite";
      data: CallsiteSnippet;
    };
