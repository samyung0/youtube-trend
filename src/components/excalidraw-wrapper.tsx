"use client";

import { useState, useEffect } from "react";

export default function ExcalidrawWrapper() {
  const [Excalidraw, setExcalidraw] = useState<React.ComponentType<Record<string, unknown>> | null>(null);

  useEffect(() => {
    import("@excalidraw/excalidraw").then((mod) => {
      setExcalidraw(() => mod.Excalidraw);
    });
  }, []);

  if (!Excalidraw) {
    return (
      <div className="flex h-[600px] items-center justify-center text-muted-foreground">
        Loading Excalidraw...
      </div>
    );
  }

  return (
    <div className="h-[600px]">
      <Excalidraw
        theme="dark"
        UIOptions={{
          canvasActions: {
            loadScene: false,
          },
        }}
      />
    </div>
  );
}
