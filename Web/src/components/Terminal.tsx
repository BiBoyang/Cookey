import type { ReactNode } from "react";

export default function Terminal({
  title,
  children,
}: {
  title: string;
  children: ReactNode;
}) {
  return (
    <div className="mt-16 overflow-hidden rounded-xl border border-border text-left">
      <div className="flex items-center gap-2 border-b border-border bg-terminal-bar px-4 py-3">
        <span className="h-[10px] w-[10px] rounded-full bg-dot-red" />
        <span className="h-[10px] w-[10px] rounded-full bg-dot-yellow" />
        <span className="h-[10px] w-[10px] rounded-full bg-dot-green" />
        <span className="flex-1 text-center text-xs text-muted">{title}</span>
      </div>
      <div className="overflow-x-auto bg-terminal-bg p-[24px_20px] font-mono text-[13px] leading-[1.8]">
        {children}
      </div>
    </div>
  );
}
