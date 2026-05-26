package grocery

import (
	"context"
	"github.com/stmo8555/HouseholdPlanner/internal/ingredient"
	"github.com/stmo8555/HouseholdPlanner/internal/product"
)

type IRepo interface {
	TopProducts(ctx context.Context, householdID int) ([]product.Product, error)
	CreateGroceries(ctx context.Context, groceries []Grocery) error
	CreateList(ctx context.Context, name string, hid int) error
	Groceries(ctx context.Context, sortBy, order string, groceryListID, householdID int) ([]Grocery, error)
	GroceryLists(ctx context.Context, hid int) ([]GroceryList, error)
	GroceryListsStats(ctx context.Context, hid int) (map[int]GroceryListStats, error)
	TogglePicked(ctx context.Context, id, householdID int) error
	DeleteGrocery(ctx context.Context, groceryID, householdId int) error
	DeletePicked(ctx context.Context, groceryListID, householdId int) error
	UpdateGrocery(ctx context.Context, ing ingredient.Ingredient, groceryID, householdID int) error
}
