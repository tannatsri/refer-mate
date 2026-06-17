package repository

import (
	"context"
	"refer-mate/internal/domain"
)

type TokenRepository interface {
	Upsert(ctx context.Context, token *domain.OAuthToken) (*domain.OAuthToken, error)
	FindByUserID(ctx context.Context, userID int64) (*domain.OAuthToken, error)
}
