package login

import (
	"context"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	Repo *Repo
}

func (s *Service) Logout(uuid string) {
	s.Repo.RemoveSession(uuid)
}

func (s *Service) Authenticate(ctx context.Context, uname, pwd string) string {
	user, err := s.Repo.SelectUser(ctx, uname)

	if err != nil {
		return ""
	}

	if !verifyPassword(pwd, user.Hash) {
		return ""
	}

	sessionID := uuid.New().String()
	session := Session{
		UserID:      user.ID,
		HouseholdID: nil,
	}

	hid, err := s.Repo.getHouseholdId(user.ID)

	if err == nil {
		session.HouseholdID = &hid
	} else {
		panic("not implemented yet")
	}

	s.Repo.AddSession(sessionID, session)
	return sessionID
}

func verifyPassword(pwd, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pwd))

	if err != nil {
		return false
	}

	return true
}

func (s *Service) GetSession(sessionID string) (Session, error) {
	return s.Repo.getSession(sessionID)
}
