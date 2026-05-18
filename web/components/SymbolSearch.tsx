"use client";

import type { SymbolInfo } from "@/types/graph";
import { displayPackage, displaySymbolID } from "@/lib/displaySymbol";
import { useState } from "react";

interface SymbolSearchProps {
  symbols: SymbolInfo[];
  value: string;
  onChange: (value: string) => void;
  onSelectSymbol: (value: string) => void;
  packageFilter: string;
  onPackageFilterChange: (value: string) => void;
  modulePrefix: string;
  disabled?: boolean;
}

export function SymbolSearch({
  symbols,
  value,
  onChange,
  onSelectSymbol,
  packageFilter,
  onPackageFilterChange,
  modulePrefix,
  disabled,
}: SymbolSearchProps) {
  const [queryOverride, setQueryOverride] = useState<string | null>(null);
  const packages = Array.from(new Set(symbols.map((symbol) => symbol.package))).sort();
  const selectedDisplay = displaySymbolID(value, modulePrefix);
  const query = queryOverride ?? selectedDisplay;
  const normalizedQuery = query.trim().toLowerCase();
  const visibleSymbols = symbols
    .filter((symbol) => !packageFilter || symbol.package === packageFilter)
    .filter((symbol) => {
      if (!normalizedQuery) {
        return true;
      }
      return [
        symbol.label,
        symbol.id,
        displaySymbolID(symbol.id, modulePrefix),
        symbol.package,
        displayPackage(symbol.package, modulePrefix),
        symbol.file,
        symbol.receiver ?? "",
      ].some((field) => field.toLowerCase().includes(normalizedQuery));
    })
    .slice(0, 40);

  async function copySymbolID() {
    if (!value || !navigator.clipboard) {
      return;
    }
    await navigator.clipboard.writeText(value);
  }

  return (
    <section className="grid min-w-0 gap-3">
      <label className="grid gap-2 text-sm">
        <span className="font-semibold text-ink">Package</span>
        <select
          className="h-10 w-full min-w-0 rounded-md border border-line bg-white px-3 text-sm text-ink outline-none transition focus:border-moss focus:ring-2 focus:ring-moss/20 disabled:bg-stone-100"
          value={packageFilter}
          onChange={(event) => onPackageFilterChange(event.target.value)}
          disabled={disabled}
        >
          <option value="">All packages</option>
          {packages.map((pkg) => (
            <option key={pkg} value={pkg}>
              {displayPackage(pkg, modulePrefix)}
            </option>
          ))}
        </select>
      </label>

      <label className="grid gap-2 text-sm">
        <span className="font-semibold text-ink">Entry symbol</span>
        <div className="grid min-w-0 grid-cols-[minmax(0,1fr)_auto] gap-2">
          <input
            className="h-10 w-full min-w-0 rounded-md border border-line bg-white px-3 font-mono text-sm text-ink outline-none transition focus:border-moss focus:ring-2 focus:ring-moss/20 disabled:bg-stone-100"
            value={query}
            onChange={(event) => {
              setQueryOverride(event.target.value);
              onChange(event.target.value);
            }}
            placeholder="Search label, id, package, file, receiver"
            title={value}
            disabled={disabled}
          />
          <button
            type="button"
            onClick={copySymbolID}
            disabled={!value || disabled}
            className="h-10 w-14 shrink-0 rounded-md border border-line bg-white px-2 text-xs font-semibold text-ink transition hover:border-moss hover:text-moss disabled:cursor-not-allowed disabled:text-steel"
          >
            Copy
          </button>
        </div>
      </label>

      <div className="max-h-56 min-w-0 overflow-auto rounded-md border border-line bg-white">
        {disabled ? <p className="px-3 py-3 text-sm text-steel">Loading symbols</p> : null}
        {!disabled && visibleSymbols.length === 0 ? <p className="px-3 py-3 text-sm text-steel">No symbols found</p> : null}
        {!disabled && visibleSymbols.length > 0 ? (
          <ul className="divide-y divide-line">
            {visibleSymbols.map((symbol) => (
              <li key={symbol.id}>
                <button
                  type="button"
                  onClick={() => {
                    onChange(symbol.id);
                    onSelectSymbol(symbol.id);
                    setQueryOverride(displaySymbolID(symbol.id, modulePrefix));
                  }}
                  className="grid w-full min-w-0 gap-1 px-3 py-2 text-left transition hover:bg-paper"
                  title={symbol.id}
                >
                  <span className="text-sm font-semibold text-ink">{symbol.label}</span>
                  <span className="break-all font-mono text-xs leading-5 text-steel">
                    {displayPackage(symbol.package, modulePrefix)}
                  </span>
                  <span className="break-all text-xs leading-5 text-steel">{symbol.file}</span>
                </button>
              </li>
            ))}
          </ul>
        ) : null}
      </div>
    </section>
  );
}
