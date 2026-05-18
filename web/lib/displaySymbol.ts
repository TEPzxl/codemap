import type { SymbolInfo } from "@/types/graph";

export function inferModulePrefix(symbols: SymbolInfo[]): string {
  const packages = symbols.map((symbol) => symbol.package).filter(Boolean);
  if (packages.length === 0) {
    return "";
  }

  const first = packages[0].split("/");
  let length = first.length;
  for (const pkg of packages.slice(1)) {
    const parts = pkg.split("/");
    length = Math.min(length, parts.length);
    for (let index = 0; index < length; index += 1) {
      if (first[index] !== parts[index]) {
        length = index;
        break;
      }
    }
  }
  return first.slice(0, length).join("/");
}

export function displayPackage(pkg: string, modulePrefix: string): string {
  if (!modulePrefix) {
    return pkg;
  }
  if (pkg === modulePrefix) {
    return ".";
  }
  if (pkg.startsWith(`${modulePrefix}/`)) {
    return pkg.slice(modulePrefix.length + 1);
  }
  return pkg;
}

export function displaySymbolID(id: string, modulePrefix: string): string {
  if (!modulePrefix) {
    return id;
  }
  if (id.startsWith(`${modulePrefix}/`)) {
    return id.slice(modulePrefix.length + 1);
  }
  if (id.startsWith(`${modulePrefix}.`)) {
    return id.slice(modulePrefix.length + 1);
  }
  return id;
}

