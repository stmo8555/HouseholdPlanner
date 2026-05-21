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

	ingredients := make([]ingredient.Ingredient, len(products))
	for i := range ingredients {
		ingredients[i] = ingredient.Ingredient{
			Product: product.Product{
				Name:     products[i],
				Brand:    brands[i],
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

	if !strings.HasPrefix(redirectPath, "/") {
		c.AbortWithStatus(500)
		c.String(500, errors.New("Redirect variable does not start with /. Value: "+redirectPath).Error())
		return
	}

	c.Redirect(303, redirectPath)
}

func (h *Handler) List(c *gin.Context) {
	hid := c.GetInt("household_id")

	data, err := h.groceryListData(c, hid)
	if err != nil {
		c.String(500, err.Error())
		return
	}

	data["Title"] = "Groceries"
	data["CurrentPath"] = c.Request.URL.Path

	c.HTML(200, "groceries.html", data)
}

func (h *Handler) ListPartial(c *gin.Context) {
	hid := c.GetInt("household_id")

	data, err := h.groceryListData(c, hid)
	if err != nil {
		c.String(500, err.Error())
		return
	}

	c.HTML(200, "grocery_list", data)
}

func (h *Handler) TogglePicked(c *gin.Context) {
	hid := c.GetInt("household_id")

	id, err := strconv.Atoi(c.PostForm("id"))
	if err != nil {
		c.String(400, "invalid grocery id")
		return
	}

	if err := h.service.TogglePicked(c, id, hid); err != nil {
		c.String(500, err.Error())
		return
	}

	h.ListPartial(c)
}

func (h *Handler) Add(c *gin.Context) {
	hid := c.GetInt("household_id")
	ingredients := make([]ingredient.Ingredient, 0, 1)
	ingredients = append(ingredients, ingredient.Ingredient{
		Product: product.Product{
			Name:     c.PostForm("name"),
			Brand:    c.PostForm("brand"),
			Category: "",
		},
		Amount: c.PostForm("amount"),
	})

	err := h.service.AddGroceries(c, ingredients, hid)

	if err != nil {
		c.AbortWithStatus(500)
		c.String(500, err.Error())
		return
	}

	h.ListPartial(c)
}

func (h *Handler) ExtractedView(c *gin.Context, ingredients []ingredient.Ingredient) {
	data := gin.H{
		"Ingredients":  ingredients,
		"CancelURL":    "/groceries",
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

func (h *Handler) Delete(c *gin.Context) {
	hid := c.GetInt("household_id")
	groceryID, err := strconv.Atoi(c.PostForm("id"))
	
	if err != nil {
		c.String(500, err.Error())
		return
	}

	err = h.service.Delete(c, groceryID, hid)

	if err != nil {
		c.String(500, err.Error())
		return
	}

	h.ListPartial(c)
}

func (h *Handler) DeletePicked(c *gin.Context) {
	hid := c.GetInt("household_id")
	err := h.service.DeletePicked(c, hid)

	if err != nil {
		c.String(500, err.Error())
		return
	}

	h.ListPartial(c)
}

func (h *Handler) Edit(c *gin.Context) {
	hid := c.GetInt("household_id")

	prod := product.Product{
		Name:     c.PostForm("name"),
		Brand:    c.PostForm("brand"),
		Category: "",
	}

	ing := ingredient.Ingredient{
		Product: prod,
		Amount:  c.PostForm("amount"),
	}

	id, err := strconv.Atoi(c.PostForm("id"))

	if err != nil {
		panic(err)
	}

	err = h.service.Edit(c, ing, id, hid)

	if err != nil {
		c.AbortWithStatus(500)
		c.String(500, err.Error())
		return
	}

	h.ListPartial(c)
}

func (h *Handler) groceryListData(c *gin.Context, hid int) (gin.H, error) {
	sortBy := c.DefaultQuery("sort", c.DefaultPostForm("sort", "product"))
	order := c.DefaultQuery("order", c.DefaultPostForm("order", "asc"))
	filter := c.DefaultQuery("filter", c.DefaultPostForm("filter", "all"))

	if order != "desc" {
		order = "asc"
	}

	nextOrder := "asc"
	if order == "asc" {
		nextOrder = "desc"
	}

	groceries, err := h.service.List(c, sortBy, order, hid)
	if err != nil {
		return nil, err
	}

	topProducts, err := h.service.GetTopProducts(c, hid)
	if err != nil {
		return nil, err
	}

	var filteredGroceries GroceriesView

	filteredGroceries.Picked = groceries.Picked
	switch filter {
	case "all":
		filteredGroceries = groceries
	case "dairy":
		filteredGroceries.Dairy = groceries.Dairy
	case "fruitandvegetables":
		filteredGroceries.FruitAndVegetables = groceries.FruitAndVegetables
	case "meatandfish":
		filteredGroceries.MeatAndFish = groceries.MeatAndFish
	case "pantry":
		filteredGroceries.Pantry = groceries.Pantry
	case "other":
		filteredGroceries.Other = groceries.Other
	default:
		filter = "all"
		filteredGroceries = groceries
	}

	return gin.H{
		"Data":        filteredGroceries,
		"Total":       filteredGroceries.Total(),
		"TopProducts": topProducts,
		"Sort":        sortBy,
		"Order":       order,
		"NextOrder":   nextOrder,
		"Filter":      filter,
	}, nil
}
