package grocery

import (
	"context"
)

type IRepo interface {
	getTopProducts(ctx context.Context, householdID int) ([]string, error)
	AddGroceries(ctx context.Context, groceries []Grocery) error
	List(ctx context.Context, sortBy, order string, householdID int) ([]Grocery, error)
	AmountOfUnpickedGroceries(ctx context.Context, hid int) (int, error)
	TogglePicked(ctx context.Context, id, householdID int) error
	DeletePicked(ctx context.Context, householdId int) error
	Edit(ctx context.Context, groceries []Grocery) error
}
