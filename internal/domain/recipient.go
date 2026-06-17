package domain

import "time"

type CampaignRecipient struct {
	ID              int64
	CampaignID      int64
	RecipientName   string
	RecipientEmail  string
	Company         string
	Role            string
	CustomVariables map[string]string
	CreatedAt       time.Time
}
