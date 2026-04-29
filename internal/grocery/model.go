package grocery

type Grocery struct {
	Id          int    `json:"id"`
	Product     string `json:"product"`
	Amount      string `json:"amount"`
	Brand       string `json:"brand"`
	Store       string `json:"store"`
	Category    string
	HouseholdID int
	Picked      bool
}

type GroceriesView struct {
	Dairy              []Grocery
	FruitAndVegetables []Grocery
	MeatAndFish        []Grocery
	Pantry             []Grocery
	Other              []Grocery
	Picked             []Grocery
}
