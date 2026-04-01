import type { ReactNode } from "react";

export default function Badge({ children }: { children: ReactNode }) {
  return (
    <div className="inline-flex items-center gap-2 rounded-full border border-border bg-tag-bg px-[14px] py-[6px] text-[13px] text-muted">
      <span className="inline-block h-[6px] w-[6px] rounded-full bg-accent" />
      {children}
    </div>
  );
}
