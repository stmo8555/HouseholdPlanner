package todos

import "database/sql"

type Todo struct {
	Id           int
	Title        string
	Household_id int
	Completed_at sql.NullTime
}

type TodoList struct {
	Active, Completed []Todo
}
