package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"refer-mate/internal/domain"
	"refer-mate/internal/infrastrcture/gmail"
	"refer-mate/internal/repository"
	"regexp"
	"strings"

	"golang.org/x/oauth2"
)

type CampaignUseCase struct {
	campaignRepo  repository.CampaignRepository
	recipientRepo repository.RecipientRepository
	emailRepo     repository.EmailRepository
	templateRepo  repository.TemplateRepository
	gmailClient   *gmail.Client
	baseURL       string
}

func NewCampaignUseCase(
	campaignRepo repository.CampaignRepository,
	recipientRepo repository.RecipientRepository,
	emailRepo repository.EmailRepository,
	templateRepo repository.TemplateRepository,
	gmailClient *gmail.Client,
	baseURL string,
) *CampaignUseCase {
	return &CampaignUseCase{
		campaignRepo:  campaignRepo,
		recipientRepo: recipientRepo,
		emailRepo:     emailRepo,
		templateRepo:  templateRepo,
		gmailClient:   gmailClient,
		baseURL:       baseURL,
	}
}

func (uc *CampaignUseCase) Create(ctx context.Context, userID, templateID int64, name, description string) (*domain.Campaign, error) {
	return uc.campaignRepo.Create(ctx, &domain.Campaign{
		UserID:       userID,
		TemplateID:   templateID,
		CampaignName: name,
		Description:  description,
	})
}

func (uc *CampaignUseCase) GetByID(ctx context.Context, id, userID int64) (*domain.Campaign, error) {
	c, err := uc.campaignRepo.FindByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, errors.New("campaign not found")
	}
	return c, nil
}

func (uc *CampaignUseCase) List(ctx context.Context, userID int64) ([]domain.Campaign, error) {
	return uc.campaignRepo.ListByUserID(ctx, userID)
}

func (uc *CampaignUseCase) Update(ctx context.Context, id, userID, templateID int64, name, description string) (*domain.Campaign, error) {
	c, err := uc.campaignRepo.FindByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, errors.New("campaign not found")
	}
	if c.Status != domain.CampaignStatusDraft {
		return nil, errors.New("only DRAFT campaigns can be updated")
	}
	c.CampaignName = name
	c.Description = description
	c.TemplateID = templateID
	return uc.campaignRepo.Update(ctx, c)
}

func (uc *CampaignUseCase) AddRecipients(ctx context.Context, campaignID, userID int64, recipients []domain.CampaignRecipient) error {
	c, err := uc.campaignRepo.FindByID(ctx, campaignID, userID)
	if err != nil {
		return err
	}
	if c == nil {
		return errors.New("campaign not found")
	}
	if c.Status != domain.CampaignStatusDraft {
		return errors.New("only DRAFT campaigns accept recipients")
	}

	for i := range recipients {
		recipients[i].CampaignID = campaignID
	}

	if err := uc.recipientRepo.BulkCreate(ctx, recipients); err != nil {
		return err
	}

	_, err = uc.campaignRepo.Update(ctx, &domain.Campaign{
		ID:              c.ID,
		UserID:          c.UserID,
		TemplateID:      c.TemplateID,
		CampaignName:    c.CampaignName,
		Description:     c.Description,
		TotalRecipients: c.TotalRecipients + len(recipients),
	})
	return err
}

func (uc *CampaignUseCase) Launch(ctx context.Context, campaignID, userID int64, oauthToken *oauth2.Token, senderEmail string) error {
	c, err := uc.campaignRepo.FindByID(ctx, campaignID, userID)
	if err != nil {
		return err
	}
	if c == nil {
		return errors.New("campaign not found")
	}
	if c.Status != domain.CampaignStatusDraft && c.Status != domain.CampaignStatusPaused {
		return fmt.Errorf("campaign cannot be launched from status %s", c.Status)
	}

	tmpl, err := uc.templateRepo.FindByID(ctx, c.TemplateID, userID)
	if err != nil {
		return err
	}
	if tmpl == nil {
		return errors.New("template not found")
	}

	recipients, err := uc.recipientRepo.ListByCampaignID(ctx, campaignID)
	if err != nil {
		return err
	}

	if err := uc.campaignRepo.UpdateStatus(ctx, campaignID, domain.CampaignStatusRunning); err != nil {
		return err
	}

	go uc.sendEmails(campaignID, oauthToken, senderEmail, tmpl, recipients)
	return nil
}

func (uc *CampaignUseCase) Pause(ctx context.Context, id, userID int64) error {
	c, err := uc.campaignRepo.FindByID(ctx, id, userID)
	if err != nil {
		return err
	}
	if c == nil {
		return errors.New("campaign not found")
	}
	if c.Status != domain.CampaignStatusRunning {
		return errors.New("only RUNNING campaigns can be paused")
	}
	return uc.campaignRepo.UpdateStatus(ctx, id, domain.CampaignStatusPaused)
}

func (uc *CampaignUseCase) GetAnalytics(ctx context.Context, id, userID int64) (*domain.Campaign, []domain.Email, error) {
	c, err := uc.campaignRepo.FindByID(ctx, id, userID)
	if err != nil {
		return nil, nil, err
	}
	if c == nil {
		return nil, nil, errors.New("campaign not found")
	}

	emails, err := uc.emailRepo.ListByCampaignID(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	return c, emails, nil
}

func (uc *CampaignUseCase) sendEmails(campaignID int64, oauthToken *oauth2.Token, senderEmail string, tmpl *domain.EmailTemplate, recipients []domain.CampaignRecipient) {
	ctx := context.Background()

	for _, rec := range recipients {
		subject := substituteVariables(tmpl.Subject, &rec)
		body := substituteVariables(tmpl.Body, &rec)

		email, err := uc.emailRepo.Create(ctx, &domain.Email{
			CampaignID:  campaignID,
			RecipientID: rec.ID,
			Subject:     subject,
			Body:        body,
		})
		if err != nil {
			log.Printf("failed to create email record for recipient %d: %v", rec.ID, err)
			uc.campaignRepo.IncrementFailed(ctx, campaignID)
			continue
		}

		trackingPixel := fmt.Sprintf(`<img src="%s/track/open/%d" width="1" height="1" style="display:none" alt="">`, uc.baseURL, email.ID)
		body = wrapLinksForTracking(body, uc.baseURL, email.ID)
		body = body + trackingPixel

		result, err := uc.gmailClient.SendEmail(ctx, oauthToken, senderEmail, rec.RecipientEmail, subject, body)
		if err != nil {
			log.Printf("failed to send email %d: %v", email.ID, err)
			uc.emailRepo.UpdateFailed(ctx, email.ID, err.Error())
			uc.campaignRepo.IncrementFailed(ctx, campaignID)
			continue
		}

		uc.emailRepo.UpdateSent(ctx, email.ID, result.MessageID, result.ThreadID)
		uc.campaignRepo.IncrementSent(ctx, campaignID)
	}

	uc.campaignRepo.UpdateStatus(ctx, campaignID, domain.CampaignStatusCompleted)
}

var linkPattern = regexp.MustCompile(`href="(https?://[^"]+)"`)

func wrapLinksForTracking(body, baseURL string, emailID int64) string {
	return linkPattern.ReplaceAllStringFunc(body, func(match string) string {
		sub := linkPattern.FindStringSubmatch(match)
		if len(sub) < 2 {
			return match
		}
		originalURL := sub[1]
		wrappedURL := fmt.Sprintf("%s/track/click/%d?url=%s", baseURL, emailID, originalURL)
		return fmt.Sprintf(`href="%s"`, wrappedURL)
	})
}

func substituteVariables(text string, rec *domain.CampaignRecipient) string {
	vars := map[string]string{
		"name":           rec.RecipientName,
		"recipient_name": rec.RecipientName,
		"email":          rec.RecipientEmail,
		"company":        rec.Company,
		"role":           rec.Role,
	}
	for k, v := range rec.CustomVariables {
		vars[k] = v
	}

	result := text
	for k, v := range vars {
		result = strings.ReplaceAll(result, "{{"+k+"}}", v)
	}
	return result
}
