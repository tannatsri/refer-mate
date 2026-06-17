package postgres

import (
	"context"
	"database/sql"
	"refer-mate/internal/domain"
	"refer-mate/internal/repository"
)

type userRepo struct{ db *sql.DB }

func NewUserRepo(db *sql.DB) repository.UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) FindByGoogleID(ctx context.Context, googleID string) (*domain.User, error) {
	u := &domain.User{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, google_id, email, name, profile_picture, is_active, created_at, updated_at
		 FROM users WHERE google_id = $1`, googleID,
	).Scan(&u.ID, &u.GoogleID, &u.Email, &u.Name, &u.ProfilePicture, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return u, err
}

func (r *userRepo) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	u := &domain.User{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, google_id, email, name, profile_picture, is_active, created_at, updated_at
		 FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.GoogleID, &u.Email, &u.Name, &u.ProfilePicture, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return u, err
}

func (r *userRepo) Create(ctx context.Context, u *domain.User) (*domain.User, error) {
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO users (google_id, email, name, profile_picture)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, google_id, email, name, profile_picture, is_active, created_at, updated_at`,
		u.GoogleID, u.Email, u.Name, u.ProfilePicture,
	).Scan(&u.ID, &u.GoogleID, &u.Email, &u.Name, &u.ProfilePicture, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}

func (r *userRepo) Update(ctx context.Context, u *domain.User) (*domain.User, error) {
	err := r.db.QueryRowContext(ctx,
		`UPDATE users SET name=$1, profile_picture=$2, updated_at=NOW()
		 WHERE id=$3
		 RETURNING id, google_id, email, name, profile_picture, is_active, created_at, updated_at`,
		u.Name, u.ProfilePicture, u.ID,
	).Scan(&u.ID, &u.GoogleID, &u.Email, &u.Name, &u.ProfilePicture, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}
