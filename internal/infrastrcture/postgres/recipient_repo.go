package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"refer-mate/internal/domain"
	"refer-mate/internal/repository"
)

type recipientRepo struct{ db *sql.DB }

func NewRecipientRepo(db *sql.DB) repository.RecipientRepository {
	return &recipientRepo{db: db}
}

func (r *recipientRepo) BulkCreate(ctx context.Context, recipients []domain.CampaignRecipient) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO campaign_recipients (campaign_id, recipient_name, recipient_email, company, role, custom_variables)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, rec := range recipients {
		vars, _ := json.Marshal(rec.CustomVariables)
		if _, err := stmt.ExecContext(ctx, rec.CampaignID, rec.RecipientName, rec.RecipientEmail, rec.Company, rec.Role, vars); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *recipientRepo) ListByCampaignID(ctx context.Context, campaignID int64) ([]domain.CampaignRecipient, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, campaign_id, recipient_name, recipient_email, company, role, custom_variables, created_at
		 FROM campaign_recipients WHERE campaign_id=$1`, campaignID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recipients []domain.CampaignRecipient
	for rows.Next() {
		var rec domain.CampaignRecipient
		var rawVars []byte
		if err := rows.Scan(&rec.ID, &rec.CampaignID, &rec.RecipientName, &rec.RecipientEmail,
			&rec.Company, &rec.Role, &rawVars, &rec.CreatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal(rawVars, &rec.CustomVariables)
		recipients = append(recipients, rec)
	}
	return recipients, rows.Err()
}
