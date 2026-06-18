package product

import (
	"context"
	"errors"
	"strings"
	"unicode"

	"github.com/jackc/pgx/v5"
	"github.com/stmo8555/HouseholdPlanner/internal/ai"
)

type Service struct {
	Repo           *Repo
	AIService      *ai.Service
	FoodCategories map[string]string
}

func NewService(repo *Repo, ai *ai.Service, lookUpTable map[string]string) *Service {
	if repo == nil || ai == nil || lookUpTable == nil {
		panic("service not initialized")
	}
	normalizedLookUpTable := make(map[string]string, len(lookUpTable))
	for key, category := range lookUpTable {
		normalizedLookUpTable[normalizeLookupKey(key)] = category
	}

	return &Service{
		Repo:           repo,
		AIService:      ai,
		FoodCategories: normalizedLookUpTable,
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

	p.Category = s.categoryFor(p.Name)

	return s.Repo.Add(ctx, p)
}

func (s *Service) categoryFor(name string) string {
	key := normalizeLookupKey(name)
	if cat, ok := s.FoodCategories[key]; ok {
		return cat
	}

	fields := strings.Fields(key)
	for _, field := range fields {
		if cat, ok := s.FoodCategories[field]; ok && cat == "frozen" {
			return cat
		}
	}

	for size := len(fields) - 1; size >= 1; size-- {
		for i := 0; i+size <= len(fields); i++ {
			phrase := strings.Join(fields[i:i+size], " ")
			if cat, ok := s.FoodCategories[phrase]; ok {
				return cat
			}
		}
	}

	return "other"
}

func normalizeLookupKey(s string) string {
	var b strings.Builder
	lastWasSpace := true
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			lastWasSpace = false
			continue
		}

		if !lastWasSpace {
			b.WriteByte(' ')
			lastWasSpace = true
		}
	}

	return strings.Join(strings.Fields(b.String()), " ")
}
