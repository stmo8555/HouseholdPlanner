package product

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/stmo8555/HouseholdPlanner/internal/ai"
)

type Service struct {
	Repo           *Repo
	AIService      *ai.Service
	FoodCategories map[string]string
}

func CreateService(repo *Repo, ai *ai.Service, lookUpTable map[string]string) *Service {
	if repo == nil || ai == nil || lookUpTable == nil {
		panic("service not initialized")
	}

	return &Service{
		Repo:           repo,
		AIService:      ai,
		FoodCategories: lookUpTable,
	}
}
func (s *Service) Get(ctx context.Context, id int) (Product, error) {
	return s.Repo.Get(ctx, id)
}

func (s *Service) GetID(ctx context.Context, p Product) (int, error) {
	p.Normalize()

	if p.Name == "" {
		return 0, errors.New("Product is missing required field \"name\"")
	}

	id, err := s.Repo.GetID(ctx, p)

	found := err != pgx.ErrNoRows
	if found {
		return id, err
	}

	key := strings.ToLower(p.Name)

	category := "other"

	if cat, ok := s.FoodCategories[key]; ok {
		category = cat
	} else {
		for token := range strings.FieldsSeq(key) {
			if cat, ok := s.FoodCategories[token]; ok {
				category = cat
				break
			}
		}
	}

	p.Category = category

	return s.Repo.Add(ctx, p)
}
