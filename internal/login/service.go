package login

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

const fakeHash = "$2a$10$7EqJtq98hPqEX7fNZaFWoOePaWxn96p36C1p0uZ1tcHTTX3e8DqGa"

var errInvalidCredentials = errors.New("invalid username or password")

type Service struct {
	repo *Repo
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
		if errors.Is(err, pgx.ErrNoRows) {
			return "", errInvalidCredentials
		}

		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Hash), []byte(pwd))
	if err != nil {
		return "", errInvalidCredentials
	}

	session := Session{
		UserID:    user.ID,
		ExpiresAt: time.Now().UTC().Add(time.Hour * 36),
	}

	return s.repo.AddSession(ctx, session)
}

func (s *Service) GetSession(ctx context.Context, sessionID string) (Session, error) {
	return s.repo.getSession(ctx, sessionID)
}

func (s *Service) RemoveExpiredSessions(ctx context.Context) error {
	return s.repo.RemoveExpiredSessions(ctx)
}
