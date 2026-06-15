package postgres

import (
	"database/sql"

	_ "github.com/lib/pq"
	"go-template/internal/db"
)

func NewDB(databaseURL string) (*db.Queries, error) {
	sqlDB, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}
	return db.New(sqlDB), nil
}
