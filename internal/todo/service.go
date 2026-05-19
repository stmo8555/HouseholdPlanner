package todo

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/stmo8555/HouseholdPlanner/internal/ai"
)

type Service struct {
	Repo      *Repo
	AIService *ai.Service
}

func CreateService(repo *Repo, ai *ai.Service) *Service {
	if repo == nil || ai == nil {
		panic("service not initialized")
	}

	return &Service{
		Repo:      repo,
		AIService: ai,
	}
}

func (s *Service) AddTodo(ctx context.Context, todo Todo) error {
	task := strings.TrimSpace(todo.Task)

	if task == "" {
		return errors.New("Task have no name")
	}

	todo.Normalize()

	_, err := s.Repo.Add(ctx, todo)

	return err
}

func (s *Service) Count(ctx context.Context, hid int) (int, error) {
	return s.Repo.Count(ctx, hid)
}

func (s *Service) MarkDone(ctx context.Context, id, hid int) error {
	todo, err := s.Repo.MarkDone(ctx, id, hid, time.Now().UTC())
	if err != nil {
		return err
	}

	return s.schedule(ctx, []Todo{todo})
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

func (s *Service) ScheduleRepeats(ctx context.Context) error {
	now := time.Now().UTC()

	endOfTomorrow := time.Date(
		now.Year(),
		now.Month(),
		now.Day()+2,
		0, 0, 0, 0,
		time.UTC,
	)

	todos, err := s.Repo.ListSchedulableDueBefore(ctx, endOfTomorrow)

	err = s.schedule(ctx, todos)

	if err != nil {
		return err
	}

	return nil
}

func (s *Service) schedule(ctx context.Context, todos []Todo) error {
	for _, v := range todos {
		var newDue = v.Due
		switch v.Repeat {
		case RepeatNever:
			continue
		case RepeatDaily:
			newDue.Time = newDue.Time.AddDate(0, 0, v.Frequency)
		case RepeatWeekly:
			newDue.Time = newDue.Time.AddDate(0, 0, 7*v.Frequency)
		case RepeatMonthly:
			newDue.Time = newDue.Time.AddDate(0, v.Frequency, 0)
		case RepeatYearly:
			newDue.Time = newDue.Time.AddDate(v.Frequency, 0, 0)
		default:
			panic("WHY ARE WE HERE!")
		}

		newTodo := v
		newTodo.Due = newDue
		nextID, err := s.Repo.Add(ctx, newTodo)

		if err != nil {
			return err
		}

		err = s.Repo.updateNextID(ctx, nextID, v.Id)

		if err != nil {
			return err
		}
	}

	return nil
}
