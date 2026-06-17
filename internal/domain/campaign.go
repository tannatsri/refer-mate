package domain

import "time"

const (
	CampaignStatusDraft     = "DRAFT"
	CampaignStatusRunning   = "RUNNING"
	CampaignStatusPaused    = "PAUSED"
	CampaignStatusCompleted = "COMPLETED"
	CampaignStatusFailed    = "FAILED"
)

type Campaign struct {
	ID              int64
	UserID          int64
	TemplateID      int64
	CampaignName    string
	Description     string
	Status          string
	ScheduledAt     *time.Time
	TotalRecipients int
	SentCount       int
	OpenedCount     int
	ClickedCount    int
	RepliedCount    int
	FailedCount     int
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
