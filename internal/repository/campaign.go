package repository

import (
	"context"
	"refer-mate/internal/domain"
)

type CampaignRepository interface {
	Create(ctx context.Context, c *domain.Campaign) (*domain.Campaign, error)
	FindByID(ctx context.Context, id, userID int64) (*domain.Campaign, error)
	ListByUserID(ctx context.Context, userID int64) ([]domain.Campaign, error)
	Update(ctx context.Context, c *domain.Campaign) (*domain.Campaign, error)
	UpdateStatus(ctx context.Context, id int64, status string) error
	IncrementSent(ctx context.Context, id int64) error
	IncrementOpened(ctx context.Context, id int64) error
	IncrementClicked(ctx context.Context, id int64) error
	IncrementFailed(ctx context.Context, id int64) error
}
