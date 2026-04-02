package pages

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/net/html"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
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
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	tags := findTags(resp)

	fmt.Println("Parsed: " + tags)

	ctx := context.Background()
	client := openai.NewClient()

	templatePrompt := `Extract ingredients from the text.

Return ONLY raw JSON (no markdown, no code blocks):
[
	{ "Product": "ingredient name", "Amount": "amount", "Brand": "brand"}
]

Rules:
- Amount = only numbers + units (e.g. "2 dl", "1 msk", "efter smak")
- Product = only the ingredient name
- Brand = only the products brand
- If there is not suffiecent info, return empty json list []
- Remove preparation words (hackad, skivad, riven, etc.)
- Remove text after commas
- No explanations
- Do NOT wrap output in ANYTHING ELSE THAN []

Text: 
%s`

	question := fmt.Sprintf(templatePrompt, tags)
	ai_resp, ai_err := client.Responses.New(ctx, responses.ResponseNewParams{
		Input: responses.ResponseNewParamsInputUnion{OfString: openai.String(question)},
		Model: openai.ChatModelGPT4_1Nano,
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

	fmt.Println(string(received))
	var gs []Grocery
	err := json.Unmarshal([]byte(received), &gs)

	if err != nil {
		panic(err)
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

func findTags(resp *http.Response) string {
	doc, err := html.Parse(resp.Body)

	if err != nil {
		panic(err)
	}

	potentialCandidates := make([]*html.Node, 0)

	var processNode func(n *html.Node)
	processNode = func(n *html.Node) {
		if n.Type == html.TextNode {
			var match bool
			match, err = regexp.MatchString("^[iI]ngredien(ser|ts)$", n.Data)
			if err != nil {
				panic(err)
			}
			if match {
				parent := n.Parent
				for ; parent.Data == "div"; parent = parent.Parent {
				}

				parent = parent.Parent
				for ; parent != nil; parent = parent.NextSibling {
					potentialCandidates = append(potentialCandidates, parent)
				}
				return

			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			processNode(c)
		}
	}
	processNode(doc)

	fmt.Printf("numb of potential: %v\n", len(potentialCandidates))
	if len(potentialCandidates) == 0 {
		return ""
	}

	score_li := 1
	score_tr := 1
	score_attr := 1
	score_ingredient := 1
	score_regularText := 2
	score_units := 5
	score_numbers := 1

	var bestFit func(n *html.Node, index int)

	points := make([]int, len(potentialCandidates))
	bestFit = func(n *html.Node, index int) {
		var match bool
		if n.Type == html.TextNode {
			// fmt.Println("_____Text Node_____")
			match, err = regexp.MatchString("^[\\w\\såäöÅÄÖ\\.,]+$", n.Data)
			if err != nil {
				panic(err)
			}
			if match {
				points[index] += score_regularText
				// fmt.Print("regular ")
			}
			match, err = regexp.MatchString("[iI]ngredien(ser|ts)", n.Data)
			if err != nil {
				panic(err)
			}
			if match {
				points[index] += score_ingredient
				// fmt.Print("text ")
			}
			match, err = regexp.MatchString("^([mcd]?l|tm[sk]|k?g)$", strings.TrimSpace(n.Data))
			if err != nil {
				panic(err)
			}
			if match {
				points[index] += score_units
				// fmt.Print("units ")
			}
			match, err = regexp.MatchString("^\\d+$", n.Data)
			if err != nil {
				panic(err)
			}
			if match {
				points[index] += score_numbers
				// fmt.Print("number ")
			}

		}
		if n.Type == html.ElementNode && false {
			// fmt.Println("_____Element Node_____")
			for _, attr := range n.Attr {
				for val := range strings.SplitSeq(attr.Val, " ") {

					match, err = regexp.MatchString("[iI]ngredien(ser|ts)", val)
					if err != nil {
						panic(err)
					}
					if match {
						points[index] += score_attr
						// fmt.Print("attr ")
					}
				}
			}
			if n.Data == "li" {
				points[index] += score_li
				// fmt.Print("li ")
			}
			if n.Data == "tr" {
				points[index] += score_tr
				// fmt.Print("tr ")
			}

		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			bestFit(c, index)
		}
	}

	for i, nodes := range potentialCandidates {
		// fmt.Printf("_____BEGIN____%v\n", i)
		bestFit(nodes, i)
		// fmt.Println("______END______")
	}

	var text strings.Builder

	var findText func(n *html.Node)
	findText = func(n *html.Node) {
		if n.Type == html.TextNode {
			text.WriteString(strings.TrimSpace(n.Data) + " ")
		} else {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				findText(c)
			}
		}
	}

	maxIndex := 0
	maxValue := points[0]

	for i, v := range points {
		if v > maxValue {
			maxValue = v
			maxIndex = i
		}
		fmt.Printf("Points: %v %v\n", i, v)
	}

	fmt.Printf("Choosing: %v\n", maxIndex)
	findText(potentialCandidates[maxIndex])

	// for i, v := range potentialCandidates {
	//
	// 	findText(v)	
	// 	fmt.Printf("candidate %d------------------------------\n", i)
	// 	fmt.Println(text.String())
	// 	text.Reset()	
	// }


	// var sb strings.Builder
	//
	// html.Render(&sb, potentialCandidates[maxIndex])
	//
	// fmt.Println(sb.String())

	return text.String()
}
