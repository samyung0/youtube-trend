-- Flag videos under 5 minutes (populated on fetch from YouTube duration).

BEGIN;

ALTER TABLE videos
  ADD COLUMN IF NOT EXISTS is_short_video BOOLEAN NOT NULL DEFAULT false;

COMMIT;
