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

func NewHandler(s *Service, gs *grocery.Service) *Handler {
	if s == nil || gs == nil {
		panic("nil services for handler")
	}

	return &Handler{
		service:        s,
		groceryService: gs,
	}
}

func (h *Handler) List(c *gin.Context) {
	data, err := h.recipeListData(c)

	if err != nil {
		c.AbortWithError(500, err)
		c.String(500, err.Error())
		return
	}
	
	data["Title"] = "Groceries"
	data["CurrentPath"] = c.Request.URL.Path

	c.HTML(200, "recipes.html", data)
}

func (h *Handler) ListPartial(c *gin.Context) {
	data, err := h.recipeListData(c)
	
	if err != nil {
		c.AbortWithError(500, err)
		c.String(500, err.Error())
		return
	}

	c.HTML(200, "recipe-list", data)
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

	h.ListPartial(c)
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
		"Ingredients":  ingredients,
		"CancelURL":    "/recipes",
		"RedirectPath": "/recipes",
	}

	c.HTML(200, "groceries_extraction.html", data)
}

func (h *Handler) recipeListData(c *gin.Context) (gin.H, error) {
	hid := c.GetInt("household_id")
	recipes, recipeIngredients, err := h.service.List(c, hid)

	if err != nil {
		c.AbortWithError(500, err)
		c.String(500, err.Error())
		return nil, err
	}

	recipesView := make([]RecipeView, 0, len(recipes))
	for _, v := range recipes {
		recipesView = append(recipesView, RecipeView{
			Recipe:            v,
			RecipeIngredients: recipeIngredients[v.Id],
		})
	}

	return gin.H{ "RecipesView": recipesView, }, nil
}
