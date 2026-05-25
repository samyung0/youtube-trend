"use client";

import { cn } from "@/lib/utils";
import { LayoutGrid, List, CircleDot } from "lucide-react";

export type ViewMode = "grid" | "list" | "bubble";

export function ViewToggle({
  mode,
  onChange,
  showBubble = false,
}: {
  mode: ViewMode;
  onChange: (mode: ViewMode) => void;
  showBubble?: boolean;
}) {
  const options: { value: ViewMode; icon: typeof LayoutGrid; label: string }[] = [
    { value: "grid", icon: LayoutGrid, label: "Grid" },
    { value: "list", icon: List, label: "List" },
  ];
  if (showBubble) {
    options.push({ value: "bubble", icon: CircleDot, label: "Bubble" });
  }

  return (
    <div className="inline-flex rounded-lg bg-muted p-1">
      {options.map(({ value, icon: Icon, label }) => (
        <button
          key={value}
          onClick={() => onChange(value)}
          className={cn(
            "flex items-center gap-1.5 rounded-md px-2.5 py-1 text-sm transition-colors",
            mode === value
              ? "bg-background text-foreground shadow-sm"
              : "text-muted-foreground hover:text-foreground"
          )}
          title={label}
        >
          <Icon className="h-3.5 w-3.5" />
          <span className="hidden sm:inline">{label}</span>
        </button>
      ))}
    </div>
  );
}
