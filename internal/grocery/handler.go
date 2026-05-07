package grocery

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/stmo8555/HouseholdPlanner/internal/product"
)

type Handler struct {
	Service *Service
}

func (h *Handler) IngredientsFromRecipe(c *gin.Context) {
	link := c.PostForm("link")

	if strings.TrimSpace(link) == "" {
		c.AbortWithStatus(500)
		c.String(500, "Empty link")
		return
	}

	groceries := h.Service.IngredientsFromRecipe(c, link)

	h.ExtractedView(c, groceries)
}

func (h *Handler) AcceptExtractedGroceries(c *gin.Context) {
	products := c.PostFormArray("name")
	amounts := c.PostFormArray("amount")
	brands := c.PostFormArray("brand")
	stores := c.PostFormArray("store")

	hid := c.GetInt("household_id")
	groceries := make([]Grocery, len(products))
	for i := range groceries {
		groceries[i] = Grocery{
			Product: product.Product{
				Name:     products[i],
				Brand:    brands[i],
				Store:    stores[i],
				Category: "",
			},
			Amount:      amounts[i],
			HouseholdID: hid,
		}
	}

	err := h.Service.AddGroceries(c, groceries)

	if err != nil {
		c.AbortWithStatus(500)
		c.String(500, err.Error())
		return
	}

	c.Redirect(303, "/groceries")
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

	groceries, err := h.Service.List(c, sortBy, order, hid)
	if err != nil {
		c.AbortWithStatus(500)
		c.String(500, err.Error())
		return
	}

	var topProducts []string
	topProducts, err = h.Service.GetTopProducts(c, hid)

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

	h.Service.TogglePicked(c, id, hid)
	c.Redirect(302, "/groceries")
}

func (h *Handler) Add(c *gin.Context) {
	grocery := Grocery{
		Product: product.Product{
			Name:     c.PostForm("name"),
			Brand:    c.PostForm("brand"),
			Store:    c.PostForm("store"),
			Category: "",
		},
		Amount:      c.PostForm("amount"),
		Picked:      false,
		HouseholdID: c.GetInt("household_id"),
	}

	err := h.Service.AddGroceries(c, []Grocery{grocery})

	if err != nil {
		c.AbortWithStatus(500)
		c.String(500, err.Error())
		return
	}

	c.Redirect(302, "/groceries")
}

func (h *Handler) ExtractedView(c *gin.Context, groceries []Grocery) {
	data := gin.H{
		"Title":     "Extracted Groceries",
		"Groceries": groceries,
		"SaveURL":   "/groceries/extracted",
		"CancelURL": "/groceries",
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

	groceries, err := h.Service.SmartAdd(c, text)

	if err != nil {
		c.AbortWithStatus(500)
		c.String(500, err.Error())
		return
	}

	h.ExtractedView(c, groceries)
}

func (h *Handler) DeletePicked(c *gin.Context) {
	err := h.Service.DeletePicked(c, c.GetInt("household_id"))

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

	err = h.Service.Edit(c, groceries, c.GetInt("household_id"))

	if err != nil {
		c.AbortWithStatus(500)
		c.String(500, err.Error())
		return
	}

	c.Redirect(302, "/groceries")
}
