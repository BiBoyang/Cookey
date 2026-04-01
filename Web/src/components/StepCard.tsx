import type { ReactNode } from "react";

export default function StepCard({
  number,
  title,
  children,
  position,
}: {
  number: string;
  title: string;
  children: ReactNode;
  position: "first" | "last";
}) {
  const radiusClass =
    position === "first"
      ? "xs:rounded-l-xl xs:rounded-r-none max-xs:rounded-t-[10px] max-xs:rounded-b-none"
      : "xs:rounded-r-xl xs:rounded-l-none max-xs:rounded-b-[10px] max-xs:rounded-t-none";

  return (
    <div
      className={`border border-border bg-surface p-[28px_24px] ${radiusClass}`}
    >
      <p className="mb-4 text-[11px] font-semibold uppercase tracking-[0.1em] text-muted">
        {number}
      </p>
      <h3 className="mb-2 text-[15px] font-semibold tracking-[-0.01em]">
        {title}
      </h3>
      <p className="text-[13.5px] leading-[1.6] text-muted">{children}</p>
    </div>
  );
}
