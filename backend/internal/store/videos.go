package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thumbtrend/backend/internal/model"
)

type VideoStore struct {
	pool *pgxpool.Pool
}

func NewVideoStore(pool *pgxpool.Pool) *VideoStore {
	return &VideoStore{pool: pool}
}

func (s *VideoStore) GetTrending(ctx context.Context, since time.Time, categoryID *int, limit int) ([]model.Video, error) {
	query := `
		SELECT id, youtube_id, title, channel_name, channel_id, channel_db_id,
		       thumbnail_url, view_count, like_count, comment_count,
		       category_id, tags, published_at, duration, is_short_video, created_at, updated_at
		FROM videos
		WHERE created_at >= $1`
	args := []any{since}

	if categoryID != nil {
		query += ` AND category_id = $2`
		args = append(args, *categoryID)
	}

	query += ` ORDER BY view_count DESC LIMIT ` + fmt.Sprintf("%d", limit)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get trending: %w", err)
	}
	defer rows.Close()

	return scanVideos(rows)
}

// GetUnclustered returns videos that have no row in topic_videos yet.
func (s *VideoStore) GetUnclustered(ctx context.Context) ([]model.Video, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT v.id, v.youtube_id, v.title, v.channel_name, v.channel_id, v.channel_db_id,
		       v.thumbnail_url, v.view_count, v.like_count, v.comment_count,
		       v.category_id, v.tags, v.published_at, v.duration, v.is_short_video,
		       v.created_at, v.updated_at
		FROM videos v
		LEFT JOIN topic_videos tv ON tv.video_id = v.id
		WHERE tv.video_id IS NULL AND v.is_short_video = false
		ORDER BY v.view_count DESC`)
	if err != nil {
		return nil, fmt.Errorf("get unclustered videos: %w", err)
	}
	defer rows.Close()

	return scanVideos(rows)
}

func (s *VideoStore) UpsertVideo(ctx context.Context, v *model.Video) (int64, error) {
	var id int64
	err := s.pool.QueryRow(ctx, `
		INSERT INTO videos (youtube_id, title, channel_name, channel_id, channel_db_id,
		                     thumbnail_url, view_count, like_count, comment_count,
		                     category_id, tags, published_at, duration, is_short_video,
		                     created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,NOW(),NOW())
		ON CONFLICT (youtube_id) DO UPDATE SET
			title           = EXCLUDED.title,
			view_count      = EXCLUDED.view_count,
			like_count      = EXCLUDED.like_count,
			comment_count   = EXCLUDED.comment_count,
			thumbnail_url   = EXCLUDED.thumbnail_url,
			tags            = EXCLUDED.tags,
			duration        = EXCLUDED.duration,
			is_short_video  = EXCLUDED.is_short_video,
			updated_at      = NOW()
		RETURNING id`,
		v.YouTubeID, v.Title, v.ChannelName, v.ChannelID, v.ChannelDBID,
		v.ThumbnailURL, v.ViewCount, v.LikeCount, v.CommentCount,
		v.CategoryID, v.Tags, v.PublishedAt, v.Duration, v.IsShortVideo,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("upsert video: %w", err)
	}
	return id, nil
}

func (s *VideoStore) CreateSnapshot(ctx context.Context, snap *model.TrendingSnapshot) (int64, error) {
	var id int64
	err := s.pool.QueryRow(ctx, `
		INSERT INTO trending_snapshots (fetched_at, region, category_id, video_count)
		VALUES (NOW(), $1, $2, $3)
		RETURNING id`,
		snap.Region, snap.CategoryID, snap.VideoCount,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("create snapshot: %w", err)
	}
	return id, nil
}

func (s *VideoStore) LinkSnapshotVideos(ctx context.Context, snapshotID int64, videoIDs []int64) error {
	if len(videoIDs) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	for rank, vid := range videoIDs {
		batch.Queue(
			`INSERT INTO snapshot_videos (snapshot_id, video_id, rank) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`,
			snapshotID, vid, rank+1,
		)
	}

	br := s.pool.SendBatch(ctx, batch)
	defer br.Close()

	for range videoIDs {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("link snapshot video: %w", err)
		}
	}
	return nil
}

func (s *VideoStore) GetByID(ctx context.Context, id int64) (*model.Video, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, youtube_id, title, channel_name, channel_id, channel_db_id,
		       thumbnail_url, view_count, like_count, comment_count,
		       category_id, tags, published_at, duration, is_short_video, created_at, updated_at
		FROM videos WHERE id = $1`, id,
	)
	if err != nil {
		return nil, fmt.Errorf("get video by id: %w", err)
	}
	defer rows.Close()

	vids, err := scanVideos(rows)
	if err != nil {
		return nil, err
	}
	if len(vids) == 0 {
		return nil, nil
	}
	return &vids[0], nil
}

func (s *VideoStore) LinkVideoChannel(ctx context.Context, videoID, channelID int64) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE videos SET channel_db_id = $1 WHERE id = $2`, channelID, videoID)
	return err
}

func scanVideos(rows pgx.Rows) ([]model.Video, error) {
	out := make([]model.Video, 0)
	for rows.Next() {
		var v model.Video
		if err := rows.Scan(
			&v.ID, &v.YouTubeID, &v.Title, &v.ChannelName, &v.ChannelID, &v.ChannelDBID,
			&v.ThumbnailURL, &v.ViewCount, &v.LikeCount, &v.CommentCount,
			&v.CategoryID, &v.Tags, &v.PublishedAt, &v.Duration, &v.IsShortVideo,
			&v.CreatedAt, &v.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan video: %w", err)
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
