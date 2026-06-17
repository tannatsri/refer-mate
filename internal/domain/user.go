package domain

import "time"

type User struct {
	ID             int64
	GoogleID       string
	Email          string
	Name           string
	ProfilePicture string
	IsActive       bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
