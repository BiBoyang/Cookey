import type { ReactNode } from "react";

export default function PropertyCard({
  icon,
  title,
  children,
}: {
  icon: string;
  title: string;
  children: ReactNode;
}) {
  return (
    <div className="rounded-[10px] border border-border bg-surface p-6">
      <div className="mb-3 text-[22px]">{icon}</div>
      <h3 className="mb-[6px] text-sm font-semibold">{title}</h3>
      <p className="text-[13px] leading-[1.55] text-muted">{children}</p>
    </div>
  );
}
