package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thumbtrend/backend/internal/model"
)

type TopicStore struct {
	pool *pgxpool.Pool
}

func NewTopicStore(pool *pgxpool.Pool) *TopicStore {
	return &TopicStore{pool: pool}
}

func (s *TopicStore) GetTopics(ctx context.Context, since time.Time) ([]model.Topic, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT t.id, t.name, t.slug, t.description, t.color,
		       t.parent_category, t.snapshot_date::text, t.created_at,
		       COUNT(tv.video_id) AS video_count
		FROM topics t
		LEFT JOIN topic_videos tv ON tv.topic_id = t.id
		LEFT JOIN videos v ON v.id = tv.video_id AND v.created_at >= $1
		GROUP BY t.id
		ORDER BY video_count DESC`, since,
	)
	if err != nil {
		return nil, fmt.Errorf("get topics: %w", err)
	}
	defer rows.Close()

	out := make([]model.Topic, 0)
	for rows.Next() {
		var t model.Topic
		if err := rows.Scan(
			&t.ID, &t.Name, &t.Slug, &t.Description, &t.Color,
			&t.ParentCategory, &t.SnapshotDate, &t.CreatedAt, &t.VideoCount,
		); err != nil {
			return nil, fmt.Errorf("scan topic: %w", err)
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *TopicStore) GetBySlug(ctx context.Context, slug string) (*model.Topic, error) {
	var t model.Topic
	err := s.pool.QueryRow(ctx, `
		SELECT id, name, slug, description, color, parent_category, snapshot_date::text, created_at
		FROM topics WHERE slug = $1`, slug,
	).Scan(&t.ID, &t.Name, &t.Slug, &t.Description, &t.Color,
		&t.ParentCategory, &t.SnapshotDate, &t.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get topic by slug: %w", err)
	}
	return &t, nil
}

func (s *TopicStore) GetVideosByTopic(ctx context.Context, topicID int64, limit int) ([]model.Video, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT v.id, v.youtube_id, v.title, v.channel_name, v.channel_id, v.channel_db_id,
		       v.thumbnail_url, v.view_count, v.like_count, v.comment_count,
		       v.category_id, v.tags, v.published_at, v.duration, v.created_at, v.updated_at
		FROM videos v
		JOIN topic_videos tv ON tv.video_id = v.id
		WHERE tv.topic_id = $1
		ORDER BY v.view_count DESC
		LIMIT $2`, topicID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("get videos by topic: %w", err)
	}
	defer rows.Close()

	return scanVideos(rows)
}

func (s *TopicStore) GetBubbleData(ctx context.Context, since time.Time) ([]model.BubbleData, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT t.id, t.name AS label, COUNT(tv.video_id) AS value,
		       t.color, '/genre/' || t.slug AS href,
		       COALESCE(t.parent_category, '') AS "group"
		FROM topics t
		JOIN topic_videos tv ON tv.topic_id = t.id
		JOIN videos v ON v.id = tv.video_id AND v.created_at >= $1
		GROUP BY t.id, t.name, t.color, t.slug, t.parent_category
		HAVING COUNT(tv.video_id) > 0
		ORDER BY value DESC`, since,
	)
	if err != nil {
		return nil, fmt.Errorf("get topic bubble data: %w", err)
	}
	defer rows.Close()

	out := make([]model.BubbleData, 0)
	for rows.Next() {
		var b model.BubbleData
		if err := rows.Scan(&b.ID, &b.Label, &b.Value, &b.Color, &b.Href, &b.Group); err != nil {
			return nil, fmt.Errorf("scan bubble data: %w", err)
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (s *TopicStore) UpsertTopic(ctx context.Context, t *model.Topic) (int64, error) {
	var id int64
	err := s.pool.QueryRow(ctx, `
		INSERT INTO topics (name, slug, description, color, parent_category, snapshot_date, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (slug) DO UPDATE SET
			name            = EXCLUDED.name,
			description     = EXCLUDED.description,
			color           = EXCLUDED.color,
			parent_category = EXCLUDED.parent_category,
			snapshot_date   = EXCLUDED.snapshot_date
		RETURNING id`,
		t.Name, t.Slug, t.Description, t.Color, t.ParentCategory, t.SnapshotDate,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("upsert topic: %w", err)
	}
	return id, nil
}

func (s *TopicStore) LinkTopicVideos(ctx context.Context, topicID int64, videoIDs []int64) error {
	if len(videoIDs) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	for _, vid := range videoIDs {
		batch.Queue(
			`INSERT INTO topic_videos (topic_id, video_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			topicID, vid,
		)
	}

	br := s.pool.SendBatch(ctx, batch)
	defer br.Close()

	for range videoIDs {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("link topic video: %w", err)
		}
	}
	return nil
}
