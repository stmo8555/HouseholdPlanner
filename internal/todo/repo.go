package todo

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	DB *pgxpool.Pool
}

func CreateRepo(db *pgxpool.Pool) *Repo {
	if db == nil {
		panic("nil DB connection")
	}

	return &Repo{
		DB: db,
	}
}

func (r *Repo) Add(ctx context.Context, t Todo) (int, error) {
	sql := `
	INSERT INTO todos (task, due, repeat, frequency, household_id)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id;
	`
	var id int
	err := r.DB.QueryRow(ctx, sql, t.Task, t.Due, t.Repeat, t.Frequency, t.HouseholdID).Scan(&id)
	return id, err
}

func (r *Repo) updateNextID(ctx context.Context, nextID, todoID int) error {
	query := `
	UPDATE todos 
	SET next_id=$1 
	WHERE id=$2;`
	_, err := r.DB.Exec(ctx, query, nextID, todoID)

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

func (r *Repo) MarkDone(ctx context.Context, id, hid int, t time.Time) (Todo, error) {
	sql := `
	UPDATE todos 
	SET completed_at=$1 
	WHERE id=$2 AND household_id=$3
	RETURNING id, task, due, repeat, frequency, next_id, completed_at, household_id;
	`

	var todo Todo
	err := r.DB.QueryRow(ctx, sql, t, id, hid).Scan(
		&todo.Id,
		&todo.Task,
		&todo.Due,
		&todo.Repeat,
		&todo.Frequency,
		&todo.NextID,
		&todo.CompletedAt,
		&todo.HouseholdID,
	)

	return todo, err
}

func (r *Repo) MarkUnDone(ctx context.Context, id, hid int) error {
	query := `UPDATE todos SET completed_at=NULL WHERE id=$1 AND household_id=$2`
	_, err := r.DB.Exec(ctx, query, id, hid)

	return err
}

func (r *Repo) RemoveCompletedOlderThan(ctx context.Context, cutoff time.Time) error {
	query := `
		DELETE FROM todos
		WHERE completed_at < $1;
	`

	_, err := r.DB.Exec(ctx, query, cutoff)
	return err
}

func (r *Repo) List(ctx context.Context, hid int) ([]Todo, error) {
	sql := `
	SELECT id, task, due, repeat, frequency, next_id, completed_at, household_id
    FROM todos
	WHERE household_id = $1
	ORDER BY due ASC;
	`

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
			&t.NextID,
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

func (r *Repo) ListSchedulableDueBefore(ctx context.Context, before time.Time) ([]Todo, error) {
	sql := `
	SELECT id, task, due, repeat, frequency, next_id, completed_at, household_id
	FROM todos
	WHERE due <= $1 AND next_id IS NULL
	ORDER BY due ASC;
	`

	rows, err := r.DB.Query(ctx, sql, before)

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
			&t.NextID,
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
