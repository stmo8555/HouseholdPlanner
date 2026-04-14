package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/robfig/cron/v3"
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
	auth.GET("/todos", func(c *gin.Context) { pages.List(c, conn) })

	auth.POST("/todos/done", func(c *gin.Context) { pages.Done(c, conn) })
	auth.POST("/todos/undo", func(c *gin.Context) { pages.Undo(c, conn) })
	auth.POST("/todos/add", func(c *gin.Context) { pages.Add(c, conn) })
	auth.GET("/groceries", func(c *gin.Context) { pages.GroceriesHandleFunc(c, conn) })
	auth.POST("/groceries", func(c *gin.Context) { pages.GroceriesPickHandleFunc(c, conn) })
	auth.POST("/groceries/add", func(c *gin.Context) { pages.GroceriesAddHandleFunc(c, conn) })
	auth.POST("/groceries/edit", func(c *gin.Context) { pages.GroceriesEditHandleFunc(c, conn) })
	auth.POST("/groceries/delete/picked", func(c *gin.Context) { pages.GroceriesDeletePickedHandleFunc(c, conn) })
	auth.POST("/groceries/extract", func(c *gin.Context) { pages.GroceriesExtractFromRecipeHandleFunc(c) })
	auth.POST("/groceries/extracted", func(c *gin.Context) { pages.HandleAcceptGroceries(c, conn) })
	auth.GET("/recipes", func(c *gin.Context) { pages.RecipesHandleFunc(c, conn) })
	auth.POST("/recipes/add", func(c *gin.Context) { pages.RecipesAddHandleFunc(c, conn) })
	auth.GET("/", indexHandleFunc)
	// setup()

	c := cron.New()
	c.AddFunc("0 * * * *", func() {
		pages.RemoveTodos(conn)
	})
	c.Start()

	// err = r.RunTLS(":8443", "raspis.crt", "raspis.key")
	err = r.Run(":8080")
	if err != nil {
		panic(err)
	}
}

func setup() {
	id := addUserRetreiveId("Steffo", "apa")
	hid := createHousehold(id, "la casa")
	id = addUserRetreiveId("Anna", "gurk")
	joinHousehold(id, hid)
	g1 := pages.Grocery{Product: "skinka", Household_id: hid}
	g2 := pages.Grocery{Product: "mjölk", Household_id: hid}
	g3 := pages.Grocery{Product: "lök", Household_id: hid}
	g4 := pages.Grocery{Product: "tomat", Household_id: hid}
	g5 := pages.Grocery{Product: "vispgrädde", Household_id: hid}
	g6 := pages.Grocery{Product: "bröd", Household_id: hid}
	g7 := pages.Grocery{Product: "persilja", Household_id: hid}
	g8 := pages.Grocery{Product: "linguini", Household_id: hid}
	g9 := pages.Grocery{Product: "billys", Household_id: hid}
	g10 := pages.Grocery{Product: "oatly", Household_id: hid}
	g11 := pages.Grocery{Product: "bregott", Household_id: hid}
	groceries := [...]pages.Grocery{g1, g2, g3, g4, g5, g6, g7, g8, g9, g10, g11}
	for _, g := range groceries {
		for range 10 {
			pages.AddToHistory(conn, g)
		}
	}
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

func indexHandleFunc(c *gin.Context) {
	hid, ok := c.Get("household_id")

	if !ok {
		panic("failed to get household id from context")
	}

	groceries, err := pages.AmountOfUnpickedGroceries(context.Background(), conn, hid.(int))

	if err != nil {
		panic(err)
	}

	var todos int
	todos, err = pages.AmountOfTodos(conn, hid.(int))

	if err != nil {
		panic(err)
	}

	data := gin.H{
		"Title":       "Home",
		"CurrentPath": c.Request.URL.Path,
		"Todos":       todos,
		"Groceries":   groceries,
	}

	c.HTML(http.StatusOK, "index.html", data)
}
