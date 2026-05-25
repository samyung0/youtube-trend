"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import {
  getTrendingChannels,
  getChannelBubbleData,
  type Channel,
  type BubbleData,
  type TimeRange,
} from "@/lib/api";
import { TimeRangeSelector } from "@/components/time-range-selector";
import { ViewToggle, type ViewMode } from "@/components/view-toggle";
import { BubbleChart } from "@/components/bubble-chart";
import { Card } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Users, TrendingUp } from "lucide-react";

function formatCount(n: number): string {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`;
  if (n >= 1_000) return `${(n / 1_000).toFixed(1)}K`;
  return String(n);
}

export default function ChannelsPage() {
  const router = useRouter();
  const [channels, setChannels] = useState<Channel[]>([]);
  const [bubbleData, setBubbleData] = useState<BubbleData[]>([]);
  const [loading, setLoading] = useState(true);
  const [range, setRange] = useState<TimeRange>("24h");
  const [viewMode, setViewMode] = useState<ViewMode>("list");

  useEffect(() => {
    setLoading(true);
    Promise.all([getTrendingChannels(range), getChannelBubbleData(range)])
      .then(([c, b]) => {
        setChannels(c);
        setBubbleData(b);
      })
      .catch(() => {
        setChannels([]);
        setBubbleData([]);
      })
      .finally(() => setLoading(false));
  }, [range]);

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">
            Trending Channels
          </h1>
          <p className="mt-1 text-muted-foreground">
            Channels appearing most in trending videos
          </p>
        </div>
        <div className="flex items-center gap-3">
          <ViewToggle mode={viewMode} onChange={setViewMode} showBubble />
          <TimeRangeSelector value={range} onChange={setRange} />
        </div>
      </div>

      {loading ? (
        <div className="space-y-2">
          {Array.from({ length: 6 }).map((_, i) => (
            <Skeleton key={i} className="h-20 rounded-xl" />
          ))}
        </div>
      ) : viewMode === "bubble" ? (
        <BubbleChart
          data={bubbleData}
          height={500}
          onBubbleClick={(d) => router.push(`/channels/${d.id}`)}
        />
      ) : (
        <div
          className={
            viewMode === "grid"
              ? "grid gap-4 sm:grid-cols-2 lg:grid-cols-3"
              : "space-y-2"
          }
        >
          {channels.length === 0 ? (
            <div className="col-span-full py-20 text-center text-muted-foreground">
              No channel data yet. Run the scraper first.
            </div>
          ) : (
            channels.map((ch) => (
              <Card
                key={ch.id}
                className="flex items-center gap-4 border-border/50 bg-card/50 p-4 backdrop-blur"
              >
                {ch.avatar_url ? (
                  <img
                    src={ch.avatar_url}
                    alt={ch.name}
                    className="h-10 w-10 shrink-0 rounded-full"
                  />
                ) : (
                  <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-accent text-sm font-bold">
                    {ch.name.charAt(0)}
                  </div>
                )}
                <div className="min-w-0 flex-1">
                  <h3 className="truncate font-medium">{ch.name}</h3>
                  <div className="flex items-center gap-3 text-xs text-muted-foreground">
                    <span className="flex items-center gap-1">
                      <Users className="h-3 w-3" />
                      {formatCount(ch.subscriber_count)} subs
                    </span>
                    <span className="flex items-center gap-1">
                      <TrendingUp className="h-3 w-3" />
                      {ch.trending_count} trending
                    </span>
                  </div>
                </div>
              </Card>
            ))
          )}
        </div>
      )}
    </div>
  );
}
