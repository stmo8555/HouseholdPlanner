package grocery

type Grocery struct {
	Id          int    `json:"id"`
	Product     string `json:"product"`
	Amount      string `json:"amount"`
	Brand       string `json:"brand"`
	Store       string `json:"store"`
	HouseholdID int
	Picked      bool
}

type Groceries struct {
	Picked, NotPicked []Grocery
}
