package postgres

import (
	"context"

	"go-template/internal/db"
	"go-template/internal/domain"
	"go-template/internal/repository"
)

type EnvRepo struct {
	queries *db.Queries
}

func NewEnvRepo(q *db.Queries) repository.EnvRepository {
	return &EnvRepo{queries: q}
}

func (r *EnvRepo) GetAll(ctx context.Context) ([]domain.Env, error) {
	rows, err := r.queries.ListRoutes(ctx)
	if err != nil {
		return nil, err
	}

	var routes []domain.Env
	for _, row := range rows {
		routes = append(routes, domain.Env{
			ID:        row.ID,
			Subdomain: row.Subdomain,
			TargetURL: row.TargetUrl,
		})
	}
	return routes, nil
}
