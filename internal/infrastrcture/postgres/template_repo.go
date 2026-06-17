package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"refer-mate/internal/domain"
	"refer-mate/internal/repository"
)

type templateRepo struct{ db *sql.DB }

func NewTemplateRepo(db *sql.DB) repository.TemplateRepository {
	return &templateRepo{db: db}
}

func (r *templateRepo) Create(ctx context.Context, t *domain.EmailTemplate) (*domain.EmailTemplate, error) {
	vars, _ := json.Marshal(t.Variables)
	var rawVars []byte
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO email_templates (user_id, title, subject, body, variables)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, user_id, title, subject, body, variables, created_at, updated_at`,
		t.UserID, t.Title, t.Subject, t.Body, vars,
	).Scan(&t.ID, &t.UserID, &t.Title, &t.Subject, &t.Body, &rawVars, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(rawVars, &t.Variables)
	return t, nil
}

func (r *templateRepo) FindByID(ctx context.Context, id, userID int64) (*domain.EmailTemplate, error) {
	t := &domain.EmailTemplate{}
	var rawVars []byte
	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, title, subject, body, variables, created_at, updated_at
		 FROM email_templates WHERE id=$1 AND user_id=$2`, id, userID,
	).Scan(&t.ID, &t.UserID, &t.Title, &t.Subject, &t.Body, &rawVars, &t.CreatedAt, &t.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	json.Unmarshal(rawVars, &t.Variables)
	return t, nil
}

func (r *templateRepo) ListByUserID(ctx context.Context, userID int64) ([]domain.EmailTemplate, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, title, subject, body, variables, created_at, updated_at
		 FROM email_templates WHERE user_id=$1 ORDER BY created_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []domain.EmailTemplate
	for rows.Next() {
		var t domain.EmailTemplate
		var rawVars []byte
		if err := rows.Scan(&t.ID, &t.UserID, &t.Title, &t.Subject, &t.Body, &rawVars, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal(rawVars, &t.Variables)
		templates = append(templates, t)
	}
	return templates, rows.Err()
}

func (r *templateRepo) Update(ctx context.Context, t *domain.EmailTemplate) (*domain.EmailTemplate, error) {
	vars, _ := json.Marshal(t.Variables)
	var rawVars []byte
	err := r.db.QueryRowContext(ctx,
		`UPDATE email_templates SET title=$1, subject=$2, body=$3, variables=$4, updated_at=NOW()
		 WHERE id=$5 AND user_id=$6
		 RETURNING id, user_id, title, subject, body, variables, created_at, updated_at`,
		t.Title, t.Subject, t.Body, vars, t.ID, t.UserID,
	).Scan(&t.ID, &t.UserID, &t.Title, &t.Subject, &t.Body, &rawVars, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(rawVars, &t.Variables)
	return t, nil
}

func (r *templateRepo) Delete(ctx context.Context, id, userID int64) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM email_templates WHERE id=$1 AND user_id=$2`, id, userID,
	)
	return err
}
