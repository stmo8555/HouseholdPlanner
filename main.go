package main

import (
	_"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type LayoutData struct {
	Title string
	Data  any
}

type Grocery struct {
	Product, Brand, Unit, Store string
	Amount                      int
	Picked                      bool
}

var r *gin.Engine

func main() {
	r = gin.Default()
	r.LoadHTMLGlob("web/templates/*.html")

	r.Static("/static/", "web/static")
	r.GET("/login", loginHandleFunc)
	r.POST("/login", loginHandleFunc)
	r.GET("/logout", logoutHandlerFunc)

	auth := r.Group("/")
	auth.Use(AuthMiddleware())
	{
		auth.GET("/chores", choresHandleFunc)
		auth.GET("/groceries", groceriesHandleFunc)
		auth.GET("/", indexHandleFunc)
		auth.GET("/recipes", recipesHandleFunc)
	}
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

	uname := c.PostForm("uname")
	pwd := c.PostForm("pwd")

	if uname == "stefan" && pwd == "morin" {
		c.SetCookie("session_id", "gurkan", 0, "/", "", false, true)
		c.Redirect(302, "/")
	} else {
		c.HTML(http.StatusOK, "login.html", data)
	}
}

func logoutHandlerFunc(c *gin.Context) {
	c.SetCookie("session_id", "", -1, "/", "", false, true)
	c.Redirect(302, "/login")
}

func recipesHandleFunc(c *gin.Context) {
	data := LayoutData{Title: "Recipes", Data: nil}
	c.HTML(http.StatusOK, "recipes.html", data)
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session, err := c.Cookie("session_id")
		if err != nil || session != "gurkan" {
			c.AbortWithStatusJSON(401, gin.H{
				"error": "unauthorized",
			})
			return
		}

		c.Next()
	}
}
