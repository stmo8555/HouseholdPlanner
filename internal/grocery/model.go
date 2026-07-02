package grocery

import "github.com/stmo8555/HouseholdPlanner/internal/ingredient"

var Categories = []string{"Dairy", "Frozen", "Pantry", "Fruit & veg", "Meat & fish", "Other"}

func IsValidCategory(c string) bool {
	for _, category := range Categories {
		if category == c {
			return true
		}
	}
	return false
}

type GroceryList struct {
	ID          int
	Name        string
	HouseholdID int
}

type GroceryListStats struct {
	ListID     int
	Total      int
	Categories []CategoryCount
}

type CategoryCount struct {
	Label string
	Count int
}

type GroceryListView struct {
	GroceryList      GroceryList
	GroceryListStats GroceryListStats
}

type Grocery struct {
	ID         int                   `json:"id"`
	Ingredient ingredient.Ingredient `json:"ingredient"`
	Picked     bool                  `json:"picked"`
	ListID     int
}

type GroceriesView struct {
	Dairy              []Grocery
	FruitAndVegetables []Grocery
	MeatAndFish        []Grocery
	Frozen             []Grocery
	Pantry             []Grocery
	Other              []Grocery
	Picked             []Grocery
}

func (g GroceriesView) Total() int {
	return len(g.Pantry) +
		len(g.FruitAndVegetables) +
		len(g.MeatAndFish) +
		len(g.Frozen) +
		len(g.Dairy) +
		len(g.Other)
}
