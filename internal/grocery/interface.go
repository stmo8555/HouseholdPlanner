package grocery

import (
	"context"

	"github.com/stmo8555/HouseholdPlanner/internal/ingredient"
	"github.com/stmo8555/HouseholdPlanner/internal/product"
)

type IRepo interface {
	getTopProducts(ctx context.Context, householdID int) ([]product.Product, error)
	AddGroceries(ctx context.Context, groceries []Grocery) error
	List(ctx context.Context, sortBy, order string, householdID int) ([]Grocery, error)
	// AmountOfUnpickedGroceries(ctx context.Context, hid int) (int, error)
	TogglePicked(ctx context.Context, id, householdID int) error
	Delete(ctx context.Context, groceryID, householdId int) error
	DeletePicked(ctx context.Context, householdId int) error
	Edit(ctx context.Context, ing ingredient.Ingredient, groceryID, householdID int) error
}
