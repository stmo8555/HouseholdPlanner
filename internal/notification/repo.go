package notification

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	db *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) *Repo {
	if db == nil {
		panic("nil DB connection")
	}

	return &Repo{db: db}
}

func (r *Repo) HouseholdVersion(ctx context.Context, householdID int) (int, error) {
	sql := `
	SELECT version
	FROM households
	WHERE id = $1;
	`

	var version int
	err := r.db.QueryRow(ctx, sql, householdID).Scan(&version)

	return version, err
}
