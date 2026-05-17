"use client";

import type { SymbolInfo } from "@/types/graph";

interface SymbolSearchProps {
  symbols: SymbolInfo[];
  value: string;
  onChange: (value: string) => void;
  disabled?: boolean;
}

export function SymbolSearch({ symbols, value, onChange, disabled }: SymbolSearchProps) {
  return (
    <label className="grid gap-2 text-sm">
      <span className="font-semibold text-ink">Entry symbol</span>
      <input
        className="h-10 rounded-md border border-line bg-white px-3 font-mono text-sm text-ink outline-none transition focus:border-moss focus:ring-2 focus:ring-moss/20 disabled:bg-stone-100"
        list="symbol-options"
        value={value}
        onChange={(event) => onChange(event.target.value)}
        placeholder="main.main"
        disabled={disabled}
      />
      <datalist id="symbol-options">
        {symbols.map((symbol) => (
          <option key={symbol.id} value={symbol.id}>
            {symbol.label} · {symbol.package} · {symbol.file}
          </option>
        ))}
      </datalist>
    </label>
  );
}
