package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
	"net/http"

	"github.com/stmo8555/HouseholdPlanner/internal/groceries"
	"github.com/stmo8555/HouseholdPlanner/internal/login"
	"github.com/stmo8555/HouseholdPlanner/internal/recipes"
	"github.com/stmo8555/HouseholdPlanner/internal/todos"
)

type LayoutData struct {
	Title       string
	CurrentPath string
	Data        any
}

var conn *pgx.Conn

var loginService *login.Service
var todosService *todos.Service
var recipesService *recipes.Service
var groceriesService *groceries.Service

func main() {
	var err error

	conn, err = pgx.Connect(context.Background(), "postgres://Admin:Admin@localhost:5432/db?sslmode=disable")
	if err != nil {
		panic(err)
	}

	defer conn.Close(context.Background())

	r := gin.Default()
	r.LoadHTMLGlob("web/templates/*.html")
	r.Static("/static/", "web/static")

	loginService = &login.Service{Repo: &login.Repo{DB: conn, Sessions: make(map[string]login.Session)}}
	todosService = &todos.Service{Repo: &todos.Repo{DB: conn}}
	recipesService = &recipes.Service{Repo: &recipes.Repo{DB: conn}}
	groceriesService = &groceries.Service{Repo: &groceries.Repo{DB: conn}}

	auth := r.Group("/")
	auth.Use(login.AuthMiddleware(loginService))

	setupLogin(r)
	setupTodos(auth)
	setupRecipes(auth)
	setupGroceries(auth)
	auth.GET("/", indexHandleFunc)
	 
	// setup()
	// err = r.RunTLS(":8443", "raspis.crt", "raspis.key")
	err = r.Run(":8080")
	if err != nil {
		panic(err)
	}
}

func setupLogin(r *gin.Engine) {
	handler := &login.Handler{Service: loginService}
	r.GET("/login", handler.Login)
	r.POST("/login", handler.Authenticate)
	r.GET("/logout", handler.Logout)
}

func setupTodos(r *gin.RouterGroup) {
	handler := &todos.Handler{Service: todosService}

	r.GET("/todos", handler.List)
	r.POST("/todos/add", handler.Add)
	r.POST("/todos/done", handler.MarkDone)
	r.POST("/todos/undo", handler.MarkUnDone)

	todos.RunCleanup(context.Background(), todosService)
}

func setupRecipes(r *gin.RouterGroup) {
	handler := &recipes.Handler{Service: recipesService}

	r.GET("/recipes", handler.List)
	r.POST("/recipes/add", handler.Add)
}

func setupGroceries(r *gin.RouterGroup) {
	handler := &groceries.Handler{Service: groceriesService}

	r.GET("/groceries", handler.List)
	r.POST("/groceries", handler.TogglePicked)
	r.POST("/groceries/add", handler.Add)
	r.POST("/groceries/edit", handler.Edit)
	r.POST("/groceries/delete/picked", handler.DeletePicked)
	r.POST("/groceries/extract", handler.IngredientsFromRecipe)
	r.POST("/groceries/extracted", handler.AcceptExtractedGroceries)
}

func setup() {
	id := addUserRetreiveId("Steffo", "apa")
	hid := createHousehold(id, "la casa")
	id = addUserRetreiveId("Anna", "gurk")
	joinHousehold(id, hid)
	g1 := groceries.Grocery{Product: "skinka", HouseholdID: hid}
	g2 := groceries.Grocery{Product: "mjölk", HouseholdID: hid}
	g3 := groceries.Grocery{Product: "lök", HouseholdID: hid}
	g4 := groceries.Grocery{Product: "tomat", HouseholdID: hid}
	g5 := groceries.Grocery{Product: "vispgrädde", HouseholdID: hid}
	g6 := groceries.Grocery{Product: "bröd", HouseholdID: hid}
	g7 := groceries.Grocery{Product: "persilja", HouseholdID: hid}
	g8 := groceries.Grocery{Product: "linguini", HouseholdID: hid}
	g9 := groceries.Grocery{Product: "billys", HouseholdID: hid}
	g10 := groceries.Grocery{Product: "oatly", HouseholdID: hid}
	g11 := groceries.Grocery{Product: "bregott", HouseholdID: hid}
	gs := [...]groceries.Grocery{g1, g2, g3, g4, g5, g6, g7, g8, g9, g10, g11}

	for _, g := range gs {
		for range 5 {
			groceries.AddToHistory(conn, context.Background(), g)
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
	hid := c.GetInt("household_id")

	groceries, err := groceriesService.CountUnpicked(c, hid)

	if err != nil {
		panic(err)
	}

	repo := &todos.Repo{DB: conn}
	service := &todos.Service{Repo: repo}
	var todos int
	todos, err = service.Count(context.Background(), hid)

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
