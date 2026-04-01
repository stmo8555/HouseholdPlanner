package pages

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
	"io"
)

type Grocery struct {
	Id           int    `json:"id"`
	Product      string `json:"product"`
	Amount       string `json:"amount"`
	Brand        string `json:"brand"`
	Store        string `json:"store"`
	Household_id int
	Picked       bool
}

type groceries struct {
	Picked, NotPicked []Grocery
}

func ai(link string) string {

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
		Model: openai.ChatModelGPT5_2,
	})

	if ai_err != nil {
		panic(ai_err)
	}

	return ai_resp.OutputText()
}

func GroceriesExtractFromRecipeHandleFunc(c *gin.Context) {
	link := c.PostForm("link")
	received := ai(link)
// 	received := `[
//   { "Product": "olivolja", "Amount": "till stekning" },
//   { "Product": "vitlök", "Amount": "3 klyftor" },
//   { "Product": "smör", "Amount": "till stekning" },
//   { "Product": "färska salsicciakorvar", "Amount": "4" },
//   { "Product": "vispgrädde", "Amount": "2 dl" },
//   { "Product": "passerade tomater", "Amount": "3 dl" },
//   { "Product": "tomatpuré", "Amount": "2 msk" },
//   { "Product": "vitt vin", "Amount": "1-2 dl" },
//   { "Product": "oregano", "Amount": "några kvistar" },
//   { "Product": "timjankvist", "Amount": "några kvistar" },
//   { "Product": "pasta (av favoritsort)", "Amount": "4 port" },
//   { "Product": "salt", "Amount": "efter smak" },
//   { "Product": "svartpeppar", "Amount": "nymalen" },
//   { "Product": "persilja", "Amount": "några kvistar" }
// ]`

	var gs []Grocery
	err := json.Unmarshal([]byte(received), &gs)

	if err != nil {
		panic(err)
	}

	for _, g := range gs {
		fmt.Println(g.Product + " " + g.Amount)
	}

	data := gin.H{
		"Groceries": gs,
		"Title":     "Extracted Groceries",
	}

	c.HTML(http.StatusOK, "groceries_extraction.html", data)
}

func HandleAcceptGroceries(c *gin.Context, conn *pgx.Conn) {
	products := c.PostFormArray("product")
	amounts := c.PostFormArray("amount")
	brands := c.PostFormArray("brand")
	stores := c.PostFormArray("store")

	hid, ok := c.Get("household_id")
	if !ok {
		panic("failed to retrieve household_id")
	}

	var err error
	// var groceries []Grocery

	sql := `INSERT INTO groceries 
	(product, brand, amount, store, picked, household_id)
	VALUES ($1, $2, $3, $4, $5, $6)`

	for i := range len(products) {
		// groceries[i] = Grocery{
		// 	Product:      products[i],
		// 	Amount:       amounts[i],
		// 	Brand:        brands[i],
		// 	Store:        stores[i],
		// 	Household_id: hid.(int),
		// }

		_, err = conn.Exec(context.Background(), sql, products[i], brands[i], amounts[i], stores[i], false, hid.(int))

		if err != nil {
			fmt.Println("noooooooooo bad input")
			panic(err)
		}
	}

	c.Redirect(302, "/groceries")
}

func GroceriesHandleFunc(c *gin.Context, conn *pgx.Conn) {
	hid, ok := c.Get("household_id")

	if !ok {
		panic("failed to get household id from context")
	}

	sql := `
        SELECT id, product, brand, store, amount, household_id, picked 
        FROM groceries
        WHERE household_id = $1
        ORDER BY product;
    `

	rows, err := conn.Query(context.Background(), sql, hid)
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	var groceries groceries

	for rows.Next() {
		var g Grocery
		err := rows.Scan(
			&g.Id,
			&g.Product,
			&g.Brand,
			&g.Store,
			&g.Amount,
			&g.Household_id,
			&g.Picked,
		)

		if err != nil {
			panic(err)
		}

		if g.Picked == false {
			groceries.NotPicked = append(groceries.NotPicked, g)
		} else {
			groceries.Picked = append(groceries.Picked, g)
		}
	}

	if rows.Err() != nil {
		fmt.Println(rows.Err())
	}

	data := gin.H{
		"Title":       "Groceries",
		"CurrentPath": c.Request.URL.Path,
		"Data":        groceries,
	}
	c.HTML(http.StatusOK, "groceries.html", data)
}

func GroceriesPickHandleFunc(c *gin.Context, conn *pgx.Conn) {
	id := c.PostForm("id")
	sql := `UPDATE groceries SET picked = NOT picked WHERE id=$1;`
	_, err := conn.Exec(context.Background(), sql, id)
	if err != nil {
		panic(err)
	}
	c.Redirect(302, "/groceries")
}

func GroceriesAddHandleFunc(c *gin.Context, conn *pgx.Conn) {
	hid, ok := c.Get("household_id")
	if !ok {
		panic("failed to get household_id")
	}
	product := c.PostForm("product")
	brand := c.PostForm("brand")
	amount := c.PostForm("amount")
	store := c.PostForm("store")
	grocery := Grocery{
		Product:      product,
		Brand:        brand,
		Amount:       amount,
		Store:        store,
		Picked:       false,
		Household_id: hid.(int),
	}

	sql := `INSERT INTO groceries 
	(product, brand, amount, store, picked, household_id)
	VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := conn.Exec(context.Background(), sql, grocery.Product, grocery.Brand, grocery.Amount, grocery.Store, grocery.Picked, grocery.Household_id)

	if err != nil {
		fmt.Println("noooooooooo bad input")
		panic(err)
	}

	c.Redirect(302, "/groceries")
}

func GroceriesEditHandleFunc(c *gin.Context, conn *pgx.Conn) {

	var groceries []Grocery
	err := c.BindJSON(&groceries)
	if err != nil {
		panic(err)
	}

	tx, err := conn.Begin(context.Background())
	if err != nil {
		panic(err)
	}
	defer tx.Rollback(context.Background())

	for _, g := range groceries {
		sql := `UPDATE groceries
		SET product=$1, amount=$2, brand=$3, store=$4
		WHERE id=$5`
		_, err = tx.Exec(context.Background(), sql, g.Product, g.Amount, g.Brand, g.Store, g.Id)

		if err != nil {
			panic(err)
		}
	}

	err = tx.Commit(context.Background())
	if err != nil {
		panic(err)
	}

	// c.Redirect(302, "/groceries")
	c.JSON(200, gin.H{"status": "ok"})
}

func GroceriesDeletePickedHandleFunc(c *gin.Context, conn *pgx.Conn) {
	hid, ok := c.Get("household_id")
	if !ok {
		panic("failed to get household_id from context")
	}
	sql := `DELETE FROM groceries WHERE household_id=$1 AND picked=true`

	_, err := conn.Exec(context.Background(), sql, hid)
	if err != nil {
		panic(err)
	}
	c.Redirect(302, "/groceries")
}
