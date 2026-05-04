package recipe

import (
	"context"
)

type IRepo interface {
	List(ctx context.Context, hid int) ([]Recipe, error)
	Add(ctx context.Context, hid int, recipe Recipe) error
}
