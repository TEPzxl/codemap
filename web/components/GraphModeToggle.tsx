"use client";

export type GraphMode = "function" | "package";

interface GraphModeToggleProps {
  mode: GraphMode;
  onChange: (mode: GraphMode) => void;
}

export function GraphModeToggle({ mode, onChange }: GraphModeToggleProps) {
  return (
    <section className="grid gap-2 rounded-md border border-line bg-white p-3">
      <h2 className="text-sm font-semibold text-ink">Graph view</h2>
      <div className="grid grid-cols-2 gap-2">
        <ModeButton label="Function graph" active={mode === "function"} onClick={() => onChange("function")} />
        <ModeButton label="Package graph" active={mode === "package"} onClick={() => onChange("package")} />
      </div>
    </section>
  );
}

function ModeButton({ label, active, onClick }: { label: string; active: boolean; onClick: () => void }) {
  return (
    <button
      type="button"
      onClick={onClick}
      aria-pressed={active}
      className={
        active
          ? "h-9 rounded-md border border-ink bg-ink px-3 text-xs font-semibold text-white"
          : "h-9 rounded-md border border-line bg-white px-3 text-xs font-semibold text-ink transition hover:border-moss hover:text-moss"
      }
    >
      {label}
    </button>
  );
}
