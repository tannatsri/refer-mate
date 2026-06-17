package repository

import (
	"context"
	"refer-mate/internal/domain"
)

type TemplateRepository interface {
	Create(ctx context.Context, t *domain.EmailTemplate) (*domain.EmailTemplate, error)
	FindByID(ctx context.Context, id, userID int64) (*domain.EmailTemplate, error)
	ListByUserID(ctx context.Context, userID int64) ([]domain.EmailTemplate, error)
	Update(ctx context.Context, t *domain.EmailTemplate) (*domain.EmailTemplate, error)
	Delete(ctx context.Context, id, userID int64) error
}
