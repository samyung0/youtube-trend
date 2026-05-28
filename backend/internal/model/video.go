package model

import "time"

type Video struct {
	ID           int64     `json:"id" db:"id"`
	YouTubeID    string    `json:"youtube_id" db:"youtube_id"`
	Title        string    `json:"title" db:"title"`
	ChannelName  string    `json:"channel_name" db:"channel_name"`
	ChannelID    string    `json:"channel_id" db:"channel_id"`
	ChannelDBID  *int64    `json:"channel_db_id,omitempty" db:"channel_db_id"`
	ThumbnailURL string    `json:"thumbnail_url" db:"thumbnail_url"`
	ViewCount    int64     `json:"view_count" db:"view_count"`
	LikeCount    int64     `json:"like_count" db:"like_count"`
	CommentCount int64     `json:"comment_count" db:"comment_count"`
	CategoryID   int       `json:"category_id" db:"category_id"`
	Tags         []string  `json:"tags" db:"tags"`
	PublishedAt  *time.Time `json:"published_at,omitempty" db:"published_at"`
	Duration     string    `json:"duration" db:"duration"`
	IsShortVideo bool      `json:"is_short_video" db:"is_short_video"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type TrendingSnapshot struct {
	ID         int64     `json:"id" db:"id"`
	FetchedAt  time.Time `json:"fetched_at" db:"fetched_at"`
	Region     string    `json:"region" db:"region"`
	CategoryID *int      `json:"category_id,omitempty" db:"category_id"`
	VideoCount int       `json:"video_count" db:"video_count"`
}

type SnapshotVideo struct {
	ID         int64 `json:"id" db:"id"`
	SnapshotID int64 `json:"snapshot_id" db:"snapshot_id"`
	VideoID    int64 `json:"video_id" db:"video_id"`
	Rank       int   `json:"rank" db:"rank"`
}
