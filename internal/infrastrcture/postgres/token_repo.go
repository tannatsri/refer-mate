package postgres

import (
	"context"
	"database/sql"
	"refer-mate/internal/domain"
	"refer-mate/internal/repository"
)

type tokenRepo struct{ db *sql.DB }

func NewTokenRepo(db *sql.DB) repository.TokenRepository {
	return &tokenRepo{db: db}
}

func (r *tokenRepo) Upsert(ctx context.Context, t *domain.OAuthToken) (*domain.OAuthToken, error) {
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO oauth_tokens (user_id, provider, access_token, refresh_token, expires_at)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (user_id) DO UPDATE
		 SET access_token=$3, refresh_token=$4, expires_at=$5, updated_at=NOW()
		 RETURNING id, user_id, provider, access_token, refresh_token, expires_at, created_at, updated_at`,
		t.UserID, t.Provider, t.AccessToken, t.RefreshToken, t.ExpiresAt,
	).Scan(&t.ID, &t.UserID, &t.Provider, &t.AccessToken, &t.RefreshToken, &t.ExpiresAt, &t.CreatedAt, &t.UpdatedAt)
	return t, err
}

func (r *tokenRepo) FindByUserID(ctx context.Context, userID int64) (*domain.OAuthToken, error) {
	t := &domain.OAuthToken{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, provider, access_token, refresh_token, expires_at, created_at, updated_at
		 FROM oauth_tokens WHERE user_id = $1`, userID,
	).Scan(&t.ID, &t.UserID, &t.Provider, &t.AccessToken, &t.RefreshToken, &t.ExpiresAt, &t.CreatedAt, &t.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return t, err
}
