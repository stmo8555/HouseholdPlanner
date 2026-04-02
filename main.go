package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/stmo8555/HouseholdPlanner/pages"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

type LayoutData struct {
	Title       string
	CurrentPath string
	Data        any
}

var conn *pgx.Conn

func main() {
	var err error

	conn, err = pgx.Connect(context.Background(), "postgres://Admin:Admin@localhost:5432/db?sslmode=disable")
	if err != nil {
		panic(err)
	}

	defer conn.Close(context.Background())

	sessions := make(map[string]*pages.Session, 2)
	r := gin.Default()
	r.LoadHTMLGlob("web/templates/*.html")

	r.Static("/static/", "web/static")
	r.GET("/login", pages.LoginHandleFunc)
	r.POST("/login", func(c *gin.Context) { pages.AuthHandleFunc(c, conn, sessions) })
	r.GET("/logout", func(c *gin.Context) { pages.LogoutHandlerFunc(c, sessions) })

	auth := r.Group("/")
	auth.Use(pages.AuthMiddleware(sessions))
	auth.GET("/chores", choresHandleFunc)
	auth.GET("/groceries", func(c *gin.Context) { pages.GroceriesHandleFunc(c, conn) })
	auth.POST("/groceries", func(c *gin.Context) { pages.GroceriesPickHandleFunc(c, conn) })
	auth.POST("/groceries/add", func(c *gin.Context) { pages.GroceriesAddHandleFunc(c, conn) })
	auth.POST("/groceries/edit", func(c *gin.Context) { pages.GroceriesEditHandleFunc(c, conn) })
	auth.POST("/groceries/delete/picked", func(c *gin.Context) { pages.GroceriesDeletePickedHandleFunc(c, conn) })
	auth.POST("/groceries/extract", func(c *gin.Context) { pages.GroceriesExtractFromRecipeHandleFunc(c) })
	auth.POST("/groceries/extracted", func(c *gin.Context) { pages.HandleAcceptGroceries(c, conn) })
	auth.GET("/", indexHandleFunc)
	auth.GET("/recipes", recipesHandleFunc)

	// setup()

	// ai("https://www.ica.se/recept/flaskfilegryta-med-champinjoner-724256/")
	// ai("https://www.koket.se/pasta-salsiccia-classico")

	err = r.RunTLS(":8443", "raspis.crt", "raspis.key")
	if err != nil {
		panic(err)
	}
}

func setup() {
	id := addUserRetreiveId("Steffo", "apa")
	hid := createHousehold(id, "la casa")
	id = addUserRetreiveId("Anna", "gurk")
	joinHousehold(id, hid)
}

func joinHousehold(user_id, hid int) {
	_, err := conn.Exec(context.Background(),
		`INSERT INTO household_members (user_id, household_id, role)
     VALUES ($1,$2,'guru')`,
		user_id, hid,
	)

	if err != nil {
		panic(err)
	}
}

func createHousehold(owner int, name string) int {
	tx, err := conn.Begin(context.Background())
	if err != nil {
		panic(err)
	}

	defer tx.Rollback(context.Background()) // safe even after commit
	var hid int

	err = tx.QueryRow(context.Background(),
		`INSERT INTO households (name, created_by)
     VALUES ($1,$2)
     RETURNING id`,
		name, owner,
	).Scan(&hid)

	_, err = tx.Exec(context.Background(),
		`INSERT INTO household_members (user_id, household_id, role)
     VALUES ($1,$2,'owner')`,
		owner, hid,
	)

	err = tx.Commit(context.Background())

	if err != nil {
		panic(err)
	}

	return hid
}

func addUserRetreiveId(uname, pwd string) int {
	sql := `INSERT INTO users (username,pwd) VALUES ($1,$2) RETURNING id;`
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	var id int
	err := conn.QueryRow(context.Background(), sql, uname, hashedPassword).Scan(&id)

	if err != nil {
		panic("failed to create user: " + err.Error())
	}

	return id
}

func deleteUser(conn *pgx.Conn) {
	sql := `DELETE FROM users WHERE username=$1;`
	_, err := conn.Exec(context.Background(), sql, "Steffo")

	if err != nil {
		panic("Failed to to delete user: " + err.Error())
	}
}

func recipesHandleFunc(c *gin.Context) {
	data := LayoutData{Title: "Recipes", CurrentPath: c.Request.URL.Path, Data: nil}
	c.HTML(http.StatusOK, "recipes.html", data)
}

func choresHandleFunc(c *gin.Context) {
	data := LayoutData{Title: "Chores", CurrentPath: c.Request.URL.Path, Data: nil}
	fmt.Println(data.CurrentPath)
	c.HTML(http.StatusOK, "chores.html", data)
}

func indexHandleFunc(c *gin.Context) {
	data := LayoutData{Title: "Home", CurrentPath: c.Request.URL.Path, Data: nil}
	fmt.Println("-------------------- " + c.Request.URL.Path)
	c.HTML(http.StatusOK, "index.html", data)
}
