package repository

import (
	"context"
	"refer-mate/internal/domain"
)

type RecipientRepository interface {
	BulkCreate(ctx context.Context, recipients []domain.CampaignRecipient) error
	ListByCampaignID(ctx context.Context, campaignID int64) ([]domain.CampaignRecipient, error)
}
