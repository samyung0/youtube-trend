package model

import "time"

type User struct {
	ID        string    `json:"id" db:"id"`
	GoogleID  string    `json:"google_id" db:"google_id"`
	Email     string    `json:"email" db:"email"`
	Name      string    `json:"name" db:"name"`
	AvatarURL string    `json:"avatar_url" db:"avatar_url"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Subscription struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Plan      string    `json:"plan" db:"plan"`
	Status    string    `json:"status" db:"status"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func (s *Subscription) IsActive() bool {
	return s.Status == "active" && s.ExpiresAt.After(time.Now())
}
