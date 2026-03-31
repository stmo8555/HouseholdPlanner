package pages

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"net/http"
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

func GroceriesExtractFromRecipe(){
	received := `[
  { "Product": "olivolja", "Amount": "till stekning" },
  { "Product": "vitlök", "Amount": "3 klyftor" },
  { "Product": "smör", "Amount": "till stekning" },
  { "Product": "färska salsicciakorvar", "Amount": "4" },
  { "Product": "vispgrädde", "Amount": "2 dl" },
  { "Product": "passerade tomater", "Amount": "3 dl" },
  { "Product": "tomatpuré", "Amount": "2 msk" },
  { "Product": "vitt vin", "Amount": "1-2 dl" },
  { "Product": "oregano", "Amount": "några kvistar" },
  { "Product": "timjankvist", "Amount": "några kvistar" },
  { "Product": "pasta (av favoritsort)", "Amount": "4 port" },
  { "Product": "salt", "Amount": "efter smak" },
  { "Product": "svartpeppar", "Amount": "nymalen" },
  { "Product": "persilja", "Amount": "några kvistar" }
]`



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

	var groceries Groceries

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
