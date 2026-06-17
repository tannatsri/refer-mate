package postgres

import (
	"context"
	"database/sql"
	"refer-mate/internal/domain"
	"refer-mate/internal/repository"
)

type campaignRepo struct{ db *sql.DB }

func NewCampaignRepo(db *sql.DB) repository.CampaignRepository {
	return &campaignRepo{db: db}
}

const campaignCols = `id, user_id, template_id, campaign_name, description, status,
	scheduled_at, total_recipients, sent_count, opened_count, clicked_count,
	replied_count, failed_count, created_at, updated_at`

func scanCampaign(row interface {
	Scan(...interface{}) error
}) (*domain.Campaign, error) {
	c := &domain.Campaign{}
	err := row.Scan(
		&c.ID, &c.UserID, &c.TemplateID, &c.CampaignName, &c.Description, &c.Status,
		&c.ScheduledAt, &c.TotalRecipients, &c.SentCount, &c.OpenedCount, &c.ClickedCount,
		&c.RepliedCount, &c.FailedCount, &c.CreatedAt, &c.UpdatedAt,
	)
	return c, err
}

func (r *campaignRepo) Create(ctx context.Context, c *domain.Campaign) (*domain.Campaign, error) {
	row := r.db.QueryRowContext(ctx,
		`INSERT INTO campaigns (user_id, template_id, campaign_name, description, status, scheduled_at)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING `+campaignCols,
		c.UserID, c.TemplateID, c.CampaignName, c.Description, domain.CampaignStatusDraft, c.ScheduledAt,
	)
	return scanCampaign(row)
}

func (r *campaignRepo) FindByID(ctx context.Context, id, userID int64) (*domain.Campaign, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+campaignCols+` FROM campaigns WHERE id=$1 AND user_id=$2`, id, userID,
	)
	c, err := scanCampaign(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return c, err
}

func (r *campaignRepo) ListByUserID(ctx context.Context, userID int64) ([]domain.Campaign, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+campaignCols+` FROM campaigns WHERE user_id=$1 ORDER BY created_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var campaigns []domain.Campaign
	for rows.Next() {
		c, err := scanCampaign(rows)
		if err != nil {
			return nil, err
		}
		campaigns = append(campaigns, *c)
	}
	return campaigns, rows.Err()
}

func (r *campaignRepo) Update(ctx context.Context, c *domain.Campaign) (*domain.Campaign, error) {
	row := r.db.QueryRowContext(ctx,
		`UPDATE campaigns SET campaign_name=$1, description=$2, template_id=$3, scheduled_at=$4, updated_at=NOW()
		 WHERE id=$5 AND user_id=$6
		 RETURNING `+campaignCols,
		c.CampaignName, c.Description, c.TemplateID, c.ScheduledAt, c.ID, c.UserID,
	)
	return scanCampaign(row)
}

func (r *campaignRepo) UpdateStatus(ctx context.Context, id int64, status string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE campaigns SET status=$1, updated_at=NOW() WHERE id=$2`, status, id,
	)
	return err
}

func (r *campaignRepo) IncrementSent(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE campaigns SET sent_count=sent_count+1, updated_at=NOW() WHERE id=$1`, id,
	)
	return err
}

func (r *campaignRepo) IncrementOpened(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE campaigns SET opened_count=opened_count+1, updated_at=NOW() WHERE id=$1`, id,
	)
	return err
}

func (r *campaignRepo) IncrementClicked(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE campaigns SET clicked_count=clicked_count+1, updated_at=NOW() WHERE id=$1`, id,
	)
	return err
}

func (r *campaignRepo) IncrementFailed(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE campaigns SET failed_count=failed_count+1, updated_at=NOW() WHERE id=$1`, id,
	)
	return err
}
