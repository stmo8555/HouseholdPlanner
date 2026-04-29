package todo

import (
	"context"
	"github.com/jackc/pgx/v5"
	"time"
)

type Repo struct {
	DB *pgx.Conn
}

func (r *Repo) Add(ctx context.Context, t Todo) error {
	_, err := r.DB.Exec(ctx,
		`INSERT INTO todos (task, due, repeat, frequency, household_id)
		VALUES ($1, $2, $3, $4, $5)`,
		t.Task,
		t.Due,
		t.Repeat,
		t.Frequency,
		t.HouseholdID,
	)
	return err
}

func (r *Repo) Count(ctx context.Context, hid int) (int, error) {
	sql := `
        SELECT COUNT (*)
        FROM todos
        WHERE household_id = $1;`

	var count int
	err := r.DB.QueryRow(ctx, sql, hid).Scan(&count)

	return count, err
}

func (r *Repo) MarkDone(ctx context.Context, id, hid int, t time.Time) error {
	query := `UPDATE todos SET completed_at=$1 WHERE id=$2 AND household_id=$3`
	_, err := r.DB.Exec(ctx, query, t, id, hid)

	return err
}

func (r *Repo) MarkUnDone(ctx context.Context, id, hid int) error {
	query := `UPDATE todos SET completed_at=NULL WHERE id=$1 AND household_id=$2`
	_, err := r.DB.Exec(ctx, query, id, hid)

	return err
}

func (r *Repo) RemoveCompletedOlderThan(ctx context.Context, cutoff time.Time) error {
	query := `
		DELETE FROM todos
		WHERE completed_at < $1
	`

	_, err := r.DB.Exec(ctx, query, cutoff)
	return err
}

func (r *Repo) List(ctx context.Context, hid int) ([]Todo, error) {
	sql := `SELECT id, task, due, repeat, frequency, completed_at, household_id
        FROM todos WHERE household_id = $1;`

	rows, err := r.DB.Query(ctx, sql, hid)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []Todo

	for rows.Next() {
		var t Todo
		err := rows.Scan(
			&t.Id,
			&t.Task,
			&t.Due,
			&t.Repeat,
			&t.Frequency,
			&t.CompletedAt,
			&t.HouseholdID,
		)

		if err != nil {
			return nil, err
		}

		todos = append(todos, t)
	}

	return todos, err
}
