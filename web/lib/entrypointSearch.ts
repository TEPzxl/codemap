import type { Entrypoint } from "@/types/graph";

export interface EntrypointGroup {
  label: string;
  items: Entrypoint[];
}

const groupOrder = ["Main", "Startup", "Handlers", "Goroutines", "Exported"];

export function searchEntrypoints(entrypoints: Entrypoint[], query: string, limit = 8): EntrypointGroup[] {
  if (limit <= 0) {
    return [];
  }

  const normalizedQuery = query.trim().toLowerCase();
  const matches = normalizedQuery
    ? entrypoints.filter((entrypoint) => entrypointMatchesQuery(entrypoint, normalizedQuery)).sort(compareEntrypoints(normalizedQuery))
    : [...entrypoints].sort(compareEntrypoints(normalizedQuery));

  return groupEntrypoints(matches.slice(0, limit));
}

function entrypointMatchesQuery(entrypoint: Entrypoint, normalizedQuery: string): boolean {
  return searchableEntrypointFields(entrypoint).some((field) => field.toLowerCase().includes(normalizedQuery));
}

function compareEntrypoints(query: string): (left: Entrypoint, right: Entrypoint) => number {
  return (left, right) => {
    const leftGroup = groupOrder.indexOf(groupLabel(left));
    const rightGroup = groupOrder.indexOf(groupLabel(right));
    if (leftGroup !== rightGroup) {
      return leftGroup - rightGroup;
    }

    if (query) {
      const leftRank = matchRank(left, query);
      const rightRank = matchRank(right, query);
      if (leftRank !== rightRank) {
        return leftRank - rightRank;
      }
    }

    if (left.package !== right.package) {
      return left.package.localeCompare(right.package);
    }
    if (left.label !== right.label) {
      return left.label.localeCompare(right.label);
    }
    return left.id.localeCompare(right.id);
  };
}

function matchRank(entrypoint: Entrypoint, query: string): number {
  const fields = searchableEntrypointFields(entrypoint).map((field) => field.toLowerCase());
  if (fields.some((field) => field === query)) {
    return 0;
  }
  if (entrypoint.label.toLowerCase().startsWith(query)) {
    return 1;
  }
  if (entrypoint.id.toLowerCase().includes(query)) {
    return 2;
  }
  if (entrypoint.package.toLowerCase().includes(query)) {
    return 3;
  }
  if (entrypoint.file.toLowerCase().includes(query)) {
    return 4;
  }
  return 5;
}

function searchableEntrypointFields(entrypoint: Entrypoint): string[] {
  return [entrypoint.label, entrypoint.id, entrypoint.package, entrypoint.file, ...entrypoint.reasons].filter(Boolean);
}

function groupEntrypoints(entrypoints: Entrypoint[]): EntrypointGroup[] {
  const groups = new Map(groupOrder.map((label) => [label, [] as Entrypoint[]]));

  for (const entrypoint of entrypoints) {
    groups.get(groupLabel(entrypoint))?.push(entrypoint);
  }

  return groupOrder
    .map((label) => ({
      label,
      items: groups.get(label) ?? [],
    }))
    .filter((group) => group.items.length > 0);
}

function groupLabel(entrypoint: Entrypoint): string {
  if (entrypoint.reasons.includes("main-function")) {
    return "Main";
  }
  if (entrypoint.reasons.some((reason) => reason === "receiver:Handler" || reason === "name:Handler" || reason === "name:Handle")) {
    return "Handlers";
  }
  if (entrypoint.reasons.some((reason) => reason.startsWith("name:"))) {
    return "Startup";
  }
  if (entrypoint.reasons.includes("contains-goroutine")) {
    return "Goroutines";
  }
  return "Exported";
}
