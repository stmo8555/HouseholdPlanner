package recipe

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/stmo8555/HouseholdPlanner/internal/grocery"
	"github.com/stmo8555/HouseholdPlanner/internal/ingredient"
)

type Handler struct {
	service        *Service
	groceryService *grocery.Service
}

func CreateHandler(s *Service, gs *grocery.Service) *Handler {
	if s == nil || gs == nil {
		panic("nil services for handler")
	}

	return &Handler{
		service:        s,
		groceryService: gs,
	}
}

func (h *Handler) List(c *gin.Context) {
	hid := c.GetInt("household_id")

	recipes, recipeIngredients, err := h.service.List(c, hid)

	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	recipesView := make([]RecipeView, 0, len(recipes))
	for _, v := range recipes {
		recipesView = append(recipesView, RecipeView{
			Recipe:            v,
			RecipeIngredients: recipeIngredients[v.Id],
		})
	}
	data := gin.H{
		"Title":       "Groceries",
		"CurrentPath": c.Request.URL.Path,
		"RecipesView": recipesView,
	}

	c.HTML(200, "recipes.html", data)
}

func (h *Handler) Add(c *gin.Context) {
	link := strings.TrimSpace(c.PostForm("link"))

	if link == "" {
		c.AbortWithStatus(500)
		return
	}

	hid := c.GetInt("household_id")

	err := h.service.Add(c, hid, link)

	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	c.Redirect(302, "/recipes")
}

func (h *Handler) IngredientsFromRecipe(c *gin.Context) {
	hid := c.GetInt("household_id")
	recipeID, err := strconv.Atoi(c.PostForm("recipe_id"))

	if err != nil {
		panic(err)
	}

	recipeIngredients, err := h.service.Ingredients(c.Request.Context(), recipeID, hid)

	if err != nil {
		panic(err)
	}

	ingredients := make([]ingredient.Ingredient, 0, len(recipeIngredients))

	for _, v := range recipeIngredients {
		ingredients = append(ingredients, v.Ingredient)
	}

	data := gin.H{
		"Title":        "Extracted Groceries",
		"Ingredients":  ingredients,
		"CancelURL":    "/recipes",
		"RedirectPath": "/recipes",
	}

	c.HTML(200, "groceries_extraction.html", data)
}
