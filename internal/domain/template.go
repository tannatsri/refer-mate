package domain

import "time"

type EmailTemplate struct {
	ID        int64
	UserID    int64
	Title     string
	Subject   string
	Body      string
	Variables []string
	CreatedAt time.Time
	UpdatedAt time.Time
}
