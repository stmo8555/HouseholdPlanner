package grocery

import "github.com/stmo8555/HouseholdPlanner/internal/ingredient"

type GroceryList struct {
	ID          int
	Name        string
	HouseholdID int
}

type GroceryListStats struct {
    ListID    int
    Total     int
    Picked    int
}

type GroceryListView struct {
	GroceryList GroceryList
	GroceryListStats GroceryListStats
}

type Grocery struct {
	ID            int                   `json:"id"`
	Ingredient    ingredient.Ingredient `json:"ingredient"`
	HouseholdID   int                   `json:"household_id"`
	Picked        bool                  `json:"picked"`
	ListID int
}

type GroceriesView struct {
	Dairy              []Grocery
	FruitAndVegetables []Grocery
	MeatAndFish        []Grocery
	Pantry             []Grocery
	Other              []Grocery
	Picked             []Grocery
}

func (g GroceriesView) Total() int {
	return len(g.Pantry) +
		len(g.FruitAndVegetables) +
		len(g.MeatAndFish) +
		len(g.Dairy) +
		len(g.Other)
}
