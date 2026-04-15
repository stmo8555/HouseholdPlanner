package login

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
)

type Repo struct {
	Sessions map[string]Session
	DB       *pgx.Conn
}

func (r *Repo) AddSession(uuid string, session Session) {
	r.Sessions[uuid] = session
}

func (r *Repo) RemoveSession(uuid string) {
	delete(r.Sessions, uuid)
}

func (r *Repo) SelectUser(ctx context.Context, uname string) (User, error) {
	sql := "SELECT id, pwd FROM users WHERE username=$1"

	var uid int
	var hash string

	err := r.DB.QueryRow(ctx, sql, uname).Scan(&uid, &hash)

	if err != nil {
		if err == pgx.ErrNoRows {
			return User{}, nil
		} else {
			return User{}, nil
		}
	}

	return User{ID: uid, Uname: uname, Hash: hash}, err
}

func (r *Repo) getHouseholdId(user_id int) (int, error) {
	sql := `select household_id FROM household_members where user_id=$1`
	var hid int
	err := r.DB.QueryRow(context.Background(), sql, user_id).Scan(&hid)

	return hid, err
}

func (r *Repo) getSession(sessionID string) (Session, error) {
	session, ok := r.Sessions[sessionID]

	if !ok {
		return Session{}, errors.New("Missing session")
	}

	return session, nil
}
