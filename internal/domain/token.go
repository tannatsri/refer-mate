package domain

import "time"

type OAuthToken struct {
	ID           int64
	UserID       int64
	Provider     string
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
