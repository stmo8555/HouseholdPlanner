package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/stmo8555/HouseholdPlanner/internal/ai"
	"github.com/stmo8555/HouseholdPlanner/internal/grocery"
	"github.com/stmo8555/HouseholdPlanner/internal/household"
	"github.com/stmo8555/HouseholdPlanner/internal/ingredient"
	"github.com/stmo8555/HouseholdPlanner/internal/login"
	"github.com/stmo8555/HouseholdPlanner/internal/notification"
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
var householdService *household.Service

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

	aiService = ai.NewService(ai.NewClient())
	productService = product.NewService(product.NewRepo(pool), aiService, foodMap)
	ingredientExtractor = ingredient.NewExtractor(aiService)
	loginService = login.NewService(login.NewRepo(pool))
	todoService = todo.CreateService(todo.CreateRepo(pool), aiService)
	groceryService = grocery.NewService(grocery.NewRepo(pool), productService, ingredientExtractor)
	recipeService = recipe.NewService(recipe.NewRepo(pool), productService, ingredientExtractor)
	householdService = household.NewService(household.NewRepo(pool))

	tmpl := template.Must(parseTemplates("web/templates"))

	r := gin.Default()
	r.TrustedPlatform = gin.PlatformCloudflare
	r.SetHTMLTemplate(tmpl)
	r.Static("/static/", "web/static")

	loginHandler := login.NewHandler(loginService)
	setupLogin(r, loginHandler)

	auth := r.Group("/")
	auth.Use(login.AuthMiddleware(loginService))

	// Authenticated but not yet tied to a household — onboarding only.
	auth.GET("/welcome", loginHandler.WelcomeView)
	auth.POST("/welcome", loginHandler.SetupHousehold)

	hh := auth.Group("/")
	hh.Use(login.RequireHousehold())

	// setupTodos(hh)
	// setupRecipes(hh)
	setupGroceries(hh)
	setupNotifications(hh)
	setupHousehold(hh)
	// setupHome(hh)

	err = r.Run(":" + getenv("PORT", "8080"))
	if err != nil {
		panic(err)
	}
}

func setupHousehold(r *gin.RouterGroup) {
	handler := household.NewHandler(householdService)
	r.POST("/settings/code/regenerate", handler.RegenerateHouseholdCode)
	r.POST("/settings/invite", handler.GenerateInviteToken)
	r.POST("/settings/members/:id/promote", handler.PromoteMember)
	r.POST("/settings/members/:id/remove", handler.RemoveMember)
	r.POST("/settings/leave", handler.LeaveHousehold)
	r.POST("/settings/delete", handler.DeleteHousehold)
	r.POST("/settings/account/delete", handler.DeleteAccount)
	r.GET("/settings", handler.Settings)
}

func setupNotifications(r *gin.RouterGroup) {
	handler := notification.NewHandler(householdService)

	r.GET("/notifications/household-version", handler.Check)
	r.GET("/notifications/ack", handler.Ack)
}

func setupLogin(r *gin.Engine, handler *login.Handler) {
	r.GET("/login", handler.Login)
	r.POST("/login", handler.Authenticate)
	r.POST("/logout", handler.Logout)

	r.GET("/register", handler.RegisterView)
	r.POST("/register/:token", handler.Register)

	login.RunCleanup(context.Background(), loginService)
}

func setupTodos(r *gin.RouterGroup) {
	handler := todo.NewHandler(todoService)

	r.GET("/todos", handler.List)
	r.POST("/todos/add", handler.Add)
	r.POST("/todos/smartadd", handler.SmartAdd)
	r.POST("/todos/done", handler.MarkDone)
	r.POST("/todos/undo", handler.MarkUnDone)
	r.POST("/todos/delete", handler.Delete)

	todo.RunCleanup(context.Background(), todoService)
	todo.ScheduleRepeats(context.Background(), todoService)
}

func setupRecipes(r *gin.RouterGroup) {
	handler := recipe.NewHandler(recipeService, groceryService)
	r.GET("/recipes", handler.List)
	r.POST("/recipes/add", handler.Add)
	r.POST("/recipes/extract", handler.IngredientsFromRecipe)
}

func setupGroceries(r *gin.RouterGroup) {
	handler := grocery.NewHandler(groceryService, householdService, ingredientExtractor)

	r.GET("/", func(c *gin.Context) { c.Redirect(302, "/groceries") })
	r.GET("/groceries", handler.OverviewPage)

	// Grocery lists
	r.POST("/groceries/lists", handler.CreateList)
	r.GET("/groceries/lists/:id/edit", handler.EditGroceryListForm)
	r.GET("/groceries/lists/:id", handler.ListPage)
	r.PATCH("/groceries/lists/:id", handler.UpdateGroceryList)
	r.DELETE("/groceries/lists/:id", handler.DeleteGroceryList)
	r.POST("/groceries/lists/:id/transfer", handler.TransferGroceryList)

	// Items within a list
	r.GET("/groceries/lists/:id/items", handler.List)
	r.POST("/groceries/lists/:id/items", handler.CreateGrocery)
	r.DELETE("/groceries/lists/:id/items/picked", handler.DeletePicked)
	r.POST("/groceries/lists/:id/smart-add", handler.SmartAdd)
	r.POST("/groceries/lists/:id/extract", handler.ExtractFromRecipe)
	r.POST("/groceries/lists/:id/extracted", handler.SaveExtracted)
	r.PATCH("/groceries/lists/:id/items/:itemId", handler.UpdateGrocery)
	r.PATCH("/groceries/lists/:id/items/:itemId/move", handler.MoveGrocery)
	r.DELETE("/groceries/lists/:id/items/:itemId", handler.DeleteGrocery)
	r.PATCH("/groceries/lists/:id/items/:itemId/picked", handler.TogglePicked)
	r.GET("/groceries/lists/:id/items/:itemId/edit", handler.EditGroceryForm)
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

func parseTemplates(patternRoot string) (*template.Template, error) {
	tmpl := template.New("")

	err := filepath.WalkDir(patternRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || filepath.Ext(path) != ".html" {
			return nil
		}

		_, err = tmpl.ParseFiles(path)
		return err
	})

	if err != nil {
		return nil, err
	}

	return tmpl, nil
}
