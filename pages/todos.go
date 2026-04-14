package pages

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type Todos struct {
	Id           int
	Title        string
	Household_id int
	Completed_at sql.NullTime
}

func AmountOfTodos(conn *pgx.Conn, hid int) (int, error) {
	sql := `
        SELECT COUNT (*)
        FROM todos
        WHERE household_id = $1;`

	var count int

	err := conn.QueryRow(context.Background(), sql, hid).Scan(&count)
	if err != nil {
		return 0, err
	}
	
	return count, nil
}
func Done(c *gin.Context, conn *pgx.Conn) {
	hid, ok := c.Get("household_id")

	id := c.PostForm("id")

	if !ok {
		panic("failed to get household id from context")
	}
	query := `UPDATE todos SET completed_at=$1 WHERE id=$2 AND household_id=$3`

	completed := sql.NullTime{
		Time:  time.Now().UTC(),
		Valid: true,
	}

	_, err := conn.Exec(context.Background(), query, completed, id, hid)

	if err != nil {
		panic(err)
	}

	c.Redirect(302, "/todos")
}

func Undo(c *gin.Context, conn *pgx.Conn) {
	hid, ok := c.Get("household_id")

	id := c.PostForm("id")
	if !ok {
		panic("failed to get household id from context")
	}
	query := `UPDATE todos SET completed_at=$1 WHERE id=$2 AND household_id=$3`

	_, err := conn.Exec(context.Background(), query, nil, id, hid)

	if err != nil {
		panic(err)
	}

	c.Redirect(302, "/todos")
}

func Add(c *gin.Context, conn *pgx.Conn) {
	hid, ok := c.Get("household_id")
	todo := c.PostForm("todo")

	if !ok {
		panic("failed to get household id from context")
	}

	sql := `INSERT INTO todos (title, household_id) VALUES ($1, $2);`
	_, err := conn.Exec(context.Background(), sql, todo, hid)

	if err != nil {
		panic(err)
	}

	c.Redirect(302, "/todos")
}

func RemoveTodos(conn *pgx.Conn) {
	sql := `DELETE FROM todos WHERE completed_at < NOW() - INTERVAL '7 days';`
	_, err := conn.Exec(context.Background(), sql)

	if err != nil {
		panic(err)
	}
}

func List(c *gin.Context, conn *pgx.Conn) {
	hid, ok := c.Get("household_id")

	if !ok {
		panic("failed to get household id from context")
	}

	sql := `SELECT id, title, household_id, completed_at 
        FROM todos WHERE household_id = $1;`

	rows, err := conn.Query(context.Background(), sql, hid)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var todos []Todos
	var completed []Todos

	for rows.Next() {
		var t Todos
		err := rows.Scan(
			&t.Id,
			&t.Title,
			&t.Household_id,
			&t.Completed_at,
		)

		if err != nil {
			panic(err)
		}

		if t.Completed_at.Valid {
			completed = append(completed, t)
		} else {
			todos = append(todos, t)
		}
	}

	if rows.Err() != nil {
		panic(rows.Err())
	}

	data := gin.H{
		"Title":       "Groceries",
		"CurrentPath": c.Request.URL.Path,
		"Todos":       todos,
		"Completed":   completed,
	}
	c.HTML(http.StatusOK, "todos.html", data)
}
