package recipe

import (
	"github.com/stmo8555/HouseholdPlanner/internal/ingredient"
)

type Recipe struct {
	Id          int    `json:"id"`
	Title       string `json:"title"`
	ImgURL      string `json:"img_url"`
	Link        string `json:"link"`
	HouseholdID int    `json:"household_id"`
}

type RecipeIngredient struct {
	Id         int                    `json:"id"`
	RecipeID   int                    `json:"recipe_id"`
	Ingredient ingredient.Ingredient  `json:"ingredient"`
}

type RecipeView struct {
	Recipe
	RecipeIngredients []RecipeIngredient
}
