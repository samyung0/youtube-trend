package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thumbtrend/backend/internal/model"
)

type UserStore struct {
	pool *pgxpool.Pool
}

func NewUserStore(pool *pgxpool.Pool) *UserStore {
	return &UserStore{pool: pool}
}

func (s *UserStore) FindByGoogleID(ctx context.Context, googleID string) (*model.User, error) {
	var u model.User
	err := s.pool.QueryRow(ctx, `
		SELECT id, google_id, email, name, avatar_url, created_at, updated_at
		FROM users WHERE google_id = $1`, googleID,
	).Scan(&u.ID, &u.GoogleID, &u.Email, &u.Name, &u.AvatarURL, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find user by google id: %w", err)
	}
	return &u, nil
}

func (s *UserStore) Create(ctx context.Context, user *model.User) error {
	user.ID = uuid.NewString()
	_, err := s.pool.Exec(ctx, `
		INSERT INTO users (id, google_id, email, name, avatar_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())`,
		user.ID, user.GoogleID, user.Email, user.Name, user.AvatarURL,
	)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (s *UserStore) GetSubscription(ctx context.Context, userID string) (*model.Subscription, error) {
	var sub model.Subscription
	err := s.pool.QueryRow(ctx, `
		SELECT id, user_id, plan, status, expires_at, created_at
		FROM subscriptions WHERE user_id = $1
		ORDER BY created_at DESC LIMIT 1`, userID,
	).Scan(&sub.ID, &sub.UserID, &sub.Plan, &sub.Status, &sub.ExpiresAt, &sub.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get subscription: %w", err)
	}
	return &sub, nil
}

func (s *UserStore) FindByID(ctx context.Context, id string) (*model.User, error) {
	var u model.User
	err := s.pool.QueryRow(ctx, `
		SELECT id, google_id, email, name, avatar_url, created_at, updated_at
		FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.GoogleID, &u.Email, &u.Name, &u.AvatarURL, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return &u, nil
}

func (s *UserStore) IsPro(ctx context.Context, userID string) (bool, error) {
	sub, err := s.GetSubscription(ctx, userID)
	if err != nil {
		return false, err
	}
	if sub == nil {
		return false, nil
	}
	return sub.IsActive() && sub.Plan == "pro", nil
}
