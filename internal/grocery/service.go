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

func NewService(repo IRepo, product *product.Service, ingredient *ingredient.Extractor) *Service {
	if repo == nil || product == nil || ingredient == nil {
		panic("service not initialized")
	}

	return &Service{
		repo:                repo,
		productService:      product,
		ingredientExtractor: ingredient,
	}
}

func (s *Service) GroceryLists(ctx context.Context, hid int) ([]GroceryListView, error) {
	groceryLists, err := s.repo.GroceryLists(ctx, hid)

	if err != nil {
		return nil, err
	}

	groceryListsStats, err := s.repo.GroceryListsStats(ctx, hid)

	if err != nil {
		return nil, err
	}

	groceryListsView := make([]GroceryListView, 0, len(groceryLists))
	for _, v := range groceryLists {
		groceryListsView = append(groceryListsView, GroceryListView{
			GroceryList: v,
			GroceryListStats: groceryListsStats[v.ID],
		})
	}

	return groceryListsView, nil
}

func (s *Service) CreateList(ctx context.Context, name string, hid int) error {
	return s.repo.CreateList(ctx, name, hid)
}

func (s *Service) TopProducts(ctx context.Context, householdID int) ([]string, error) {
	products, err := s.repo.TopProducts(ctx, householdID)

	if err != nil {
		return nil, err
	}

	strSlice := make([]string, 0, len(products))

	for _, v := range products {
		strSlice = append(strSlice, v.Name)
	}

	return strSlice, err
}

func (s *Service) CreateGroceries(ctx context.Context, ingredients []ingredient.Ingredient, groceryListID, hid int) error {
	groceries := make([]Grocery, 0, len(ingredients))
	for i := range ingredients {
		if ingredients[i].ProductID == 0 {
			id, err := s.productService.GetID(ctx, ingredients[i].Product)

			if err != nil {
				return err
			}

			ingredients[i].ProductID = id
		}
		groceries = append(groceries, Grocery{
			Ingredient:    ingredients[i],
			GroceryListID: groceryListID,
			HouseholdID:   hid,
			Picked:        false,
		})
	}

	return s.repo.CreateGroceries(ctx, groceries)
}

func (s *Service) ParseGroceries(ctx context.Context, text string) ([]ingredient.Ingredient, error) {
	return s.ingredientExtractor.FromText(ctx, text)
}

func (s *Service) GroceriesView(ctx context.Context, sortBy, order string, groceryListID, householdID int) (GroceriesView, error) {
	allowedSorts := map[string]string{
		"product": "p.name",
		"brand":   "p.brand",
		"amount":  "g.amount",
	}

	column, ok := allowedSorts[sortBy]

	if !ok {
		column = "p.name"
	}

	groceries, err := s.repo.Groceries(ctx, column, order, groceryListID, householdID)

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

func (s *Service) DeleteGrocery(ctx context.Context, groceryID, householdId int) error {
	return s.repo.DeleteGrocery(ctx, groceryID, householdId)
}

func (s *Service) DeletePicked(ctx context.Context, groceryListID, householdId int) error {
	return s.repo.DeletePicked(ctx, groceryListID, householdId)
}

func (s *Service) UpdateGrocery(ctx context.Context, ing ingredient.Ingredient, groceryID, householdID int) error {
	productID, err := s.productService.GetID(ctx, ing.Product)
	if err != nil {
		panic(err)
	}

	ing.ProductID = productID

	return s.repo.UpdateGrocery(ctx, ing, groceryID, householdID)
}
