package login

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/stmo8555/HouseholdPlanner/internal/code"
)

type Repo struct {
	db *pgxpool.Pool
}

func (r *Repo) ConsumeToken(ctx context.Context, token string) error {
	sql := `
	DELETE
	FROM invites
	WHERE token=$1;
	`
	row, err := r.db.Exec(ctx, sql, token)
	if err != nil {
		return err
	}

	if row.RowsAffected() == 0 {
		return fmt.Errorf("token not found")
	}

	return nil
}

func (r *Repo) ValidateToken(ctx context.Context, token string) (bool, error) {
	sql := `
	SELECT
	EXISTS (
		SELECT 1 
		FROM invites 
		WHERE token = $1 AND expires_at > NOW() 
	);
	`
	var exists bool
	err := r.db.QueryRow(ctx, sql, token).Scan(&exists)

	return exists, err
}

func NewRepo(db *pgxpool.Pool) *Repo {
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
	RETURNING id;
	`
	var id string
	err := r.db.QueryRow(ctx, sql, session.User.ID, session.ExpiresAt).Scan(&id)

	return id, err
}

func (r *Repo) RemoveSession(ctx context.Context, id string) error {
	sql := `
	DELETE FROM sessions
	WHERE id=$1;
	`
	_, err := r.db.Exec(ctx, sql, id)

	return err
}

func (r *Repo) ExtendSession(ctx context.Context, id string, newExpiry time.Time) error {
	sql := `
	UPDATE sessions
	SET expires_at=$1
	WHERE id=$2;
	`
	_, err := r.db.Exec(ctx, sql, newExpiry, id)

	return err
}

func (r *Repo) User(ctx context.Context, uname string) (User, error) {
	sql := `
	SELECT id, pwd, is_admin 
	FROM users
	WHERE username=$1
	`

	var user User

	err := r.db.QueryRow(ctx, sql, uname).Scan(
		&user.ID,
		&user.Hash,
		&user.IsAdmin,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		err = ErrNotFound
	}

	return user, err
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
	SELECT s.id, s.expires_at, s.created_at, hm.household_id, u.id, u.username, u.pwd
	FROM sessions s
	JOIN users u ON s.user_id=u.id
	LEFT JOIN household_members hm ON s.user_id=hm.user_id
	WHERE s.id=$1;
	`
	var session Session
	err := r.db.QueryRow(ctx, sql, id).Scan(
		&session.ID,
		&session.ExpiresAt,
		&session.CreatedAt,
		&session.HouseholdID,
		&session.User.ID,
		&session.User.Uname,
		&session.User.Hash,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		err = ErrNotFound
	}

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

func (r *Repo) CreateUser(ctx context.Context, username string, hash string) (int, error) {
	sql := `
	INSERT INTO users (username, pwd)
	VALUES ($1, $2)
	RETURNING id;
	`
	var id int
	err := r.db.QueryRow(ctx, sql, username, hash).Scan(&id)

	return id, err
}

func (r *Repo) CreateHousehold(ctx context.Context, householdName string, userId int) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	sql := `
	INSERT INTO households (name, code, created_by)
	VALUES ($1, $2, $3)
	RETURNING id;
	`

	householdCode, err := code.Generate()
	if err != nil {
		return fmt.Errorf("generating household code: %w", err)
	}

	var hid int
	err = tx.QueryRow(ctx, sql, householdName, householdCode, userId).Scan(&hid)
	if err != nil {
		return fmt.Errorf("inserting household: %w", err)
	}

	sql = `
	INSERT INTO household_members (user_id, household_id, role)
	VALUES ($1, $2, 'owner');
	`
	_, err = tx.Exec(ctx, sql, userId, hid)
	if err != nil {
		return fmt.Errorf("adding owner to household: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *Repo) JoinHousehold(ctx context.Context, inviteCode string, userId int) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	sql := `
	SELECT id FROM households WHERE code = $1;
	`
	var hid int
	err = tx.QueryRow(ctx, sql, inviteCode).Scan(&hid)
	if err != nil {
		return fmt.Errorf("household not found for code %q: %w", inviteCode, err)
	}

	sql = `
	INSERT INTO household_members (user_id, household_id, role)
	VALUES ($1, $2, 'member');
	`
	_, err = tx.Exec(ctx, sql, userId, hid)
	if err != nil {
		return fmt.Errorf("joining household: %w", err)
	}

	return tx.Commit(ctx)
}




