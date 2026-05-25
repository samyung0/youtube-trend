"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import {
  getTopics,
  getTopicBubbleData,
  type Topic,
  type BubbleData,
  type TimeRange,
} from "@/lib/api";
import { TimeRangeSelector } from "@/components/time-range-selector";
import { ViewToggle, type ViewMode } from "@/components/view-toggle";
import { BubbleChart } from "@/components/bubble-chart";
import { Card } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import Link from "next/link";

export default function TopicsPage() {
  const router = useRouter();
  const [topics, setTopics] = useState<Topic[]>([]);
  const [bubbleData, setBubbleData] = useState<BubbleData[]>([]);
  const [loading, setLoading] = useState(true);
  const [range, setRange] = useState<TimeRange>("24h");
  const [viewMode, setViewMode] = useState<ViewMode>("list");

  useEffect(() => {
    setLoading(true);
    Promise.all([getTopics(range), getTopicBubbleData(range)])
      .then(([t, b]) => {
        setTopics(t);
        setBubbleData(b);
      })
      .catch(() => {
        setTopics([]);
        setBubbleData([]);
      })
      .finally(() => setLoading(false));
  }, [range]);

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Trending Topics</h1>
          <p className="mt-1 text-muted-foreground">
            AI-discovered micro-genres from trending videos
          </p>
        </div>
        <div className="flex items-center gap-3">
          <ViewToggle mode={viewMode} onChange={setViewMode} showBubble />
          <TimeRangeSelector value={range} onChange={setRange} />
        </div>
      </div>

      {loading ? (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 6 }).map((_, i) => (
            <Skeleton key={i} className="h-24 rounded-xl" />
          ))}
        </div>
      ) : viewMode === "bubble" ? (
        <BubbleChart
          data={bubbleData}
          height={500}
          onBubbleClick={(d) => router.push(d.href || `/topics`)}
        />
      ) : (
        <div
          className={
            viewMode === "grid"
              ? "grid gap-4 sm:grid-cols-2 lg:grid-cols-3"
              : "space-y-2"
          }
        >
          {topics.length === 0 ? (
            <div className="col-span-full py-20 text-center text-muted-foreground">
              No topics discovered yet. Run the scraper to generate them.
            </div>
          ) : (
            topics.map((topic) => (
              <Link key={topic.id} href={`/topics/${topic.slug}`}>
                <Card className="flex items-center gap-3 border-border/50 bg-card/50 p-4 backdrop-blur transition-colors hover:border-border">
                  <div
                    className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg text-sm font-bold text-white"
                    style={{ backgroundColor: topic.color }}
                  >
                    {topic.name.charAt(0)}
                  </div>
                  <div className="min-w-0 flex-1">
                    <h3 className="truncate font-medium">{topic.name}</h3>
                    <p className="truncate text-xs text-muted-foreground">
                      {topic.description}
                    </p>
                  </div>
                  <div className="shrink-0 text-right">
                    <p className="text-lg font-bold">{topic.video_count}</p>
                    <p className="text-xs text-muted-foreground">videos</p>
                  </div>
                  {topic.parent_category && (
                    <span className="shrink-0 rounded-full bg-accent px-2 py-0.5 text-xs text-muted-foreground">
                      {topic.parent_category}
                    </span>
                  )}
                </Card>
              </Link>
            ))
          )}
        </div>
      )}
    </div>
  );
}
