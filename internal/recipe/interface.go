package recipe

import (
	"context"

	"github.com/stmo8555/HouseholdPlanner/internal/ingredient"
	"github.com/stmo8555/HouseholdPlanner/internal/product"
)

type IRepo interface {
	List(ctx context.Context, hid int) ([]Recipe, map[int][]RecipeIngredient, error)
	AddRecipe(ctx context.Context, hid int, recipe Recipe) (int, error)
	Ingredients(ctx context.Context, recipeID, hid int) ([]RecipeIngredient, error)
	AddIngredients(ctx context.Context, recipeIngredients []RecipeIngredient) error
}

type IIngredientExtractor interface {
	FromRecipeURL(ctx context.Context, url string) ([]ingredient.Ingredient, error)
}

type IProductService interface {
	GetID(ctx context.Context, p product.Product) (int, error)
}
