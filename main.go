package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/stmo8555/HouseholdPlanner/internal/ai"
	"github.com/stmo8555/HouseholdPlanner/internal/grocery"
	"github.com/stmo8555/HouseholdPlanner/internal/ingredient"
	"github.com/stmo8555/HouseholdPlanner/internal/login"
	"github.com/stmo8555/HouseholdPlanner/internal/product"
	"github.com/stmo8555/HouseholdPlanner/internal/recipe"
	"github.com/stmo8555/HouseholdPlanner/internal/todo"
)

var aiService *ai.Service
var loginService *login.Service
var todoService *todo.Service
var productService *product.Service
var ingredientExtractor *ingredient.Extractor
var recipeService *recipe.Service
var groceryService *grocery.Service

func main() {
	pool, err := pgxpool.New(context.Background(), dbDSN())

	if err != nil {
		panic(err)
	}

	foodMap, err := loadFoodMap("./food_category_lookup.json")

	if err != nil {
		panic(err)
	}

	defer pool.Close()

	aiService = ai.CreateService(ai.CreateClient())
	productService = product.CreateService(product.CreateRepo(pool), aiService, foodMap)
	ingredientExtractor = ingredient.CreateExtractor(aiService)
	loginService = login.CreateService(login.CreateRepo(pool))
	todoService = todo.CreateService(todo.CreateRepo(pool), aiService)
	groceryService = grocery.CreateService(grocery.CreateRepo(pool), productService, ingredientExtractor)
	recipeService = recipe.CreateService(recipe.CreateRepo(pool), productService, ingredientExtractor)

	tmpl := template.Must(template.ParseGlob("web/templates/*.html"))
	tmpl = template.Must(tmpl.ParseGlob("web/templates/partial/*.html"))

	r := gin.Default()
	// r.LoadHTMLGlob("web/templates/*.html")
	r.SetHTMLTemplate(tmpl)
	r.Static("/static/", "web/static")
	setupLogin(r)

	auth := r.Group("/")
	auth.Use(login.AuthMiddleware(loginService))

	setupTodos(auth)
	// setupRecipes(auth)
	setupGroceries(auth)
	// setupHome(auth)

	err = r.Run(":8080")
	if err != nil {
		panic(err)
	}
}

func setupLogin(r *gin.Engine) {
	handler := login.CreateHandler(loginService)

	r.GET("/login", handler.Login)
	r.POST("/login", handler.Authenticate)
	r.GET("/logout", handler.Logout)

	login.RunCleanup(context.Background(), loginService)
}

func setupTodos(r *gin.RouterGroup) {
	handler := todo.CreateHandler(todoService)

	r.GET("/todos", handler.List)
	r.POST("/todos/add", handler.Add)
	r.POST("/todos/smartadd", handler.SmartAdd)
	r.POST("/todos/done", handler.MarkDone)
	r.POST("/todos/undo", handler.MarkUnDone)

	todo.RunCleanup(context.Background(), todoService)
	todo.ScheduleRepeats(context.Background(), todoService)
}

func setupRecipes(r *gin.RouterGroup) {
	handler := recipe.CreateHandler(recipeService, groceryService)
	r.GET("/recipes", handler.List)
	r.POST("/recipes/add", handler.Add)
	r.POST("/recipes/extract", handler.IngredientsFromRecipe)
}

func setupGroceries(r *gin.RouterGroup) {
	handler := grocery.CreateHandler(groceryService, ingredientExtractor)

	r.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/groceries")
	})

	r.GET("/groceries", handler.List)
	r.GET("/groceries/list", handler.ListPartial)
	r.POST("/groceries", handler.TogglePicked)
	r.POST("/groceries/add", handler.Add)
	r.POST("/groceries/smartadd", handler.SmartAdd)
	r.POST("/groceries/edit", handler.Edit)
	r.POST("/groceries/delete", handler.Delete)
	r.POST("/groceries/delete/picked", handler.DeletePicked)
	r.POST("/groceries/extract", handler.IngredientsFromRecipe)
	r.POST("/groceries/extracted", handler.AcceptExtractedGroceries)
}

// func setupHome(r *gin.RouterGroup) {
// 	handler := &home.Handler{
// 		GroceriesService: groceryService,
// 		LoginService:     loginService,
// 		RecipesService:   recipeService,
// 		TodosService:     todosService,
// 	}
//
// 	r.GET("/", handler.Index)
// 	r.GET("/home", handler.Index)
// 	// r.POST("/home/add/grocery", handler.AddGrocery)
// 	r.POST("/home/add/recipe", handler.AddRecipe)
// 	r.POST("/home/ai", handler.AI)
// }

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
