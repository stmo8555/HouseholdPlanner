package recipe

import (
	"context"
)

type IRepo interface {
	List(ctx context.Context, hid int) ([]Recipe, map[int][]RecipeIngredient, error)
	AddRecipe(ctx context.Context, hid int, recipe Recipe) (int, error)
	Ingredients(ctx context.Context, recipeID, hid int) ([]RecipeIngredient, error)
	AddIngredients(ctx context.Context, recipeIngredients []RecipeIngredient) error
}
