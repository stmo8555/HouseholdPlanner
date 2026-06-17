package grocery

import (
	"errors"
	"fmt"
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

func NewHandler(s *Service, ingredient *ingredient.Extractor) *Handler {
	if s == nil || ingredient == nil {
		panic("nil service for handler")
	}

	return &Handler{
		service:             s,
		ingredientExtractor: ingredient,
	}
}

func (h *Handler) OverviewPage(c *gin.Context) {
	hid := c.GetInt("household_id")

	lists, err := h.service.GroceryLists(c, hid)
	if err != nil {
		c.String(500, err.Error())
		return
	}

	data := gin.H{"Lists": lists}
	data["Title"] = "Groceries"
	data["CurrentPath"] = c.Request.URL.Path

	c.HTML(200, "groceries.html", data)
}

func (h *Handler) ListPage(c *gin.Context) {
	hid := c.GetInt("household_id")

	groceryListID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		panic(err)
	}

	data, err := h.buildListViewData(c, groceryListID, hid)
	if err != nil {
		panic(err)
	}

	lists, err := h.service.GroceryLists(c, hid)
	if err != nil {
		panic(err)
	}

	var listName string
	for _, l := range lists {
		if l.GroceryList.ID == groceryListID {
			listName = l.GroceryList.Name
			break
		}
	}

	data["GroceryListID"] = groceryListID
	data["ListName"] = listName
	data["Lists"] = lists

	c.HTML(200, "groceries/list_page", data)
}

func (h *Handler) List(c *gin.Context) {
	groceryListID, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		panic(err)
	}

	h.RenderListPartial(c, groceryListID)
}

func (h *Handler) RenderListPartial(c *gin.Context, groceryListID int) {
	hid := c.GetInt("household_id")

	data, err := h.buildListViewData(c, groceryListID, hid)
	if err != nil {
		panic(err)
	}

	lists, err := h.service.GroceryLists(c, hid)
	if err != nil {
		panic(err)
	}

	var listName string
	for _, l := range lists {
		if l.GroceryList.ID == groceryListID {
			listName = l.GroceryList.Name
			break
		}
	}

	data["GroceryListID"] = groceryListID
	data["ListName"] = listName

	c.HTML(200, "groceries/list", data)
}

func (h *Handler) CreateList(c *gin.Context) {
	hid := c.GetInt("household_id")
	name := c.PostForm("name")

	if strings.TrimSpace(name) == "" {
		c.AbortWithStatus(500)
		c.String(500, "Empty text")
		return
	}

	err := h.service.CreateList(c.Request.Context(), name, hid)

	if err != nil {
		panic(err)
	}

	lists, err := h.service.GroceryLists(c, hid)
	if err != nil {
		panic(err)
	}

	data := gin.H{"Lists": lists}

	c.HTML(200, "groceries/overview_page", data)
}

func (h *Handler) UpdateGroceryList(c *gin.Context) {
	hid := c.GetInt("household_id")
	newName := c.PostForm("new-name")

	if strings.TrimSpace(newName) == "" {
		panic(errors.New("No new name"))
	}

	groceryListID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		panic(err)
	}

	err = h.service.UpdateGroceryList(c.Request.Context(), newName, groceryListID, hid)
	if err != nil {
		panic(err)
	}

	lists, err := h.service.GroceryLists(c, hid)
	if err != nil {
		panic(err)
	}

	data := gin.H{"Lists": lists}

	c.HTML(200, "groceries/overview_page", data)
}

func (h *Handler) DeleteGroceryList(c *gin.Context) {
	hid := c.GetInt("household_id")

	groceryListID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		panic(err)
	}

	err = h.service.DeleteGroceryList(c.Request.Context(), groceryListID, hid)
	if err != nil {
		panic(err)
	}

	lists, err := h.service.GroceryLists(c, hid)
	if err != nil {
		panic(err)
	}

	data := gin.H{"Lists": lists}

	c.HTML(200, "groceries/overview_page", data)
}

func (h *Handler) TransferGroceryList(c *gin.Context) {
	hid := c.GetInt("household_id")

	groceryListID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		panic(err)
	}

	groceryListTargetID, err := strconv.Atoi(c.PostForm("grocery-list-target-id"))
	if err != nil {
		panic(err)
	}

	err = h.service.TransferGroceries(c.Request.Context(), groceryListTargetID, groceryListID, hid)
	if err != nil {
		panic(err)
	}

	lists, err := h.service.GroceryLists(c, hid)
	if err != nil {
		panic(err)
	}

	data := gin.H{"Lists": lists}

	c.HTML(200, "groceries/overview_page", data)
}

func (h *Handler) CreateGrocery(c *gin.Context) {
	hid := c.GetInt("household_id")

	groceryListID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		panic(err)
	}

	ingredients := make([]ingredient.Ingredient, 0, 1)
	ingredients = append(ingredients, ingredient.Ingredient{
		Product: product.Product{
			Name:     c.PostForm("name"),
			Brand:    c.PostForm("brand"),
			Category: "",
		},
		Amount: c.PostForm("amount"),
	})

	err = h.service.CreateGroceries(c, ingredients, groceryListID, hid)

	if err != nil {
		panic(err)
	}

	data, err := h.buildListViewData(c, groceryListID, hid)

	if err != nil {
		panic(err)
	}

	data["GroceryListID"] = groceryListID
	data["OOB"] = true
	c.HTML(200, "groceries/add_response", data)
}

func (h *Handler) EditGroceryForm(c *gin.Context) {
	//   listID, err := strconv.Atoi(c.Param("listId"))
	//   if err != nil {
	// panic(err)
	//   }
	hid := c.GetInt("household_id")

	itemID, err := strconv.Atoi(c.Param("itemId"))
	if err != nil {
		panic(err)
	}

	item, err := h.service.Grocery(c.Request.Context(), itemID, hid)
	if err != nil {
		panic(err)
	}
	
	lists, err := h.service.GroceryLists(c, hid)
	if err != nil {
		panic(err)
	}

	data := gin.H{
		"Lists": lists,
		"Grocery": item,
	}

	c.HTML(200, "groceries/edit_modal", data)
}

func (h *Handler) UpdateGrocery(c *gin.Context) {
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

	groceryListID, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		panic(err)
	}

	itemID, err := strconv.Atoi(c.Param("itemId"))

	if err != nil {
		panic(err)
	}

	err = h.service.UpdateGrocery(c, ing, itemID, hid)

	if err != nil {
		panic(err)
	}

	h.RenderListPartial(c, groceryListID)
}

func (h *Handler) MoveGrocery(c *gin.Context) {
	hid := c.GetInt("household_id")

	groceryListID, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		panic(err)
	}

	itemID, err := strconv.Atoi(c.Param("itemId"))

	if err != nil {
		panic(err)
	}

	groceryListTargetID, err := strconv.Atoi(c.PostForm("grocery-list-target-id"))
	if err != nil {
		panic(err)
	}

	err = h.service.MoveGrocery(c, itemID, groceryListTargetID, hid)
	if err != nil {
		panic(err)
	}

	h.RenderListPartial(c, groceryListID)
}

func (h *Handler) DeleteGrocery(c *gin.Context) {
	hid := c.GetInt("household_id")
	groceryListID, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		panic(err)
	}

	itemID, err := strconv.Atoi(c.Param("itemId"))

	if err != nil {
		panic(err)
	}

	err = h.service.DeleteGrocery(c, itemID, hid)

	if err != nil {
		panic(err)
	}

	h.RenderListPartial(c, groceryListID)
}

func (h *Handler) TogglePicked(c *gin.Context) {
	hid := c.GetInt("household_id")
	groceryListID, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		panic(err)
	}

	itemID, err := strconv.Atoi(c.Param("itemId"))

	if err != nil {
		panic(err)
	}

	if err := h.service.TogglePicked(c, itemID, hid); err != nil {
		panic(err)
	}

	h.RenderListPartial(c, groceryListID)
}

func (h *Handler) DeletePicked(c *gin.Context) {
	hid := c.GetInt("household_id")

	groceryListID, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		panic(err)
	}

	err = h.service.DeletePicked(c, groceryListID, hid)

	if err != nil {
		panic(err)
	}

	h.RenderListPartial(c, groceryListID)
}

func (h *Handler) SmartAdd(c *gin.Context) {
	text := c.PostForm("text")

	groceryListID, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		panic(err)
	}

	if strings.TrimSpace(text) == "" {
		panic("no text")
	}

	groceries, err := h.service.ParseGroceries(c, text)

	if err != nil {
		panic(err)
	}

	h.ExtractReviewPage(c, groceries, groceryListID)
}

func (h *Handler) ExtractFromRecipe(c *gin.Context) {
	link := c.PostForm("link")

	if strings.TrimSpace(link) == "" {
		panic("no link")
	}

	groceryListID, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		panic(err)
	}

	ingredients, err := h.ingredientExtractor.FromRecipeURL(c, link)

	if err != nil {
		panic(err)
	}

	h.ExtractReviewPage(c, ingredients, groceryListID)
}

func (h *Handler) SaveExtracted(c *gin.Context) {
	hid := c.GetInt("household_id")

	groceryListID, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		panic(err)
	}

	redirectPath := c.PostForm("redirect-path")

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

	err = h.service.CreateGroceries(c, ingredients, groceryListID, hid)

	if err != nil {
		panic(err)
	}

	if !strings.HasPrefix(redirectPath, "/") {
		panic(errors.New("Redirect variable does not start with /. Value: " + redirectPath).Error())
	}

	c.Redirect(303, redirectPath)
}

func (h *Handler) ExtractReviewPage(c *gin.Context, ingredients []ingredient.Ingredient, groceryListID int) {
	path := fmt.Sprintf("/groceries/lists/%d", groceryListID)
	data := gin.H{
		"Ingredients":   ingredients,
		"CancelURL":     path,
		"RedirectPath":  path,
		"GroceryListID": groceryListID,
	}

	c.HTML(200, "groceries/extract_review_page", data)
}

func (h *Handler) buildListViewData(c *gin.Context, groceryListID, hid int) (gin.H, error) {
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

	groceries, err := h.service.GroceriesView(c, sortBy, order, groceryListID, hid)
	if err != nil {
		return nil, err
	}

	topProducts, err := h.service.TopProducts(c, hid)
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
		"OOB":         false,
	}, nil
}
