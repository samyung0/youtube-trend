package store

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thumbtrend/backend/internal/model"
)

type AnalysisStore struct {
	pool *pgxpool.Pool
}

func NewAnalysisStore(pool *pgxpool.Pool) *AnalysisStore {
	return &AnalysisStore{pool: pool}
}

func (s *AnalysisStore) GetStats(ctx context.Context, since time.Time) (*model.AnalysisStats, error) {
	var total int
	var faceCount int
	var avgBrightness float64
	var avgFaceCount float64

	err := s.pool.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COALESCE(SUM(CASE WHEN a.has_face THEN 1 ELSE 0 END), 0),
			COALESCE(AVG(a.brightness), 0),
			COALESCE(AVG(a.face_count), 0)
		FROM thumbnail_analyses a
		JOIN videos v ON v.id = a.video_id
		WHERE v.created_at >= $1`, since,
	).Scan(&total, &faceCount, &avgBrightness, &avgFaceCount)
	if err != nil {
		return nil, fmt.Errorf("get analysis stats: %w", err)
	}

	facePct := 0
	if total > 0 {
		facePct = faceCount * 100 / total
	}

	colorFreq, err := s.getColorFrequency(ctx, since)
	if err != nil {
		return nil, err
	}

	ocrWords, err := s.getOCRWordFrequency(ctx, since)
	if err != nil {
		return nil, err
	}

	return &model.AnalysisStats{
		TotalAnalyzed:  total,
		FacePercentage: facePct,
		AvgBrightness:  avgBrightness,
		AvgFaceCount:   avgFaceCount,
		ColorFrequency: colorFreq,
		OCRWords:       ocrWords,
	}, nil
}

func (s *AnalysisStore) getColorFrequency(ctx context.Context, since time.Time) (map[string]int, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT color, COUNT(*) AS cnt
		FROM (
			SELECT jsonb_array_elements_text(a.dominant_colors) AS color
			FROM thumbnail_analyses a
			JOIN videos v ON v.id = a.video_id
			WHERE v.created_at >= $1
		) sub
		GROUP BY color
		ORDER BY cnt DESC
		LIMIT 20`, since,
	)
	if err != nil {
		return nil, fmt.Errorf("get color frequency: %w", err)
	}
	defer rows.Close()

	freq := make(map[string]int)
	for rows.Next() {
		var color string
		var cnt int
		if err := rows.Scan(&color, &cnt); err != nil {
			return nil, fmt.Errorf("scan color freq: %w", err)
		}
		freq[color] = cnt
	}
	return freq, rows.Err()
}

func (s *AnalysisStore) getOCRWordFrequency(ctx context.Context, since time.Time) (map[string]int, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT a.ocr_text
		FROM thumbnail_analyses a
		JOIN videos v ON v.id = a.video_id
		WHERE v.created_at >= $1 AND a.ocr_text != ''`, since,
	)
	if err != nil {
		return nil, fmt.Errorf("get ocr texts: %w", err)
	}
	defer rows.Close()

	freq := make(map[string]int)
	for rows.Next() {
		var text string
		if err := rows.Scan(&text); err != nil {
			return nil, fmt.Errorf("scan ocr text: %w", err)
		}
		for _, word := range strings.Fields(strings.ToLower(text)) {
			word = strings.Trim(word, ".,!?;:\"'()-")
			if len(word) >= 3 {
				freq[word]++
			}
		}
	}
	return freq, rows.Err()
}

func (s *AnalysisStore) Upsert(ctx context.Context, a *model.ThumbnailAnalysis) error {
	colorsJSON, err := json.Marshal(a.DominantColors)
	if err != nil {
		return fmt.Errorf("marshal colors: %w", err)
	}

	_, err = s.pool.Exec(ctx, `
		INSERT INTO thumbnail_analyses (video_id, dominant_colors, has_face, face_count, ocr_text, brightness, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (video_id) DO UPDATE SET
			dominant_colors = EXCLUDED.dominant_colors,
			has_face        = EXCLUDED.has_face,
			face_count      = EXCLUDED.face_count,
			ocr_text        = EXCLUDED.ocr_text,
			brightness      = EXCLUDED.brightness`,
		a.VideoID, colorsJSON, a.HasFace, a.FaceCount, a.OCRText, a.Brightness,
	)
	if err != nil {
		return fmt.Errorf("upsert analysis: %w", err)
	}
	return nil
}

func (s *AnalysisStore) GetUnanalyzedVideos(ctx context.Context, since time.Time, limit int) ([]model.Video, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT v.id, v.youtube_id, v.title, v.channel_name, v.channel_id, v.channel_db_id,
		       v.thumbnail_url, v.view_count, v.like_count, v.comment_count,
		       v.category_id, v.tags, v.published_at, v.duration, v.is_short_video,
		       v.created_at, v.updated_at
		FROM videos v
		LEFT JOIN thumbnail_analyses a ON a.video_id = v.id
		WHERE a.id IS NULL AND v.created_at >= $1
		ORDER BY v.view_count DESC
		LIMIT $2`, since, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("get unanalyzed videos: %w", err)
	}
	defer rows.Close()

	return scanVideos(rows)
}
