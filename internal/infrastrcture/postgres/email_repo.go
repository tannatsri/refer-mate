package postgres

import (
	"context"
	"database/sql"
	"refer-mate/internal/domain"
	"refer-mate/internal/repository"
)

type emailRepo struct{ db *sql.DB }

func NewEmailRepo(db *sql.DB) repository.EmailRepository {
	return &emailRepo{db: db}
}

const emailCols = `id, campaign_id, recipient_id, gmail_message_id, gmail_thread_id,
	subject, body, status, error_message, sent_at, opened_at, clicked_at, created_at, updated_at`

func scanEmail(row interface{ Scan(...interface{}) error }) (*domain.Email, error) {
	e := &domain.Email{}
	return e, row.Scan(
		&e.ID, &e.CampaignID, &e.RecipientID, &e.GmailMessageID, &e.GmailThreadID,
		&e.Subject, &e.Body, &e.Status, &e.ErrorMessage,
		&e.SentAt, &e.OpenedAt, &e.ClickedAt, &e.CreatedAt, &e.UpdatedAt,
	)
}

func (r *emailRepo) Create(ctx context.Context, e *domain.Email) (*domain.Email, error) {
	row := r.db.QueryRowContext(ctx,
		`INSERT INTO emails (campaign_id, recipient_id, subject, body, status)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING `+emailCols,
		e.CampaignID, e.RecipientID, e.Subject, e.Body, domain.EmailStatusQueued,
	)
	return scanEmail(row)
}

func (r *emailRepo) FindByID(ctx context.Context, id int64) (*domain.Email, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+emailCols+` FROM emails WHERE id=$1`, id,
	)
	e, err := scanEmail(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return e, err
}

func (r *emailRepo) UpdateSent(ctx context.Context, id int64, gmailMsgID, gmailThreadID string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE emails SET status=$1, gmail_message_id=$2, gmail_thread_id=$3, sent_at=NOW(), updated_at=NOW()
		 WHERE id=$4`,
		domain.EmailStatusSent, gmailMsgID, gmailThreadID, id,
	)
	return err
}

func (r *emailRepo) UpdateOpened(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE emails SET status=$1, opened_at=NOW(), updated_at=NOW() WHERE id=$2 AND opened_at IS NULL`,
		domain.EmailStatusOpened, id,
	)
	return err
}

func (r *emailRepo) UpdateClicked(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE emails SET status=$1, clicked_at=NOW(), updated_at=NOW() WHERE id=$2 AND clicked_at IS NULL`,
		domain.EmailStatusClicked, id,
	)
	return err
}

func (r *emailRepo) UpdateFailed(ctx context.Context, id int64, errMsg string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE emails SET status=$1, error_message=$2, updated_at=NOW() WHERE id=$3`,
		domain.EmailStatusFailed, errMsg, id,
	)
	return err
}

func (r *emailRepo) ListByCampaignID(ctx context.Context, campaignID int64) ([]domain.Email, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+emailCols+` FROM emails WHERE campaign_id=$1 ORDER BY created_at`, campaignID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var emails []domain.Email
	for rows.Next() {
		e, err := scanEmail(rows)
		if err != nil {
			return nil, err
		}
		emails = append(emails, *e)
	}
	return emails, rows.Err()
}
