"use client";

import { useEffect, useState, use } from "react";
import Link from "next/link";
import { getTopicBySlug, type Topic, type Video } from "@/lib/api";
import { ThumbnailGrid } from "@/components/thumbnail-grid";
import { Skeleton } from "@/components/ui/skeleton";
import { ArrowLeft } from "lucide-react";

export default function TopicDetailPage({
  params,
}: {
  params: Promise<{ slug: string }>;
}) {
  const { slug } = use(params);
  const [topic, setTopic] = useState<Topic | null>(null);
  const [videos, setVideos] = useState<Video[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(true);
    getTopicBySlug(slug)
      .then((data) => {
        setTopic(data.topic);
        setVideos(data.videos);
      })
      .catch(() => {})
      .finally(() => setLoading(false));
  }, [slug]);

  if (loading) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-32 rounded-xl" />
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
          {Array.from({ length: 8 }).map((_, i) => (
            <Skeleton key={i} className="aspect-video rounded-xl" />
          ))}
        </div>
      </div>
    );
  }

  if (!topic) {
    return (
      <div className="py-20 text-center text-muted-foreground">
        Topic not found.
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <Link
        href="/topics"
        className="inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground"
      >
        <ArrowLeft className="h-4 w-4" />
        Back to Topics
      </Link>

      <div className="rounded-xl border border-border/50 bg-card/50 p-6 backdrop-blur">
        <div className="flex items-start gap-4">
          <div
            className="flex h-12 w-12 shrink-0 items-center justify-center rounded-xl text-xl font-bold text-white"
            style={{ backgroundColor: topic.color }}
          >
            {topic.name.charAt(0)}
          </div>
          <div>
            <h1 className="text-2xl font-bold">{topic.name}</h1>
            {topic.description && (
              <p className="mt-1 text-muted-foreground">{topic.description}</p>
            )}
            <p className="mt-2 text-sm text-muted-foreground">
              {videos.length} trending videos
              {topic.parent_category && (
                <span className="ml-2 rounded-full bg-accent px-2 py-0.5 text-xs">
                  {topic.parent_category}
                </span>
              )}
            </p>
          </div>
        </div>
      </div>

      <ThumbnailGrid videos={videos} />
    </div>
  );
}
