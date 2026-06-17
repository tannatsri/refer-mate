package usecase

import (
	"context"
	"errors"
	"refer-mate/internal/domain"
	"refer-mate/internal/repository"
	"regexp"
)

var varPattern = regexp.MustCompile(`\{\{(\w+)\}\}`)

type TemplateUseCase struct {
	repo repository.TemplateRepository
}

func NewTemplateUseCase(repo repository.TemplateRepository) *TemplateUseCase {
	return &TemplateUseCase{repo: repo}
}

func (uc *TemplateUseCase) Create(ctx context.Context, userID int64, title, subject, body string) (*domain.EmailTemplate, error) {
	vars := extractVariables(subject + " " + body)
	return uc.repo.Create(ctx, &domain.EmailTemplate{
		UserID:    userID,
		Title:     title,
		Subject:   subject,
		Body:      body,
		Variables: vars,
	})
}

func (uc *TemplateUseCase) GetByID(ctx context.Context, id, userID int64) (*domain.EmailTemplate, error) {
	t, err := uc.repo.FindByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, errors.New("template not found")
	}
	return t, nil
}

func (uc *TemplateUseCase) List(ctx context.Context, userID int64) ([]domain.EmailTemplate, error) {
	return uc.repo.ListByUserID(ctx, userID)
}

func (uc *TemplateUseCase) Update(ctx context.Context, id, userID int64, title, subject, body string) (*domain.EmailTemplate, error) {
	existing, err := uc.repo.FindByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("template not found")
	}
	existing.Title = title
	existing.Subject = subject
	existing.Body = body
	existing.Variables = extractVariables(subject + " " + body)
	return uc.repo.Update(ctx, existing)
}

func (uc *TemplateUseCase) Delete(ctx context.Context, id, userID int64) error {
	return uc.repo.Delete(ctx, id, userID)
}

func extractVariables(text string) []string {
	matches := varPattern.FindAllStringSubmatch(text, -1)
	seen := map[string]bool{}
	var vars []string
	for _, m := range matches {
		if !seen[m[1]] {
			seen[m[1]] = true
			vars = append(vars, m[1])
		}
	}
	return vars
}
