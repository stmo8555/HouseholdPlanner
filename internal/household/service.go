package household

import "context"

type Service struct {
	repo *Repo
}

func NewService(repo *Repo) *Service {
	if repo == nil {
		panic("nil repo for service")
	}

	return &Service{repo: repo}
}

func (s *Service) HouseholdVersion(ctx context.Context, householdID int) (int, error) {
	return s.repo.HouseholdVersion(ctx, householdID)
}
