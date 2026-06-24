package household

import (
	"context"
	"time"

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

func (r *Repo) RemoveMember(ctx context.Context, id int, hid int) error {
	sql := `
	DELETE
	FROM household_members
	WHERE user_id=$1 and household_id=$2 
	`
	_, err := r.db.Exec(ctx, sql, id, hid)

	return err
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

func (r *Repo) CreateInvite(ctx context.Context, householdID, createdBy int, expiresAt time.Time) (string, error) {
	sql := `
	INSERT INTO invites (household_id, created_by, expires_at)
	VALUES ($1, $2, $3)
	RETURNING token;
	`

	var token string
	err := r.db.QueryRow(ctx, sql, householdID, createdBy, expiresAt).Scan(&token)

	return token, err
}
