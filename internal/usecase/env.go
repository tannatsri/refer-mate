package usecase

import (
	"context"
	"go-template/internal/domain"
	"go-template/internal/repository"
)

type EnvUseCase struct {
	repo repository.EnvRepository
}

func NewEnvUseCase(r repository.EnvRepository) *EnvUseCase {
	return &EnvUseCase{repo: r}
}

func (u *EnvUseCase) ListRoutes(ctx context.Context) ([]domain.Env, error) {
	return u.repo.GetAll(ctx)
}
