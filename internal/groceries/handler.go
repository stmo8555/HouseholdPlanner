package groceries

import (
	"strings"
	"strconv"
	"github.com/gin-gonic/gin"
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

	data := gin.H{
		"Title":     "Extracted Groceries",
		"Groceries": groceries,
	}

	c.HTML(200, "groceries_extraction.html", data)
}

func (h *Handler) AcceptExtractedGroceries(c *gin.Context) {
	products := c.PostFormArray("product")
	amounts := c.PostFormArray("amount")
	brands := c.PostFormArray("brand")
	stores := c.PostFormArray("store")

	hid := c.GetInt("household_id")
	groceries := make([]Grocery, len(products))
	for i := range len(groceries) {
		groceries[i] = Grocery{
			Product:     products[i],
			Amount:      amounts[i],
			Brand:       brands[i],
			Store:       stores[i],
			HouseholdID: hid,
		}
	}

	err := h.Service.AddGroceries(c, groceries)

	if err != nil {
		c.AbortWithStatus(500)
		c.String(500, err.Error())
		return
	}

	c.Redirect(302, "/groceries")
}

func (h *Handler) List(c *gin.Context) {
	hid := c.GetInt("household_id")

	groceries, err := h.Service.List(c, hid)
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
		"TopProducts": topProducts,
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
		Product:      c.PostForm("product"),
		Brand:        c.PostForm("brand"),
		Amount:       c.PostForm("amount"),
		Store:        c.PostForm("store"),
		Picked:       false,
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

	err = h.Service.Edit(c, groceries, c.GetInt("household_id") )

	if err != nil {
		c.AbortWithStatus(500)
		c.String(500, err.Error())
		return
	}

	c.Redirect(302, "/groceries")
}
