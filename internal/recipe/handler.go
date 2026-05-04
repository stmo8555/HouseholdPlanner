package recipe

import (
	"github.com/gin-gonic/gin"
	"github.com/stmo8555/HouseholdPlanner/internal/grocery"
	"strings"
)

type Handler struct {
	Service        *Service
	GroceryService *grocery.Service
}

func (h *Handler) List(c *gin.Context) {
	hid := c.GetInt("household_id")

	recipes, err := h.Service.List(c, hid)

	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	data := gin.H{
		"Title":       "Groceries",
		"CurrentPath": c.Request.URL.Path,
		"Data":        recipes,
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

	err := h.Service.Add(c, hid, link)

	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	c.Redirect(302, "/recipes")
}

func (h *Handler) IngredientsFromRecipe(c *gin.Context) {
	link := c.PostForm("link")
	groceries := h.GroceryService.IngredientsFromRecipe(c, link)

	data := gin.H{
		"Title":     "Extracted Groceries",
		"Groceries": groceries,
		"SaveURL":   "/recipes/extracted",
		"CancelURL": "/recipes",
	}

	c.HTML(200, "groceries_extraction.html", data)
}

func (h *Handler) AcceptExtractedGroceries(c *gin.Context) {
	products := c.PostFormArray("product")
	amounts := c.PostFormArray("amount")
	brands := c.PostFormArray("brand")
	stores := c.PostFormArray("store")

	hid := c.GetInt("household_id")
	groceries := make([]grocery.Grocery, len(products))
	for i := range len(groceries) {
		groceries[i] = grocery.Grocery{
			Product:     products[i],
			Amount:      amounts[i],
			Brand:       brands[i],
			Store:       stores[i],
			HouseholdID: hid,
		}
	}

	err := h.GroceryService.AddGroceries(c, groceries)

	if err != nil {
		c.AbortWithStatus(500)
		c.String(500, err.Error())
		return
	}

	c.Redirect(302, "/recipes")
}
