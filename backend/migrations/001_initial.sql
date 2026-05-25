-- 001_initial.sql: Core ThumbTrend schema

BEGIN;

-- Trigger function for auto-updating updated_at columns
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- videos: deduplicated by youtube_id
CREATE TABLE IF NOT EXISTS videos (
  id BIGSERIAL PRIMARY KEY,
  youtube_id TEXT NOT NULL UNIQUE,
  title TEXT NOT NULL,
  channel_name TEXT NOT NULL,
  channel_id TEXT NOT NULL,
  channel_db_id BIGINT,
  thumbnail_url TEXT NOT NULL,
  view_count BIGINT NOT NULL DEFAULT 0,
  like_count BIGINT NOT NULL DEFAULT 0,
  comment_count BIGINT NOT NULL DEFAULT 0,
  category_id INT DEFAULT 0,
  tags TEXT[] DEFAULT '{}',
  published_at TIMESTAMPTZ,
  duration TEXT DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- trending_snapshots
CREATE TABLE IF NOT EXISTS trending_snapshots (
  id BIGSERIAL PRIMARY KEY,
  fetched_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  region TEXT NOT NULL DEFAULT 'US',
  category_id INT,
  video_count INT NOT NULL DEFAULT 0
);

-- snapshot_videos junction
CREATE TABLE IF NOT EXISTS snapshot_videos (
  id BIGSERIAL PRIMARY KEY,
  snapshot_id BIGINT NOT NULL REFERENCES trending_snapshots(id) ON DELETE CASCADE,
  video_id BIGINT NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
  rank INT NOT NULL,
  UNIQUE(snapshot_id, video_id)
);

-- channels: tracked YouTube channels
CREATE TABLE IF NOT EXISTS channels (
  id BIGSERIAL PRIMARY KEY,
  youtube_channel_id TEXT NOT NULL UNIQUE,
  name TEXT NOT NULL,
  avatar_url TEXT DEFAULT '',
  subscriber_count BIGINT NOT NULL DEFAULT 0,
  video_count INT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- channel_snapshots: daily subscriber tracking
CREATE TABLE IF NOT EXISTS channel_snapshots (
  id BIGSERIAL PRIMARY KEY,
  channel_id BIGINT NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
  subscriber_count BIGINT NOT NULL DEFAULT 0,
  fetched_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- topics (renamed from micro_genres)
CREATE TABLE IF NOT EXISTS topics (
  id BIGSERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  slug TEXT NOT NULL UNIQUE,
  description TEXT,
  color TEXT DEFAULT '#6366f1',
  parent_category TEXT,
  snapshot_date DATE NOT NULL DEFAULT CURRENT_DATE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- topic_videos junction
CREATE TABLE IF NOT EXISTS topic_videos (
  id BIGSERIAL PRIMARY KEY,
  video_id BIGINT NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
  topic_id BIGINT NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
  UNIQUE(video_id, topic_id)
);

-- thumbnail_analyses
CREATE TABLE IF NOT EXISTS thumbnail_analyses (
  id BIGSERIAL PRIMARY KEY,
  video_id BIGINT NOT NULL REFERENCES videos(id) ON DELETE CASCADE UNIQUE,
  dominant_colors JSONB DEFAULT '[]',
  has_face BOOLEAN DEFAULT false,
  face_count INT DEFAULT 0,
  ocr_text TEXT DEFAULT '',
  brightness FLOAT DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- FK from videos to channels (deferred so channels can be backfilled)
ALTER TABLE videos
  ADD CONSTRAINT fk_videos_channel
  FOREIGN KEY (channel_db_id) REFERENCES channels(id) ON DELETE SET NULL;

-- Indexes
CREATE INDEX idx_videos_youtube_id ON videos(youtube_id);
CREATE INDEX idx_videos_channel_id ON videos(channel_id);
CREATE INDEX idx_videos_channel_db_id ON videos(channel_db_id);
CREATE INDEX idx_videos_published_at ON videos(published_at);
CREATE INDEX idx_videos_category_id ON videos(category_id);

CREATE INDEX idx_trending_snapshots_fetched_at ON trending_snapshots(fetched_at);
CREATE INDEX idx_trending_snapshots_region ON trending_snapshots(region);

CREATE INDEX idx_snapshot_videos_snapshot_id ON snapshot_videos(snapshot_id);
CREATE INDEX idx_snapshot_videos_video_id ON snapshot_videos(video_id);

CREATE INDEX idx_channels_youtube_channel_id ON channels(youtube_channel_id);

CREATE INDEX idx_channel_snapshots_channel_id ON channel_snapshots(channel_id);
CREATE INDEX idx_channel_snapshots_fetched_at ON channel_snapshots(fetched_at);

CREATE INDEX idx_topics_slug ON topics(slug);
CREATE INDEX idx_topics_snapshot_date ON topics(snapshot_date);

CREATE INDEX idx_topic_videos_video_id ON topic_videos(video_id);
CREATE INDEX idx_topic_videos_topic_id ON topic_videos(topic_id);

CREATE INDEX idx_thumbnail_analyses_video_id ON thumbnail_analyses(video_id);

-- Triggers
CREATE TRIGGER trg_videos_updated_at
  BEFORE UPDATE ON videos
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_channels_updated_at
  BEFORE UPDATE ON channels
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

COMMIT;
