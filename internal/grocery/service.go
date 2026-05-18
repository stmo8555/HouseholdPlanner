package grocery

import (
	"context"

	"github.com/stmo8555/HouseholdPlanner/internal/ingredient"
	"github.com/stmo8555/HouseholdPlanner/internal/product"
)

type Service struct {
	repo                IRepo
	productService      *product.Service
	ingredientExtractor *ingredient.Extractor
}

func CreateService(repo IRepo, product *product.Service, ingredient *ingredient.Extractor) *Service {
	if repo == nil || product == nil || ingredient == nil {
		panic("service not initialized")
	}

	return &Service{
		repo:                repo,
		productService:      product,
		ingredientExtractor: ingredient,
	}
}

func (s *Service) GetTopProducts(ctx context.Context, householdID int) ([]string, error) {
	products, err := s.repo.getTopProducts(ctx, householdID)

	if err != nil {
		panic(err)
	}

	strSlice := make([]string, 0, len(products))

	for _, v := range products {
		strSlice = append(strSlice, v.Name)
	}

	return strSlice, err
}

func (s *Service) AddGroceries(ctx context.Context, ingredients []ingredient.Ingredient, hid int) error {
	groceries := make([]Grocery, 0, len(ingredients))
	for i := range ingredients {
		if ingredients[i].ProductID == 0 {
			id, err := s.productService.GetID(ctx, ingredients[i].Product)

			if err != nil {
				panic(err)
			}

			ingredients[i].ProductID = id
		}
		groceries = append(groceries, Grocery{
			Ingredient:  ingredients[i],
			HouseholdID: hid,
			Picked:      false,
		})
	}

	return s.repo.AddGroceries(ctx, groceries)
}

func (s *Service) SmartAdd(ctx context.Context, text string) ([]ingredient.Ingredient, error) {
	return s.ingredientExtractor.FromText(ctx, text)
}

func (s *Service) List(ctx context.Context, sortBy, order string, householdID int) (GroceriesView, error) {
	allowedSorts := map[string]string{
		"product": "p.name",
		"brand":   "p.brand",
		"store":   "p.store",
		"amount":  "g.amount",
	}

	column, ok := allowedSorts[sortBy]

	if !ok {
		column = "p.name"
	}

	groceries, err := s.repo.List(ctx, column, order, householdID)

	if err != nil {
		return GroceriesView{}, err
	}

	var sortedGroceries GroceriesView

	for _, g := range groceries {
		if g.Picked {
			sortedGroceries.Picked = append(sortedGroceries.Picked, g)
		} else {
			switch g.Ingredient.Product.Category {
			case "dairy":
				sortedGroceries.Dairy = append(sortedGroceries.Dairy, g)
			case "pantry":
				sortedGroceries.Pantry = append(sortedGroceries.Pantry, g)
			case "fruit & vegetables":
				sortedGroceries.FruitAndVegetables = append(sortedGroceries.FruitAndVegetables, g)
			case "meat and fish":
				sortedGroceries.MeatAndFish = append(sortedGroceries.MeatAndFish, g)
			default:
				sortedGroceries.Other = append(sortedGroceries.Other, g)
			}
		}
	}

	return sortedGroceries, nil
}

func (s *Service) TogglePicked(ctx context.Context, id, householdID int) error {
	return s.repo.TogglePicked(ctx, id, householdID)
}

func (s *Service) Delete(ctx context.Context, groceryID, householdId int) error {
	return s.repo.Delete(ctx, groceryID, householdId)
}

func (s *Service) DeletePicked(ctx context.Context, householdId int) error {
	return s.repo.DeletePicked(ctx, householdId)
}

func (s *Service) Edit(ctx context.Context, ing ingredient.Ingredient, groceryID, householdID int) error {
	productID, err := s.productService.GetID(ctx, ing.Product)
	if err != nil {
		panic(err)
	}

	ing.ProductID = productID



	return s.repo.Edit(ctx, ing, groceryID, householdID)
}
