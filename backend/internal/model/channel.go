package model

import "time"

type Channel struct {
	ID               int64     `json:"id" db:"id"`
	YouTubeChannelID string    `json:"youtube_channel_id" db:"youtube_channel_id"`
	Name             string    `json:"name" db:"name"`
	AvatarURL        string    `json:"avatar_url" db:"avatar_url"`
	SubscriberCount  int64     `json:"subscriber_count" db:"subscriber_count"`
	VideoCount       int       `json:"video_count" db:"video_count"`
	TrendingCount    int       `json:"trending_count,omitempty" db:"-"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

type ChannelSnapshot struct {
	ID              int64     `json:"id" db:"id"`
	ChannelID       int64     `json:"channel_id" db:"channel_id"`
	SubscriberCount int64     `json:"subscriber_count" db:"subscriber_count"`
	FetchedAt       time.Time `json:"fetched_at" db:"fetched_at"`
}
