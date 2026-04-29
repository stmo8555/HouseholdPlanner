package todo

import (
	"database/sql"
	"time"
)

type Todo struct {
	Id          int
	Task        string       `json:"task"`
	Due         sql.NullTime `json:"due"`
	Repeat      string       `json:"repeat"`
	Frequency   int          `json:"frequency"`
	CompletedAt sql.NullTime
	HouseholdID int
}

func (t *Todo) DueIn() int {
	if !t.Due.Valid {
		return 0
	}

	diff := time.Until(t.Due.Time)

	if diff < 0 {
		diff = -diff
	}

	return int(diff.Hours() / 24)
}

type TodosCategorized struct {
	Overdue   []Todo
	Today     []Todo
	Soon      []Todo
	Completed []Todo
	TheRest   []Todo
}
