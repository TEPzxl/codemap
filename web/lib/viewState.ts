import type { GraphRequest } from "@/lib/api";
import type { GraphDirection } from "@/types/graph";

export interface ViewState {
  entry?: string;
  depth?: number;
  direction?: GraphDirection;
  packageFilter?: string;
  showExternal: boolean;
  showUnresolved: boolean;
  showInterface: boolean;
  expandInterface: boolean;
}

export function parseViewState(search: string): ViewState {
  const params = new URLSearchParams(search);
  const rawDepth = params.get("depth");
  const depth = rawDepth === null ? undefined : Number(rawDepth);

  return {
    entry: valueOrUndefined(params.get("entry")),
    depth: depth !== undefined && Number.isInteger(depth) && depth >= 0 ? depth : undefined,
    direction: parseDirection(params.get("direction")),
    packageFilter: valueOrUndefined(params.get("package")),
    showExternal: isTrue(params, "show_external", "showExternal"),
    showUnresolved: isTrue(params, "show_unresolved", "showUnresolved"),
    showInterface: isTrue(params, "show_interface", "showInterface"),
    expandInterface: isTrue(params, "expand_interface", "expandInterface"),
  };
}

export function viewStateFromGraphRequest(options: GraphRequest): Required<ViewState> {
  return {
    entry: options.entry,
    depth: options.depth,
    direction: options.direction ?? "downstream",
    packageFilter: options.packagePrefix ?? "",
    showExternal: options.showExternal ?? false,
    showUnresolved: options.showUnresolved ?? false,
    showInterface: options.showInterface ?? false,
    expandInterface: options.expandInterface ?? false,
  };
}

export function serializeViewState(options: GraphRequest): URLSearchParams {
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

function isTrue(params: URLSearchParams, canonical: string, legacy: string): boolean {
  return params.get(canonical) === "true" || params.get(legacy) === "true";
}

function valueOrUndefined(value: string | null): string | undefined {
  return value && value.trim() ? value : undefined;
}

function parseDirection(value: string | null): GraphDirection | undefined {
  if (value === "downstream" || value === "upstream" || value === "both") {
    return value;
  }
  return undefined;
}
