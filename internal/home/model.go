package home

import (
	"encoding/json"
	"github.com/stmo8555/HouseholdPlanner/internal/groceries"
	"github.com/stmo8555/HouseholdPlanner/internal/recipes"
	"github.com/stmo8555/HouseholdPlanner/internal/todos"
)

type Envelope struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type Content struct {
	Groceries []groceries.Grocery
	Recipes   []recipes.Recipe
	Todos     []todos.Todo
}
