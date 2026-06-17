package usecase

import (
	"context"
	"refer-mate/internal/repository"
)

type TrackingUseCase struct {
	emailRepo    repository.EmailRepository
	campaignRepo repository.CampaignRepository
}

func NewTrackingUseCase(emailRepo repository.EmailRepository, campaignRepo repository.CampaignRepository) *TrackingUseCase {
	return &TrackingUseCase{emailRepo: emailRepo, campaignRepo: campaignRepo}
}

func (uc *TrackingUseCase) TrackOpen(ctx context.Context, emailID int64) error {
	email, err := uc.emailRepo.FindByID(ctx, emailID)
	if err != nil || email == nil {
		return err
	}
	if email.OpenedAt != nil {
		return nil
	}
	if err := uc.emailRepo.UpdateOpened(ctx, emailID); err != nil {
		return err
	}
	return uc.campaignRepo.IncrementOpened(ctx, email.CampaignID)
}

func (uc *TrackingUseCase) TrackClick(ctx context.Context, emailID int64) error {
	email, err := uc.emailRepo.FindByID(ctx, emailID)
	if err != nil || email == nil {
		return err
	}
	if email.ClickedAt != nil {
		return nil
	}
	if err := uc.emailRepo.UpdateClicked(ctx, emailID); err != nil {
		return err
	}
	return uc.campaignRepo.IncrementClicked(ctx, email.CampaignID)
}
