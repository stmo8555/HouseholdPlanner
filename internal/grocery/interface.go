package grocery

import (
	"context"

	"github.com/stmo8555/HouseholdPlanner/internal/ingredient"
	"github.com/stmo8555/HouseholdPlanner/internal/product"
)

type IRepo interface {
	GroceryLists(ctx context.Context, hid int) ([]GroceryList, error)
	GroceryList(ctx context.Context, groceryListID, hid int) (GroceryList, error)
	GroceryListsStats(ctx context.Context, hid int) (map[int]GroceryListStats, error)

	CreateList(ctx context.Context, name string, hid int) error
	UpdateGroceryList(ctx context.Context, newName string, groceryListID int, hid int) error
	DeleteGroceryList(ctx context.Context, groceryListID int, hid int) error
	TransferGroceries(ctx context.Context, groceryListTargetID int, groceryListID int, hid int) error

	CreateGroceries(ctx context.Context, groceries []Grocery) error
	Groceries(ctx context.Context, sortBy, order string, groceryListID, householdID int) ([]Grocery, error)
	Grocery(ctx context.Context, itemID int, hid int) (Grocery, error) 
	UpdateGrocery(ctx context.Context, ing ingredient.Ingredient, groceryID, householdID int) error
	MoveGrocery(ctx context.Context, groceryID, groceryListTargetID, householdID int) error
	DeleteGrocery(ctx context.Context, groceryID, householdId int) error
	DeletePicked(ctx context.Context, groceryListID, householdId int) error
	TogglePicked(ctx context.Context, id, householdID int) error

	TopProducts(ctx context.Context, householdID int) ([]product.Product, error)
}

