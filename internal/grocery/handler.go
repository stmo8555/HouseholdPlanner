package grocery

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/stmo8555/HouseholdPlanner/internal/ingredient"
	"github.com/stmo8555/HouseholdPlanner/internal/product"
)

type Handler struct {
	service             *Service
	ingredientExtractor *ingredient.Extractor
}

func CreateHandler(s *Service, ingredient *ingredient.Extractor) *Handler {
	if s == nil || ingredient == nil {
		panic("nil service for handler")
	}

	return &Handler{
		service:             s,
		ingredientExtractor: ingredient,
	}
}

func (h *Handler) IngredientsFromRecipe(c *gin.Context) {
	link := c.PostForm("link")

	if strings.TrimSpace(link) == "" {
		c.AbortWithStatus(500)
		c.String(500, "Empty link")
		return
	}

	ingredients, err := h.ingredientExtractor.FromRecipeURL(c, link)

	if err != nil {
		panic(err)
	}

	h.ExtractedView(c, ingredients)
}

func (h *Handler) AcceptExtractedGroceries(c *gin.Context) {
	hid := c.GetInt("household_id")

	redirectPath := c.PostForm("redirect_path")

	products := c.PostFormArray("name")
	amounts := c.PostFormArray("amount")
	brands := c.PostFormArray("brand")
	stores := c.PostFormArray("store")

	ingredients := make([]ingredient.Ingredient, len(products))
	for i := range ingredients {
		ingredients[i] = ingredient.Ingredient{
			Product: product.Product{
				Name:     products[i],
				Brand:    brands[i],
				Store:    stores[i],
				Category: "",
			},
			Amount: amounts[i],
		}
	}

	err := h.service.AddGroceries(c, ingredients, hid)

	if err != nil {
		c.AbortWithStatus(500)
		c.String(500, err.Error())
		return
	}

	if !strings.HasPrefix(redirectPath,"/"){
		c.AbortWithStatus(500)
		c.String(500, errors.New("Redirect variable does not start with /. Value: " + redirectPath).Error())	
		return
	}

	c.Redirect(303, redirectPath)
}

func (h *Handler) List(c *gin.Context) {
	hid := c.GetInt("household_id")

	sortBy := c.DefaultQuery("sort", "product")
	order := c.DefaultQuery("order", "asc")

	if order != "desc" {
		order = "asc"
	}

	nextOrder := "asc"
	if order == "asc" {
		nextOrder = "desc"
	}

	groceries, err := h.service.List(c, sortBy, order, hid)
	if err != nil {
		c.AbortWithStatus(500)
		c.String(500, err.Error())
		return
	}

	var topProducts []string
	topProducts, err = h.service.GetTopProducts(c, hid)

	if err != nil {
		c.AbortWithStatus(500)
		c.String(500, err.Error())
		return
	}

	data := gin.H{
		"Title":       "Groceries",
		"CurrentPath": c.Request.URL.Path,
		"Data":        groceries,
		"Total": len(groceries.Pantry) +
			len(groceries.FruitAndVegetables) +
			len(groceries.MeatAndFish) +
			len(groceries.Dairy) +
			len(groceries.Other),
		"TopProducts": topProducts,
		"Sort":        sortBy,
		"Order":       order,
		"NextOrder":   nextOrder,
	}

	c.HTML(200, "groceries.html", data)
}

func (h *Handler) TogglePicked(c *gin.Context) {
	hid := c.GetInt("household_id")
	idStr := c.PostForm("id")

	id, err := strconv.Atoi(idStr)

	if err != nil {
		panic(err)
	}

	h.service.TogglePicked(c, id, hid)
	c.Redirect(302, "/groceries")
}

func (h *Handler) Add(c *gin.Context) {
	ingredients := make([]ingredient.Ingredient, 0, 1)
	ingredients = append(ingredients, ingredient.Ingredient{
		Product: product.Product{
			Name:     c.PostForm("name"),
			Brand:    c.PostForm("brand"),
			Store:    c.PostForm("store"),
			Category: "",
		},
		Amount: c.PostForm("amount"),
	})

	err := h.service.AddGroceries(c, ingredients, c.GetInt("household_id"))

	if err != nil {
		c.AbortWithStatus(500)
		c.String(500, err.Error())
		return
	}

	c.Redirect(302, "/groceries")
}

func (h *Handler) ExtractedView(c *gin.Context, ingredients []ingredient.Ingredient) {
	data := gin.H{
		"Title":      "Extracted Groceries",
		"Ingredients": ingredients,
		"CancelURL": "/groceries",
		"RedirectPath": "/groceries",
	}

	c.HTML(200, "groceries_extraction.html", data)
}

func (h *Handler) SmartAdd(c *gin.Context) {
	text := c.PostForm("text")

	if strings.TrimSpace(text) == "" {
		c.AbortWithStatus(500)
		c.String(500, "Empty text")
		return
	}

	groceries, err := h.service.SmartAdd(c, text)

	if err != nil {
		c.AbortWithStatus(500)
		c.String(500, err.Error())
		return
	}

	h.ExtractedView(c, groceries)
}

func (h *Handler) DeletePicked(c *gin.Context) {
	err := h.service.DeletePicked(c, c.GetInt("household_id"))

	if err != nil {
		c.AbortWithStatus(500)
		c.String(500, err.Error())
		return
	}

	c.Redirect(302, "/groceries")
}

func (h *Handler) Edit(c *gin.Context) {
	var groceries []Grocery
	err := c.BindJSON(&groceries)
	if err != nil {
		panic(err)
	}

	err = h.service.Edit(c, groceries, c.GetInt("household_id"))

	if err != nil {
		c.AbortWithStatus(500)
		c.String(500, err.Error())
		return
	}

	c.Redirect(302, "/groceries")
}
