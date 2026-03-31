package main

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
	"github.com/stmo8555/HouseholdPlanner/pages"
	"golang.org/x/crypto/bcrypt"
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

func ai(link string) {

	resp, err := http.Get(link)
	body, _ := io.ReadAll(resp.Body)

	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	client := openai.NewClient()

	templatePrompt := `You are an expert recipe parser. I will provide the full HTML content of a recipe page. Extract all ingredients from the HTML. 

Requirements:
1. The recipe may be in Swedish or English. Extract ingredients regardless of language.
2. Return only valid JSON in this format:
[
  { "Product": "ingredient name", "Amount": "amount with units" }
]
3. Use the exact units and quantities as written in the HTML (e.g., g, dl, msk, tsk). Include ranges and "to taste"/"efter smak" exactly as written.
4. Include only ingredients. Ignore instructions, steps, preparation methods, notes, optional tips, or serving suggestions.
5. If an ingredient has no specific quantity, use the value "to taste" or "efter smak" exactly as it appears in the HTML.
6. Output JSON only. Do not include any extra text, explanation, or markdown.
7. If the HTML contains extra text, ads, or unrelated content, ignore it and focus only on the recipe ingredients.
8. Be robust to messy HTML. Parse only the meaningful ingredient list.

HTML content:
$1`

	question := fmt.Sprintf(templatePrompt, string(body))
	ai_resp, ai_err := client.Responses.New(ctx, responses.ResponseNewParams{
		Input: responses.ResponseNewParamsInputUnion{OfString: openai.String(question)},
		Model: openai.ChatModelGPT5Mini,
	})

	if ai_err != nil {
		panic(ai_err)
	}

	println(ai_resp.OutputText())
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
