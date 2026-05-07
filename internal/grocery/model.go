package grocery

import "github.com/stmo8555/HouseholdPlanner/internal/product"

type Grocery struct {
	Id          int     `json:"id"`
	ProductID   int     `json:"product_id"`
	Product     product.Product `json:"product"`
	Amount      string  `json:"amount"`
	HouseholdID int     `json:"household_id"`
	Picked      bool    `json:"picked"`
}


type GroceriesView struct {
	Dairy              []Grocery
	FruitAndVegetables []Grocery
	MeatAndFish        []Grocery
	Pantry             []Grocery
	Other              []Grocery
	Picked             []Grocery
}
