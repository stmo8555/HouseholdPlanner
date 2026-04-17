package home

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/stmo8555/HouseholdPlanner/internal/groceries"
	"github.com/stmo8555/HouseholdPlanner/internal/login"
	"github.com/stmo8555/HouseholdPlanner/internal/recipes"
	"github.com/stmo8555/HouseholdPlanner/internal/todos"
)

type Handler struct {
	GroceriesService *groceries.Service
	LoginService     *login.Service
	RecipesService   *recipes.Service
	TodosService     *todos.Service
	Service          *Service
}

func (h *Handler) Index(c *gin.Context) {
	hid := c.GetInt("household_id")

	groceriesCount, err := h.GroceriesService.CountUnpicked(c, hid)

	if err != nil {
		c.AbortWithStatus(500)
		c.String(500, err.Error())
	}

	var todosCount int
	todosCount, err = h.TodosService.Count(c, hid)

	if err != nil {
		c.AbortWithStatus(500)
		c.String(500, err.Error())
	}

	data := gin.H{
		"Title":       "Home",
		"CurrentPath": c.Request.URL.Path,
		"Todos":       todosCount,
		"Groceries":   groceriesCount,
	}

	c.HTML(http.StatusOK, "index.html", data)
}

func (h *Handler) AddGrocery(c *gin.Context) {
	hid := c.GetInt("household_id")
	product := strings.TrimSpace(c.PostForm("product"))

	if product == "" {
		panic(errors.New("No product value in home add grocery"))
	}

	grocery := groceries.Grocery{Product: product, HouseholdID: hid}
	err := h.GroceriesService.AddGroceries(c, []groceries.Grocery{grocery})

	if err != nil {
		c.AbortWithStatus(500)
		c.String(500, err.Error())
		return
	}

	c.Redirect(302, "/home")
}

func (h *Handler) AddRecipe(c *gin.Context) {
	hid := c.GetInt("household_id")
	recipe := strings.TrimSpace(c.PostForm("recipe"))

	if recipe == "" {
		panic(errors.New("No product value in home add grocery"))
	}

	err := h.RecipesService.Add(c, hid, recipe)

	if err != nil {
		c.AbortWithStatus(500)
		c.String(500, err.Error())
		return
	}

	c.Redirect(302, "/home")
}

func (h *Handler) AI(c *gin.Context) {
	// hid := c.GetInt("household_id")
	question := c.PostForm("question")

	if strings.TrimSpace(question) == "" {
		c.AbortWithStatus(500)
		c.String(500, errors.New("Not a valid question").Error())
		return
	}

	content := h.Service.AI(c, question)

	data := gin.H{
		"Groceries": content.Groceries,
		"Todos":     content.Todos,
		"Recipes":   content.Recipes,
	}

	c.HTML(200, "ai_extraction.html", data)
}
