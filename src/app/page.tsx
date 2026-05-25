"use client";

import { useEffect, useState } from "react";
import { getTrendingVideos, type Video, type TimeRange } from "@/lib/api";
import { ThumbnailGrid } from "@/components/thumbnail-grid";
import { TimeRangeSelector } from "@/components/time-range-selector";
import { ViewToggle, type ViewMode } from "@/components/view-toggle";
import { Skeleton } from "@/components/ui/skeleton";
import { YOUTUBE_CATEGORIES } from "@/lib/types";

export default function DashboardPage() {
  const [videos, setVideos] = useState<Video[]>([]);
  const [loading, setLoading] = useState(true);
  const [range, setRange] = useState<TimeRange>("24h");
  const [viewMode, setViewMode] = useState<ViewMode>("grid");
  const [category, setCategory] = useState<number | undefined>();

  useEffect(() => {
    setLoading(true);
    getTrendingVideos(range, category, 50)
      .then(setVideos)
      .catch(() => setVideos([]))
      .finally(() => setLoading(false));
  }, [range, category]);

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">
            What&apos;s Trending
          </h1>
          <p className="mt-1 text-muted-foreground">
            YouTube thumbnail patterns and trending insights
          </p>
        </div>
        <div className="flex items-center gap-3">
          <ViewToggle mode={viewMode} onChange={setViewMode} />
          <TimeRangeSelector value={range} onChange={setRange} />
        </div>
      </div>

      <div className="flex flex-wrap gap-2">
        <button
          onClick={() => setCategory(undefined)}
          className={`rounded-md px-3 py-1.5 text-sm transition-colors ${
            category === undefined
              ? "bg-accent font-medium text-accent-foreground"
              : "text-muted-foreground hover:bg-accent/50"
          }`}
        >
          All
        </button>
        {Object.entries(YOUTUBE_CATEGORIES)
          .filter(([id]) => id !== "0")
          .map(([id, name]) => (
            <button
              key={id}
              onClick={() => setCategory(Number(id))}
              className={`rounded-md px-3 py-1.5 text-sm transition-colors ${
                category === Number(id)
                  ? "bg-accent font-medium text-accent-foreground"
                  : "text-muted-foreground hover:bg-accent/50"
              }`}
            >
              {name}
            </button>
          ))}
      </div>

      {loading ? (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
          {Array.from({ length: 8 }).map((_, i) => (
            <Skeleton key={i} className="aspect-video rounded-xl" />
          ))}
        </div>
      ) : (
        <ThumbnailGrid videos={videos} mode={viewMode} />
      )}
    </div>
  );
}
