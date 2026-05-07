package recipe

import "github.com/stmo8555/HouseholdPlanner/internal/product"

type Recipe struct {
	Id          int    `json:"id"`
	Title       string `json:"title"`
	ImgURL      string `json:"img_url"`
	Link        string `json:"link"`
	HouseholdID int    `json:"household_id"`
}

type RecipeIngredient struct {
	Id        int     `json:"id"`
	RecipeID  int     `json:"recipe_id"`
	ProductID int     `json:"product_id"`
	Amount    string  `json:"amount"`
	Product   product.Product `json:"product"`
}
