package model

import "time"

type Topic struct {
	ID             int64     `json:"id" db:"id"`
	Name           string    `json:"name" db:"name"`
	Slug           string    `json:"slug" db:"slug"`
	Description    *string   `json:"description,omitempty" db:"description"`
	Color          string    `json:"color" db:"color"`
	ParentCategory *string   `json:"parent_category,omitempty" db:"parent_category"`
	SnapshotDate   string    `json:"snapshot_date" db:"snapshot_date"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	VideoCount     int       `json:"video_count,omitempty" db:"-"`
}

type TopicVideo struct {
	ID      int64 `json:"id" db:"id"`
	VideoID int64 `json:"video_id" db:"video_id"`
	TopicID int64 `json:"topic_id" db:"topic_id"`
}

type BubbleData struct {
	ID    int64  `json:"id"`
	Label string `json:"label"`
	Value int    `json:"value"`
	Color string `json:"color"`
	Href  string `json:"href"`
	Group string `json:"group,omitempty"`
}
