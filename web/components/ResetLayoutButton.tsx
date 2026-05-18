"use client";

interface ResetLayoutButtonProps {
  onClick: () => void;
}

export function ResetLayoutButton({ onClick }: ResetLayoutButtonProps) {
  return (
    <button
      type="button"
      aria-label="Reset layout"
      title="Reset layout"
      onClick={onClick}
      className="grid h-10 w-10 place-items-center rounded-md border border-line bg-white text-ink shadow-sm transition hover:border-moss hover:text-moss"
    >
      <svg aria-hidden="true" viewBox="0 0 24 24" className="h-4 w-4" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M21 12a9 9 0 1 1-2.64-6.36" />
        <path d="M21 3v6h-6" />
      </svg>
    </button>
  );
}
