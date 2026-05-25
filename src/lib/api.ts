const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:4450";

function getToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem("token");
}

export function setToken(token: string) {
  localStorage.setItem("token", token);
}

export function clearToken() {
  localStorage.removeItem("token");
}

async function fetchAPI<T>(path: string, init?: RequestInit): Promise<T> {
  const token = getToken();
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(init?.headers as Record<string, string>),
  };
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  const res = await fetch(`${API_BASE}${path}`, { ...init, headers });

  if (res.status === 401) {
    clearToken();
    if (typeof window !== "undefined") {
      window.dispatchEvent(new CustomEvent("auth:unauthorized"));
    }
    throw new Error("Unauthorized");
  }
  if (res.status === 403) {
    throw new Error("Pro subscription required");
  }
  if (!res.ok) {
    const body = await res.text();
    throw new Error(body || `API error: ${res.status}`);
  }

  return res.json();
}

export function getGoogleLoginURL() {
  return `${API_BASE}/api/auth/google`;
}

// Auth
export interface User {
  id: string;
  email: string;
  name: string;
  avatar_url: string;
}

export async function getMe(): Promise<{ user: User; is_pro: boolean }> {
  return fetchAPI("/api/auth/me");
}

// Videos
export interface Video {
  id: number;
  youtube_id: string;
  title: string;
  channel_name: string;
  channel_id: string;
  thumbnail_url: string;
  view_count: number;
  like_count: number;
  comment_count: number;
  category_id: number;
  tags: string[];
  published_at: string | null;
  duration: string;
  created_at: string;
}

export async function getTrendingVideos(
  range = "24h",
  category?: number,
  limit = 50,
): Promise<Video[]> {
  const params = new URLSearchParams({ range, limit: String(limit) });
  if (category) params.set("category", String(category));
  return fetchAPI(`/api/videos/trending?${params}`);
}

// Topics
export interface Topic {
  id: number;
  name: string;
  slug: string;
  description: string | null;
  color: string;
  parent_category: string | null;
  video_count: number;
  created_at: string;
}

export interface BubbleData {
  id: number;
  label: string;
  value: number;
  color: string;
  href: string;
  group: string;
}

export async function getTopics(range = "24h"): Promise<Topic[]> {
  return fetchAPI(`/api/topics?range=${range}`);
}

export async function getTopicBySlug(slug: string): Promise<{ topic: Topic; videos: Video[] }> {
  return fetchAPI(`/api/topics/${slug}`);
}

export async function getTopicBubbleData(range = "24h"): Promise<BubbleData[]> {
  return fetchAPI(`/api/topics/bubble?range=${range}`);
}

// Channels
export interface Channel {
  id: number;
  youtube_channel_id: string;
  name: string;
  avatar_url: string;
  subscriber_count: number;
  video_count: number;
  trending_count: number;
}

export interface ChannelSnapshot {
  subscriber_count: number;
  fetched_at: string;
}

export async function getTrendingChannels(range = "24h", limit = 50): Promise<Channel[]> {
  return fetchAPI(`/api/channels/trending?range=${range}&limit=${limit}`);
}

export async function getChannelHistory(id: number): Promise<ChannelSnapshot[]> {
  return fetchAPI(`/api/channels/${id}/history`);
}

export async function getChannelBubbleData(range = "24h"): Promise<BubbleData[]> {
  return fetchAPI(`/api/channels/bubble?range=${range}`);
}

// Analysis
export interface AnalysisStats {
  total_analyzed: number;
  face_percentage: number;
  avg_brightness: number;
  avg_face_count: number;
  color_frequency: Record<string, number>;
  ocr_words: Record<string, number>;
}

export async function getAnalysisStats(range = "24h"): Promise<AnalysisStats> {
  return fetchAPI(`/api/analysis/stats?range=${range}`);
}

export type TimeRange = "24h" | "7d" | "30d";
