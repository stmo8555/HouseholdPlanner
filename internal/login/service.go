package login

import (
	"context"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const fakeHash = "$2a$10$7EqJtq98hPqEX7fNZaFWoOePaWxn96p36C1p0uZ1tcHTTX3e8DqGa"

var errInvalidCredentials = errors.New("invalid username or password")

type Service struct {
	repo *Repo
}

func (s *Service) JoinHousehold(ctx context.Context, inviteCode string, userId int) error {
	return s.repo.JoinHousehold(ctx, inviteCode, userId)
}

func (s *Service) CreateHousehold(ctx context.Context, householdName string, userId int) error {
	return s.repo.CreateHousehold(ctx, householdName, userId)
}

func (s *Service) CreateUser(ctx context.Context, username, pwd, token string) (int, error) {
	err := s.repo.ConsumeToken(ctx, token)

	if err != nil {
		return 0, err
	}

	hashBytes, err := bcrypt.GenerateFromPassword([]byte(pwd), 12)

	if err != nil {
		return 0, err
	}

	return s.repo.CreateUser(ctx, username, string(hashBytes))
}

func (s *Service) ValidateToken(ctx context.Context, token string) (bool, error) {
	return s.repo.ValidateToken(ctx, token)
}

func NewService(repo *Repo) *Service {
	if repo == nil {
		panic("service not initialized")
	}

	return &Service{
		repo: repo,
	}
}

func (s *Service) Logout(ctx context.Context, uuid string) {
	err := s.repo.RemoveSession(ctx, uuid)
	if err != nil {
		panic(err)
	}
}

func (s *Service) Authenticate(ctx context.Context, uname, pwd string) (string, error) {
	user, err := s.repo.User(ctx, uname)

	if err != nil {
		bcrypt.CompareHashAndPassword([]byte(fakeHash), []byte(pwd))
		if errors.Is(err, ErrNotFound) {
			return "", errInvalidCredentials
		}

		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Hash), []byte(pwd))
	if err != nil {
		return "", errInvalidCredentials
	}

	session := Session{
		User:      user,
		ExpiresAt: time.Now().UTC().Add(time.Hour * 36),
	}

	return s.repo.AddSession(ctx, session)
}

func (s *Service) GetSession(ctx context.Context, sessionID string) (Session, error) {
	return s.repo.getSession(ctx, sessionID)
}

func (s *Service) TouchLastSeen(ctx context.Context, userID int) error {
	return s.repo.TouchLastSeen(ctx, userID)
}

func (s *Service) RemoveExpiredSessions(ctx context.Context) error {
	return s.repo.RemoveExpiredSessions(ctx)
}
