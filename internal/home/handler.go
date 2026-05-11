package home

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/stmo8555/HouseholdPlanner/internal/grocery"
	"github.com/stmo8555/HouseholdPlanner/internal/login"
	"github.com/stmo8555/HouseholdPlanner/internal/recipe"
	"github.com/stmo8555/HouseholdPlanner/internal/todo"
)

type Handler struct {
	groceriesService *grocery.Service
	loginService     *login.Service
	recipesService   *recipe.Service
	todosService     *todo.Service
	service          *Service
}

func CreateHandler(s *Service) *Handler {
	if s == nil {
		panic("nil service for handler")
	}

	return &Handler{
		service: s,
	}
}


func (h *Handler) Index(c *gin.Context) {
	// hid := c.GetInt("household_id")

	// groceriesCount, err := h.GroceriesService.CountUnpicked(c, hid)
	//
	// if err != nil {
	// 	c.AbortWithStatus(500)
	// 	c.String(500, err.Error())
	// }
	//
	// var todosCount int
	// todosCount, err = h.TodosService.Count(c, hid)
	//
	// if err != nil {
	// 	c.AbortWithStatus(500)
	// 	c.String(500, err.Error())
	// }
	//
	// data := gin.H{
	// 	"Title":       "Home",
	// 	"CurrentPath": c.Request.URL.Path,
	// 	"Todos":       todosCount,
	// 	"Groceries":   groceriesCount,
	// }

	// c.HTML(http.StatusOK, "index.html", data)
}

// func (h *Handler) AddGrocery(c *gin.Context) {
// 	hid := c.GetInt("household_id")
// 	product := strings.TrimSpace(c.PostForm("product"))
//
// 	if product == "" {
// 		panic(errors.New("No product value in home add grocery"))
// 	}
//
// 	groceryItem := grocery.Grocery{Product: product.Product{}, HouseholdID: hid}
// 	err := h.GroceriesService.AddGroceries(c, []grocery.Grocery{groceryItem})
//
// 	if err != nil {
// 		c.AbortWithStatus(500)
// 		c.String(500, err.Error())
// 		return
// 	}
//
// 	c.Redirect(302, "/home")
// }

func (h *Handler) AddRecipe(c *gin.Context) {
	hid := c.GetInt("household_id")
	recipe := strings.TrimSpace(c.PostForm("recipe"))

	if recipe == "" {
		panic(errors.New("No product value in home add grocery"))
	}

	err := h.recipesService.Add(c, hid, recipe)

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

	content := h.service.AI(c, question)

	data := gin.H{
		"Groceries": content.Groceries,
		"Todos":     content.Todos,
		"Recipes":   content.Recipes,
	}

	c.HTML(200, "ai_extraction.html", data)
}
