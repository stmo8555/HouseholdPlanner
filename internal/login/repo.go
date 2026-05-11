package login

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	db *pgxpool.Pool
}

func CreateRepo(db *pgxpool.Pool) *Repo {
	if db == nil {
		panic("nil DB connection or nil session map")
	}

	return &Repo{
		db: db,
	}
}

func (r *Repo) AddSession(ctx context.Context, session Session) (string, error) {
	sql := `
	INSERT INTO sessions (user_id, expires_at)
	VALUES ($1, $2)
	RETURNING id 
	`
	var id string
	err := r.db.QueryRow(ctx, sql, session.UserID, session.ExpiresAt).Scan(&id)

	return id, err
}

func (r *Repo) RemoveSession(ctx context.Context, id string) error {
	sql := `
	DELETE FROM sessions
	WHERE id=$1`
	_, err := r.db.Exec(ctx, sql, id)

	return err
}

func (r *Repo) User(ctx context.Context, uname string) (User, error) {
	sql := `
	SELECT id, pwd 
	FROM users
	WHERE username=$1
	`

	var uid int
	var hash string

	err := r.db.QueryRow(ctx, sql, uname).Scan(&uid, &hash)

	return User{
		ID:    uid,
		Uname: uname,
		Hash:  hash,
	}, err
}

func (r *Repo) getHouseholdId(user_id int) (int, error) {
	sql := `
	SELECT household_id
	FROM household_members 
	WHERE user_id=$1
	`
	var hid int
	err := r.db.QueryRow(context.Background(), sql, user_id).Scan(&hid)

	return hid, err
}

func (r *Repo) getSession(ctx context.Context, id string) (Session, error) {
	sql := `
	SELECT s.id, s.user_id, s.expires_at, hm.household_id
	FROM sessions s 
	LEFT JOIN household_members hm ON s.user_id=hm.user_id 
	WHERE s.id=$1;
	`
	var session Session
	err := r.db.QueryRow(ctx, sql, id).Scan(
		&session.ID,
		&session.UserID,
		&session.ExpiresAt,
		&session.HouseholdID,
	)

	return session, err
}

func (r *Repo) RemoveExpiredSessions(ctx context.Context) error {
	sql := `
	DELETE FROM sessions
	WHERE NOW() >= expires_at;
	`
	_, err := r.db.Exec(ctx, sql)

	return err
}
