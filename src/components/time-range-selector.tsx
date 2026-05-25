"use client";

import { cn } from "@/lib/utils";
import type { TimeRange } from "@/lib/api";

const ranges: { value: TimeRange; label: string }[] = [
  { value: "24h", label: "24h" },
  { value: "7d", label: "7d" },
  { value: "30d", label: "30d" },
];

export function TimeRangeSelector({
  value,
  onChange,
}: {
  value: TimeRange;
  onChange: (range: TimeRange) => void;
}) {
  return (
    <div className="inline-flex rounded-lg bg-muted p-1">
      {ranges.map(({ value: v, label }) => (
        <button
          key={v}
          onClick={() => onChange(v)}
          className={cn(
            "rounded-md px-3 py-1.5 text-sm font-medium transition-colors",
            value === v
              ? "bg-background text-foreground shadow-sm"
              : "text-muted-foreground hover:text-foreground"
          )}
        >
          {label}
        </button>
      ))}
    </div>
  );
}
