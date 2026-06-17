package repository

import (
	"context"
	"refer-mate/internal/domain"
)

type EmailRepository interface {
	Create(ctx context.Context, e *domain.Email) (*domain.Email, error)
	FindByID(ctx context.Context, id int64) (*domain.Email, error)
	UpdateSent(ctx context.Context, id int64, gmailMsgID, gmailThreadID string) error
	UpdateOpened(ctx context.Context, id int64) error
	UpdateClicked(ctx context.Context, id int64) error
	UpdateFailed(ctx context.Context, id int64, errMsg string) error
	ListByCampaignID(ctx context.Context, campaignID int64) ([]domain.Email, error)
}
