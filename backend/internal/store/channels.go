package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thumbtrend/backend/internal/model"
)

type ChannelStore struct {
	pool *pgxpool.Pool
}

func NewChannelStore(pool *pgxpool.Pool) *ChannelStore {
	return &ChannelStore{pool: pool}
}

func (s *ChannelStore) GetTrendingChannels(ctx context.Context, since time.Time, limit int) ([]model.Channel, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT c.id, c.youtube_channel_id, c.name, c.avatar_url,
		       c.subscriber_count, c.video_count, c.created_at, c.updated_at,
		       COUNT(v.id) AS trending_count
		FROM channels c
		JOIN videos v ON v.channel_db_id = c.id AND v.created_at >= $1
		GROUP BY c.id
		ORDER BY trending_count DESC
		LIMIT $2`, since, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("get trending channels: %w", err)
	}
	defer rows.Close()

	out := make([]model.Channel, 0)
	for rows.Next() {
		var ch model.Channel
		if err := rows.Scan(
			&ch.ID, &ch.YouTubeChannelID, &ch.Name, &ch.AvatarURL,
			&ch.SubscriberCount, &ch.VideoCount, &ch.CreatedAt, &ch.UpdatedAt,
			&ch.TrendingCount,
		); err != nil {
			return nil, fmt.Errorf("scan channel: %w", err)
		}
		out = append(out, ch)
	}
	return out, rows.Err()
}

func (s *ChannelStore) GetByID(ctx context.Context, id int64) (*model.Channel, error) {
	var ch model.Channel
	err := s.pool.QueryRow(ctx, `
		SELECT id, youtube_channel_id, name, avatar_url,
		       subscriber_count, video_count, created_at, updated_at
		FROM channels WHERE id = $1`, id,
	).Scan(
		&ch.ID, &ch.YouTubeChannelID, &ch.Name, &ch.AvatarURL,
		&ch.SubscriberCount, &ch.VideoCount, &ch.CreatedAt, &ch.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get channel by id: %w", err)
	}
	return &ch, nil
}

func (s *ChannelStore) GetHistory(ctx context.Context, channelID int64) ([]model.ChannelSnapshot, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, channel_id, subscriber_count, fetched_at
		FROM channel_snapshots
		WHERE channel_id = $1 AND fetched_at >= NOW() - INTERVAL '30 days'
		ORDER BY fetched_at ASC`, channelID,
	)
	if err != nil {
		return nil, fmt.Errorf("get channel history: %w", err)
	}
	defer rows.Close()

	out := make([]model.ChannelSnapshot, 0)
	for rows.Next() {
		var snap model.ChannelSnapshot
		if err := rows.Scan(&snap.ID, &snap.ChannelID, &snap.SubscriberCount, &snap.FetchedAt); err != nil {
			return nil, fmt.Errorf("scan channel snapshot: %w", err)
		}
		out = append(out, snap)
	}
	return out, rows.Err()
}

func (s *ChannelStore) GetBubbleData(ctx context.Context, since time.Time) ([]model.BubbleData, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT c.id, c.name AS label, COUNT(v.id) AS value,
		       '#6366f1' AS color, '' AS href, '' AS "group"
		FROM channels c
		JOIN videos v ON v.channel_db_id = c.id AND v.created_at >= $1
		GROUP BY c.id, c.name
		HAVING COUNT(v.id) > 0
		ORDER BY value DESC`, since,
	)
	if err != nil {
		return nil, fmt.Errorf("get channel bubble data: %w", err)
	}
	defer rows.Close()

	out := make([]model.BubbleData, 0)
	for rows.Next() {
		var b model.BubbleData
		if err := rows.Scan(&b.ID, &b.Label, &b.Value, &b.Color, &b.Href, &b.Group); err != nil {
			return nil, fmt.Errorf("scan channel bubble: %w", err)
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (s *ChannelStore) UpsertChannel(ctx context.Context, c *model.Channel) (int64, error) {
	var id int64
	err := s.pool.QueryRow(ctx, `
		INSERT INTO channels (youtube_channel_id, name, avatar_url, subscriber_count, video_count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		ON CONFLICT (youtube_channel_id) DO UPDATE SET
			name             = EXCLUDED.name,
			avatar_url       = EXCLUDED.avatar_url,
			subscriber_count = EXCLUDED.subscriber_count,
			video_count      = EXCLUDED.video_count,
			updated_at       = NOW()
		RETURNING id`,
		c.YouTubeChannelID, c.Name, c.AvatarURL, c.SubscriberCount, c.VideoCount,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("upsert channel: %w", err)
	}
	return id, nil
}

func (s *ChannelStore) InsertSnapshot(ctx context.Context, snap *model.ChannelSnapshot) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO channel_snapshots (channel_id, subscriber_count, fetched_at)
		VALUES ($1, $2, NOW())`,
		snap.ChannelID, snap.SubscriberCount,
	)
	if err != nil {
		return fmt.Errorf("insert channel snapshot: %w", err)
	}
	return nil
}

func (s *ChannelStore) GetAll(ctx context.Context) ([]model.Channel, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, youtube_channel_id, name, avatar_url,
		       subscriber_count, video_count, created_at, updated_at
		FROM channels ORDER BY updated_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("get all channels: %w", err)
	}
	defer rows.Close()

	out := make([]model.Channel, 0)
	for rows.Next() {
		var ch model.Channel
		if err := rows.Scan(
			&ch.ID, &ch.YouTubeChannelID, &ch.Name, &ch.AvatarURL,
			&ch.SubscriberCount, &ch.VideoCount, &ch.CreatedAt, &ch.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan channel: %w", err)
		}
		out = append(out, ch)
	}
	return out, rows.Err()
}

func (s *ChannelStore) PruneOldSnapshots(ctx context.Context, olderThan time.Time) error {
	_, err := s.pool.Exec(ctx,
		`DELETE FROM channel_snapshots WHERE fetched_at < $1`, olderThan,
	)
	if err != nil {
		return fmt.Errorf("prune old snapshots: %w", err)
	}
	return nil
}
