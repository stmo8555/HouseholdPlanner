package household

import (
	"context"
	"time"
)

type Service struct {
	repo *Repo
}

func NewService(r *Repo) *Service {
	if r == nil {
		panic("nil service for handler")
	}

	return &Service{repo: r}
}

func (s *Service) RemoveMember(ctx context.Context, id int, hid int) error {
	return s.repo.RemoveMember(ctx, id, hid)
}

func (s *Service) GenerateInviteToken(ctx context.Context, uid, hid int) (string, error) {
	expiresAt := time.Now().Add(3 * 24 * time.Hour).UTC()
	return s.repo.CreateInvite(ctx, hid, uid, expiresAt)
}

func (s *Service) Settings(context context.Context, hid int) {
	panic("unimplemented")
}

func (s *Service) HouseholdVersion(ctx context.Context, householdID int) (int, error) {
	return s.repo.HouseholdVersion(ctx, householdID)
}
