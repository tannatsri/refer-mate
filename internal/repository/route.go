package repository

import (
	"context"
	"go-template/internal/domain"
)

type EnvRepository interface {
	GetAll(ctx context.Context) ([]domain.Env, error)
}
