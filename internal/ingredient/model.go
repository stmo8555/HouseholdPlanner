package ingredient

import "github.com/stmo8555/HouseholdPlanner/internal/product"

type Ingredient struct {
	ProductID int
	Product   product.Product
	Amount    string
	Note      string
}
