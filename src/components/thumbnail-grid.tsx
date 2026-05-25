"use client";

import Image from "next/image";
import { Card } from "@/components/ui/card";
import type { Video } from "@/lib/api";
import { Eye, ThumbsUp, MessageSquare } from "lucide-react";
import type { ViewMode } from "@/components/view-toggle";

function formatCount(n: number): string {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`;
  if (n >= 1_000) return `${(n / 1_000).toFixed(1)}K`;
  return String(n);
}

function timeAgo(dateStr: string | null): string {
  if (!dateStr) return "";
  const diff = Date.now() - new Date(dateStr).getTime();
  const hours = Math.floor(diff / 3600000);
  if (hours < 1) return "Just now";
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  if (days < 7) return `${days}d ago`;
  return `${Math.floor(days / 7)}w ago`;
}

export function ThumbnailGrid({ videos, mode = "grid" }: { videos: Video[]; mode?: ViewMode }) {
  if (!videos || videos.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-20 text-muted-foreground">
        <p className="text-lg font-medium">No videos found</p>
        <p className="text-sm">Try adjusting the time range or check back later.</p>
      </div>
    );
  }

  if (mode === "list") {
    return (
      <div className="space-y-2">
        {videos.map((video, i) => (
          <Card
            key={video.id}
            className="flex items-center gap-4 border-border/50 bg-card/50 p-3 backdrop-blur"
          >
            <div className="relative h-16 w-28 shrink-0 overflow-hidden rounded-md">
              <Image
                src={video.thumbnail_url}
                alt={video.title}
                fill
                className="object-cover"
                sizes="112px"
              />
              {i < 3 && (
                <div className="absolute left-1 top-1 rounded bg-red-600 px-1 py-0.5 text-[10px] font-bold text-white">
                  #{i + 1}
                </div>
              )}
            </div>
            <div className="min-w-0 flex-1">
              <h3 className="truncate text-sm font-medium">{video.title}</h3>
              <p className="truncate text-xs text-muted-foreground">{video.channel_name}</p>
            </div>
            <div className="flex shrink-0 items-center gap-4 text-xs text-muted-foreground">
              <span className="flex items-center gap-1">
                <Eye className="h-3 w-3" />
                {formatCount(video.view_count)}
              </span>
              <span className="flex items-center gap-1">
                <ThumbsUp className="h-3 w-3" />
                {formatCount(video.like_count)}
              </span>
              <span className="hidden items-center gap-1 sm:flex">
                <MessageSquare className="h-3 w-3" />
                {formatCount(video.comment_count)}
              </span>
              <span className="w-14 text-right">{timeAgo(video.published_at)}</span>
            </div>
          </Card>
        ))}
      </div>
    );
  }

  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
      {videos.map((video, i) => (
        <Card
          key={video.id}
          className="group overflow-hidden border-border/50 bg-card/50 backdrop-blur transition-all hover:border-border hover:shadow-lg hover:shadow-black/20"
        >
          <div className="relative aspect-video overflow-hidden">
            <Image
              src={video.thumbnail_url}
              alt={video.title}
              fill
              className="object-cover transition-transform duration-300 group-hover:scale-105"
              sizes="(max-width: 640px) 100vw, (max-width: 1024px) 50vw, 25vw"
            />
            <div className="absolute inset-0 bg-gradient-to-t from-black/80 via-transparent to-transparent opacity-0 transition-opacity group-hover:opacity-100" />
            <div className="absolute bottom-2 left-2 right-2 flex items-center gap-3 text-xs text-white opacity-0 transition-opacity group-hover:opacity-100">
              <span className="flex items-center gap-1">
                <Eye className="h-3 w-3" />
                {formatCount(video.view_count)}
              </span>
              <span className="flex items-center gap-1">
                <ThumbsUp className="h-3 w-3" />
                {formatCount(video.like_count)}
              </span>
              <span className="flex items-center gap-1">
                <MessageSquare className="h-3 w-3" />
                {formatCount(video.comment_count)}
              </span>
            </div>
            {i < 3 && (
              <div className="absolute left-2 top-2 rounded bg-red-600 px-1.5 py-0.5 text-xs font-bold text-white">
                #{i + 1}
              </div>
            )}
          </div>
          <div className="p-3">
            <h3 className="line-clamp-2 text-sm font-medium leading-snug">{video.title}</h3>
            <div className="mt-1.5 flex items-center justify-between text-xs text-muted-foreground">
              <span className="truncate">{video.channel_name}</span>
              <span className="shrink-0">{timeAgo(video.published_at)}</span>
            </div>
          </div>
        </Card>
      ))}
    </div>
  );
}
