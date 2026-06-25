package household

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/stmo8555/HouseholdPlanner/internal/code"
)

// ErrOwnerMustTransfer is returned when an owner tries to leave or delete their
// account while other members remain — they must transfer ownership first.
var ErrOwnerMustTransfer = errors.New("transfer ownership before leaving")

// ErrHouseholdNotEmpty is returned when an owner tries to delete a household that
// still has other members.
var ErrHouseholdNotEmpty = errors.New("household still has members")

type Service struct {
	repo *Repo
}

func NewService(r *Repo) *Service {
	if r == nil {
		panic("nil service for handler")
	}

	return &Service{repo: r}
}

func (s *Service) RegenerateHouseholdCode(ctx context.Context, hid int) (string, error) {
	codeStr, err := code.Generate()
	if err != nil {
		return "", fmt.Errorf("generating household code: %w", err)
	}

	err = s.repo.RegenerateHouseholdCode(ctx, codeStr, hid)
	return codeStr, err
}

func (s *Service) RemoveMember(ctx context.Context, userID, targetID, hid int) error {
	return s.repo.RemoveMember(ctx, userID, targetID, hid)
}

func (s *Service) GenerateInviteToken(ctx context.Context, uid, hid int) (string, error) {
	expiresAt := time.Now().Add(3 * 24 * time.Hour).UTC()
	return s.repo.CreateInvite(ctx, hid, uid, expiresAt)
}

func (s *Service) CurrentInviteToken(ctx context.Context, uid, hid int) (string, error) {
	token, err := s.repo.CurrentInvite(ctx, hid)
	if err == nil {
		return token, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return "", fmt.Errorf("fetching current invite: %w", err)
	}
	return s.GenerateInviteToken(ctx, uid, hid)
}

func (s *Service) PromoteMember(ctx context.Context, uid int, targetUID int, hid int) error {
	return s.repo.PromoteMember(ctx, uid, targetUID, hid)
}

func (s *Service) Settings(ctx context.Context, uid, hid int) (SettingsView, error) {
	household, err := s.repo.Household(ctx, hid)
	if err != nil {
		return SettingsView{}, fmt.Errorf("fetching household: %w", err)
	}

	members, err := s.repo.HouseholdMembers(ctx, hid)
	if err != nil {
		return SettingsView{}, fmt.Errorf("fetching members: %w", err)
	}

	var callerIsOwner bool
	for i := range members {
		if members[i].ID == uid {
			members[i].IsYou = true
			callerIsOwner = members[i].IsOwner
		}
	}

	return SettingsView{
		Household: household,
		Members:   members,
		IsOwner:   callerIsOwner,
	}, nil
}

func (s *Service) LeaveHousehold(ctx context.Context, uid, hid int) error {
	return s.repo.LeaveHousehold(ctx, uid, hid)
}

func (s *Service) DeleteHousehold(ctx context.Context, uid, hid int) error {
	return s.repo.DeleteHousehold(ctx, uid, hid)
}

func (s *Service) DeleteAccount(ctx context.Context, uid int) error {
	return s.repo.DeleteAccount(ctx, uid)
}

func (s *Service) HouseholdVersion(ctx context.Context, householdID int) (int, error) {
	return s.repo.HouseholdVersion(ctx, householdID)
}


