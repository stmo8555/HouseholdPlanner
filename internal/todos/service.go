package todos

import (
	"context"
	"time"
)

type Service struct {
	Repo *Repo
}

func (s *Service) AddTodo(ctx context.Context, title string, hid int) error {
	return s.Repo.Add(ctx, title, hid)
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

func (s *Service) List(ctx context.Context, hid int) (TodoList, error) {
	todos, err := s.Repo.List(ctx, hid)

	if err != nil {
		return TodoList{}, err
	}

	var todolist TodoList

	for _, t := range todos {
		if t.Completed_at.Valid {
			todolist.Completed = append(todolist.Completed, t)
		} else {
			todolist.Active = append(todolist.Active, t)
		}
	}

	return todolist, nil
}
