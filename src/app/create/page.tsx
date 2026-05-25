"use client";

import dynamic from "next/dynamic";
import { ProGate } from "@/components/auth-provider";

const ExcalidrawWrapper = dynamic(() => import("@/components/excalidraw-wrapper"), {
  ssr: false,
  loading: () => (
    <div className="flex h-[600px] items-center justify-center text-muted-foreground">
      Loading canvas...
    </div>
  ),
});

export default function CreatePage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Create</h1>
        <p className="mt-1 text-muted-foreground">
          Design and plan your next thumbnail
        </p>
      </div>

      <ProGate>
        <div className="overflow-hidden rounded-xl border border-border/50">
          <ExcalidrawWrapper />
        </div>
      </ProGate>
    </div>
  );
}
