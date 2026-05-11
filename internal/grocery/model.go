package grocery

import "github.com/stmo8555/HouseholdPlanner/internal/ingredient"

type Grocery struct {
	Id          int                   `json:"id"`
	Ingredient  ingredient.Ingredient `json:"ingredient"`
	HouseholdID int                   `json:"household_id"`
	Picked      bool                  `json:"picked"`
}

type GroceriesView struct {
	Dairy              []Grocery
	FruitAndVegetables []Grocery
	MeatAndFish        []Grocery
	Pantry             []Grocery
	Other              []Grocery
	Picked             []Grocery
}
