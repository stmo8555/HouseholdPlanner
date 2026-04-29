package todo

import (
	"context"
	"errors"
	"strings"
	"time"
)

type Service struct {
	Repo *Repo
}

func (s *Service) AddTodo(ctx context.Context, todo Todo) error {
	task := strings.TrimSpace(todo.Task)

	if task == "" {
		return errors.New("Task have no name")
	}

	todo.Task = strings.ToUpper(task[:1]) + strings.ToLower(task[1:])

	return s.Repo.Add(ctx, todo)
}

func (s *Service) Count(ctx context.Context, hid int) (int, error) {
	return s.Repo.Count(ctx, hid)
}

func (s *Service) MarkDone(ctx context.Context, id, hid int) error {
	return s.Repo.MarkDone(ctx, id, hid, time.Now().UTC())
}

func (s *Service) MarkUnDone(ctx context.Context, id, hid int) error {
	return s.Repo.MarkUnDone(ctx, id, hid)
}

func (s *Service) RemoveOldCompleted(ctx context.Context) error {
	cutoff := time.Now().UTC().Add(-7 * 24 * time.Hour)
	return s.Repo.RemoveCompletedOlderThan(ctx, cutoff)
}

func (s *Service) List(ctx context.Context, hid int) (TodosCategorized, error) {
	todos, err := s.Repo.List(ctx, hid)

	if err != nil {
		return TodosCategorized{}, err
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	soon := today.AddDate(0, 0, 4)

	var todosCategorized TodosCategorized

	for _, t := range todos {
		due := t.Due.Time

		dueDate := time.Date(
			due.Year(), due.Month(), due.Day(),
			0, 0, 0, 0,
			today.Location(),
		)

		switch {
		case t.CompletedAt.Valid:
			todosCategorized.Completed = append(todosCategorized.Completed, t)

		case !t.Due.Valid:
			todosCategorized.TheRest = append(todosCategorized.TheRest, t)

		case dueDate.Before(today):
			todosCategorized.Overdue = append(todosCategorized.Overdue, t)

		case dueDate.Equal(today):
			todosCategorized.Today = append(todosCategorized.Today, t)

		case dueDate.After(today) && !dueDate.After(soon):
			todosCategorized.Soon = append(todosCategorized.Soon, t)

		default:
			todosCategorized.TheRest = append(todosCategorized.TheRest, t)
		}
	}

	return todosCategorized, nil
}
