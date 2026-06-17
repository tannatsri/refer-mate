package repository

import (
	"context"
	"refer-mate/internal/domain"
)

type UserRepository interface {
	FindByGoogleID(ctx context.Context, googleID string) (*domain.User, error)
	FindByID(ctx context.Context, id int64) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) (*domain.User, error)
}
