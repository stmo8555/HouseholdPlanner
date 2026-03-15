package main

import (
	"net/http"
    "github.com/gin-gonic/gin"
)

type LayoutData struct {
    Title string
    Data  any
}

type Grocery struct {
	Product, Brand, Unit, Store string
	Amount int
	Picked bool
}

var r *gin.Engine

func main() {
	r = gin.Default()
	r.LoadHTMLGlob("../../web/templates/*.html")

	r.GET("/chores", choresHandleFunc)
	r.GET("/groceries", groceriesHandleFunc)
	r.GET("/", indexHandleFunc)
	r.GET("/login", loginHandleFunc)
	r.POST("/login", loginHandleFunc)
	r.GET("/logout", loginHandleFunc)
	r.GET("/recipes", recipesHandleFunc)
	r.Static("/static/", "../../web/static")
	
	r.Run()
}

func choresHandleFunc(c *gin.Context) {
	data := LayoutData{Title: "Chores", Data: nil}
	c.HTML(http.StatusOK, "chores.html", data)
}

func groceriesHandleFunc(c *gin.Context) {
	groceries := make([]Grocery, 0, 5)
	groceries = append(groceries, Grocery{Amount: 5, Product: "Mjölk", Brand: "Arla", Unit: "kg", Store: "ICA", Picked: false})
	groceries = append(groceries, Grocery{Amount: 1, Product: "Gurk", Brand: "Arla", Unit: "kg", Store: "ICA", Picked: true})
	data := LayoutData{Title: "Groceries", Data: groceries}
	c.HTML(http.StatusOK, "groceries.html", data)
}

func indexHandleFunc(c *gin.Context) {
	data := LayoutData{Title: "Home", Data: nil}
	c.HTML(http.StatusOK, "index.html", data)
}

func loginHandleFunc(c *gin.Context) {
	data := LayoutData{Title: "login", Data: nil}
	c.HTML(http.StatusOK, "login.html", data)
}

func recipesHandleFunc(c *gin.Context) {
	data := LayoutData{Title: "Recipes", Data: nil}
	c.HTML(http.StatusOK, "recipes.html", data)
}
