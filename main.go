package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/stmo8555/HouseholdPlanner/internal/grocery"
	"github.com/stmo8555/HouseholdPlanner/internal/home"
	"github.com/stmo8555/HouseholdPlanner/internal/login"
	"github.com/stmo8555/HouseholdPlanner/internal/recipe"
	"github.com/stmo8555/HouseholdPlanner/internal/todo"
)

type LayoutData struct {
	Title       string
	CurrentPath string
	Data        any
}

var conn *pgx.Conn

var loginService *login.Service
var todosService *todo.Service
var recipesService *recipe.Service
var groceriesService *grocery.Service

func main() {
	var err error

	conn, err = pgx.Connect(context.Background(), "postgres://Admin:Admin@db:5432/db?sslmode=disable")
	if err != nil {
		panic(err)
	}

	defer conn.Close(context.Background())

	loginService = &login.Service{Repo: &login.Repo{DB: conn, Sessions: make(map[string]login.Session)}}
	todosService = &todo.Service{Repo: &todo.Repo{DB: conn}}
	recipesService = &recipe.Service{Repo: &recipe.Repo{DB: conn}}
	groceriesService = &grocery.Service{Repo: &grocery.Repo{DB: conn}}

	r := gin.Default()
	r.LoadHTMLGlob("web/templates/*.html")
	r.Static("/static/", "web/static")
	setupLogin(r)

	auth := r.Group("/")
	auth.Use(login.AuthMiddleware(loginService))

	setupTodos(auth)
	setupRecipes(auth)
	setupGroceries(auth)
	setupHome(auth)

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
	handler := &todo.Handler{Service: todosService}

	r.GET("/todos", handler.List)
	r.POST("/todos/add", handler.Add)
	r.POST("/todos/done", handler.MarkDone)
	r.POST("/todos/undo", handler.MarkUnDone)

	todo.RunCleanup(context.Background(), todosService)
}

func setupRecipes(r *gin.RouterGroup) {
	handler := &recipe.Handler{Service: recipesService}

	r.GET("/recipes", handler.List)
	r.POST("/recipes/add", handler.Add)
}

func setupGroceries(r *gin.RouterGroup) {
	handler := &grocery.Handler{Service: groceriesService}

	r.GET("/groceries", handler.List)
	r.POST("/groceries", handler.TogglePicked)
	r.POST("/groceries/add", handler.Add)
	r.POST("/groceries/edit", handler.Edit)
	r.POST("/groceries/delete/picked", handler.DeletePicked)
	r.POST("/groceries/extract", handler.IngredientsFromRecipe)
	r.POST("/groceries/extracted", handler.AcceptExtractedGroceries)
}

func setupHome(r *gin.RouterGroup) {
	handler := &home.Handler{
		GroceriesService: groceriesService,
		LoginService:     loginService,
		RecipesService:   recipesService,
		TodosService:     todosService,
	}

	r.GET("/", handler.Index)
	r.GET("/home", handler.Index)
	r.POST("/home/add/grocery", handler.AddGrocery)
	r.POST("/home/add/recipe", handler.AddRecipe)
	r.POST("/home/ai", handler.AI)
}

func setup() {
	id := addUserRetreiveId("Steffo", "apa")
	hid := createHousehold(id, "la casa")
	id = addUserRetreiveId("Anna", "gurk")
	joinHousehold(id, hid)
	g1 := grocery.Grocery{Product: "skinka", HouseholdID: hid}
	g2 := grocery.Grocery{Product: "mjölk", HouseholdID: hid}
	g3 := grocery.Grocery{Product: "lök", HouseholdID: hid}
	g4 := grocery.Grocery{Product: "tomat", HouseholdID: hid}
	g5 := grocery.Grocery{Product: "vispgrädde", HouseholdID: hid}
	g6 := grocery.Grocery{Product: "bröd", HouseholdID: hid}
	g7 := grocery.Grocery{Product: "persilja", HouseholdID: hid}
	g8 := grocery.Grocery{Product: "linguini", HouseholdID: hid}
	g9 := grocery.Grocery{Product: "billys", HouseholdID: hid}
	g10 := grocery.Grocery{Product: "oatly", HouseholdID: hid}
	g11 := grocery.Grocery{Product: "bregott", HouseholdID: hid}
	gs := [...]grocery.Grocery{g1, g2, g3, g4, g5, g6, g7, g8, g9, g10, g11}

	for _, g := range gs {
		for range 5 {
			grocery.AddToHistory(conn, context.Background(), g)
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
