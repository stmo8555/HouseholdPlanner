package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"

	"github.com/stmo8555/HouseholdPlanner/internal/grocery"
	"github.com/stmo8555/HouseholdPlanner/internal/home"
	"github.com/stmo8555/HouseholdPlanner/internal/login"
	"github.com/stmo8555/HouseholdPlanner/internal/recipe"
	"github.com/stmo8555/HouseholdPlanner/internal/todo"
)

var conn *pgx.Conn

var loginService *login.Service
var todosService *todo.Service
var recipesService *recipe.Service
var groceriesService *grocery.Service

func main() {
	var err error

	conn, err = pgx.Connect(context.Background(), dbDSN())
	if err != nil {
		panic(err)
	}

	foodMap, err := loadFoodMap("food_category_lookup.json")
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close(context.Background())

	loginService = &login.Service{Repo: &login.Repo{DB: conn, Sessions: make(map[string]login.Session)}}
	todosService = &todo.Service{Repo: &todo.Repo{DB: conn}}
	recipesService = &recipe.Service{Repo: &recipe.Repo{DB: conn}}
	groceriesService = &grocery.Service{Repo: &grocery.Repo{DB: conn}, lookUp map[string]string}

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

func getenv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func dbDSN() string {
	user := getenv("POSTGRES_USER", "Admin")
	password := getenv("POSTGRES_PASSWORD", "Admin")
	host := getenv("POSTGRES_HOST", "localhost")
	port := getenv("POSTGRES_PORT", "5432")
	name := getenv("POSTGRES_DB", "db")
	sslmode := getenv("POSTGRES_SSLMODE", "disable")

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user, password, host, port, name, sslmode,
	)
}

func loadFoodMap(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var m map[string]string
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}

	return m, nil
}
