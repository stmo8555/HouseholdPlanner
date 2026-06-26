package grocery

import "github.com/stmo8555/HouseholdPlanner/internal/ingredient"

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
	Categories map[string][]Grocery
	Picked     []Grocery
}

func (g GroceriesView) Total() int {
	total := 0
	for _, items := range g.Categories {
		total += len(items)
	}
	return total
}
