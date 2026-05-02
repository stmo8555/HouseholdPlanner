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

type GroceryAI struct {
	Product string `json:"product" jsonschema_description:"Grocery item name only. No brand, amount, store, or preparation words."`
	Amount  string `json:"amount" jsonschema_description:"Explicit amount only. Empty string if not written. Never guess."`
	Brand   string `json:"brand" jsonschema_description:"Explicit brand only. Empty string if not written. Never guess."`
	Store   string `json:"store" jsonschema_description:"Explicit store only. Empty string if not written. Never guess."`
}

type GroceriesAI struct {
	List []GroceryAI `json:"groceries" jsonschema_description:"Extracted grocery items."`
}

type GroceriesView struct {
	Dairy              []Grocery
	FruitAndVegetables []Grocery
	MeatAndFish        []Grocery
	Pantry             []Grocery
	Other              []Grocery
	Picked             []Grocery
}
