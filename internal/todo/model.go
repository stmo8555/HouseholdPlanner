package todo

import (
	"database/sql"
	"fmt"
	"strings"
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

func (t *Todo) RepeatLabel() string {
	var occurence string
	switch strings.ToLower(t.Repeat) {
	case "never":
		return ""
	case "daily":
		occurence = "days"
	case "weekly":
		occurence = "weeks"
	case "yearly":
		occurence = "years"
	default:
		panic("We should never get to default. Our code is broken")
	}
	if t.Frequency == 1 {

		return fmt.Sprintf("⟳  %v", t.Repeat)
	}

	return fmt.Sprintf("⟳  Every %v %v", t.Frequency, occurence)
}

func (t *Todo) DueLabel() string {
	if !t.Due.Valid {
		return "No due date"
	}

	now := time.Now()

	// Extract only date parts
	y1, m1, d1 := now.Date()
	y2, m2, d2 := t.Due.Time.Date()

	today := time.Date(y1, m1, d1, 0, 0, 0, 0, now.Location())
	due := time.Date(y2, m2, d2, 0, 0, 0, 0, now.Location())

	days := int(due.Sub(today).Hours() / 24)

	switch days {
	case 0:
		return "Due today"
	case 1:
		return "Due tomorrow"
	case -1:
		return "Due yesterday"
	default:
		if days > 0 {
			return fmt.Sprintf("Due in %d days", days)
		}
		return fmt.Sprintf("Overdue by %d days", -days)
	}
}

type TodosCategorized struct {
	Overdue   []Todo
	Today     []Todo
	Soon      []Todo
	Completed []Todo
	TheRest   []Todo
}
