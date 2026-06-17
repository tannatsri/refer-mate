package domain

import "time"

const (
	EmailStatusQueued  = "QUEUED"
	EmailStatusSent    = "SENT"
	EmailStatusOpened  = "OPENED"
	EmailStatusClicked = "CLICKED"
	EmailStatusReplied = "REPLIED"
	EmailStatusFailed  = "FAILED"
)

type Email struct {
	ID             int64
	CampaignID     int64
	RecipientID    int64
	GmailMessageID string
	GmailThreadID  string
	Subject        string
	Body           string
	Status         string
	ErrorMessage   string
	SentAt         *time.Time
	OpenedAt       *time.Time
	ClickedAt      *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
