package home

import (
	"encoding/json"
	"github.com/stmo8555/HouseholdPlanner/internal/grocery"
	"github.com/stmo8555/HouseholdPlanner/internal/recipe"
	"github.com/stmo8555/HouseholdPlanner/internal/todo"
)

type Envelope struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type Content struct {
	Groceries []grocery.Grocery
	Recipes   []recipe.Recipe
	Todos     []todo.Todo
}
