package grocery

import (
	"cmp"
	"context"
	"slices"

	"github.com/stmo8555/HouseholdPlanner/internal/ingredient"
	"github.com/stmo8555/HouseholdPlanner/internal/product"
)

type Service struct {
	repo                IRepo
	productService      *product.Service
	ingredientExtractor *ingredient.Extractor
}

func (s *Service) Grocery(ctx context.Context, itemID, hid int) (Grocery, error) {
	return s.repo.Grocery(ctx, itemID, hid)
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
			GroceryList:      v,
			GroceryListStats: groceryListsStats[v.ID],
		})
	}

	slices.SortFunc(groceryListsView, func(a, b GroceryListView) int {
		return cmp.Compare(b.GroceryListStats.Total, a.GroceryListStats.Total)
	})

	return groceryListsView, nil
}

func (s *Service) GroceryList(ctx context.Context, groceryListID, hid int) (GroceryList, error) {
	return s.repo.GroceryList(ctx, groceryListID, hid)
}

func (s *Service) CreateList(ctx context.Context, name string, hid int) error {
	return s.repo.CreateList(ctx, name, hid)
}

func (s *Service) UpdateGroceryList(ctx context.Context, newName string, groceryListID int, hid int) error {
	return s.repo.UpdateGroceryList(ctx, newName, groceryListID, hid)
}

func (s *Service) DeleteGroceryList(ctx context.Context, groceryListID int, hid int) error {
	return s.repo.DeleteGroceryList(ctx, groceryListID, hid)
}

func (s *Service) TransferGroceries(ctx context.Context, groceryListTargetID int, groceryListID int, hid int) error {
	return s.repo.TransferGroceries(ctx, groceryListTargetID, groceryListID, hid)
}

func (s *Service) MoveGrocery(ctx context.Context, groceryID int, groceryListTargetID int, hid int) error {
	return s.repo.MoveGrocery(ctx, groceryID, groceryListTargetID, hid)
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
			ListID: groceryListID,
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

func (s *Service) UpdateGrocery(ctx context.Context, ing ingredient.Ingredient, groceryID, householdID int) error {
	productID, err := s.productService.GetID(ctx, ing.Product)
	if err != nil {
		return err
	}

	ing.ProductID = productID

	return s.repo.UpdateGrocery(ctx, ing, groceryID, householdID)
}

func (s *Service) DeleteGrocery(ctx context.Context, groceryID, householdId int) error {
	return s.repo.DeleteGrocery(ctx, groceryID, householdId)
}

func (s *Service) DeletePicked(ctx context.Context, groceryListID, householdId int) error {
	return s.repo.DeletePicked(ctx, groceryListID, householdId)
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
