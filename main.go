package main

// maybe use method instead of globals
//
// func (a app*) f() {
// 	a.dostuff()
// }
// app.f()

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type LayoutData struct {
	Title       string
	CurrentPath string
	Data        any
}

type Grocery struct {
	Id           int    `json:"id"`
	Product      string `json:"product"`
	Amount       string `json:"amount"`
	Brand        string `json:"brand"`
	Store        string `json:"store"`
	Household_id int
	Picked       bool
}

type Session struct {
	user_id      int
	household_id *int
}

var r *gin.Engine
var conn *pgx.Conn
var sessions map[string]*Session

func main() {
	r = gin.Default()
	r.LoadHTMLGlob("web/templates/*.html")

	r.Static("/static/", "web/static")
	r.GET("/login", loginHandleFunc)
	r.POST("/login", authHandleFunc)
	r.GET("/logout", logoutHandlerFunc)

	auth := r.Group("/")
	auth.Use(authMiddleware())
	auth.GET("/chores", choresHandleFunc)
	auth.GET("/groceries", groceriesHandleFunc)
	auth.POST("/groceries", groceriesPickHandleFunc)
	auth.POST("/groceries/add", groceriesAddHandleFunc)
	auth.POST("/groceries/edit", groceriesEditHandleFunc)
	auth.GET("/", indexHandleFunc)
	auth.GET("/recipes", recipesHandleFunc)

	var err error
	conn, err = pgx.Connect(context.Background(), "postgres://Admin:Admin@localhost:5432/db?sslmode=disable")
	if err != nil {
		panic(err)
	}

	defer conn.Close(context.Background())

	sessions = make(map[string]*Session, 2)
	// setup()

	r.Run()
}

func setup() {
	id := addUserRetreiveId("Steffo", "apa")
	hid := createHousehold(id, "la casa")
	id = addUserRetreiveId("Anna", "gurk")
	joinHousehold(id, hid)

	session := &Session{
		household_id: &hid,
	}

	addGrocery(Grocery{Amount: "5", Product: "Mjölk", Brand: "Arla", Store: "ICA", Picked: false, Household_id: *session.household_id})
	addGrocery(Grocery{Amount: "1", Product: "Gurk", Brand: "Arla", Store: "ICA", Picked: false, Household_id: *session.household_id})
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

func verifyPassword(pwd, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pwd))

	if err != nil {
		fmt.Println("Wrong credentials!!!")
		return false
	}

	return true
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

func choresHandleFunc(c *gin.Context) {
	data := LayoutData{Title: "Chores", CurrentPath: c.Request.URL.Path, Data: nil}
	c.HTML(http.StatusOK, "chores.html", data)
}

func addGrocery(grocery Grocery) {
	sql := `INSERT INTO groceries 
	(product, brand, amount, store, picked, household_id)
	VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := conn.Exec(context.Background(), sql, grocery.Product, grocery.Brand, grocery.Amount, grocery.Store, grocery.Picked, grocery.Household_id)

	if err != nil {
		fmt.Println("noooooooooo bad input")
		panic(err)
	}
}

type Groceries struct {
	Picked, NotPicked []Grocery
}

func groceriesHandleFunc(c *gin.Context) {
	session, _ := c.Cookie("session_id")
	hid := *sessions[session].household_id
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

		if g.Picked == true {
			groceries.NotPicked = append(groceries.NotPicked, g)
		} else {
			groceries.Picked = append(groceries.Picked, g)
		}
	}

	if rows.Err() != nil {
		fmt.Println(rows.Err())
	}

	data := LayoutData{Title: "Groceries", CurrentPath: c.Request.URL.Path, Data: groceries}
	c.HTML(http.StatusOK, "groceries.html", data)
}

func groceriesPickHandleFunc(c *gin.Context) {
	id := c.PostForm("id")
	sql := `UPDATE groceries SET picked = NOT picked WHERE id=$1;`
	_, err := conn.Exec(context.Background(), sql, id)
	if err != nil {
		panic(err)
	}
	c.Redirect(302, "/groceries")
}

func groceriesAddHandleFunc(c *gin.Context) {
	product := c.PostForm("product")
	brand := c.PostForm("brand")
	amount := c.PostForm("amount")
	store := c.PostForm("store")
	cookie, _ := c.Cookie("session_id")
	session := sessions[cookie]
	grocery := Grocery{
		Product:      product,
		Brand:        brand,
		Amount:       amount,
		Store:        store,
		Picked:       false,
		Household_id: *session.household_id,
	}

	addGrocery(grocery)
	c.Redirect(302, "/groceries")
}

func groceriesEditHandleFunc(c *gin.Context) {

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

func indexHandleFunc(c *gin.Context) {
	data := LayoutData{Title: "Home", CurrentPath: c.Request.URL.Path, Data: nil}
	fmt.Println("-------------------- " + c.Request.URL.Path)
	c.HTML(http.StatusOK, "index.html", data)
}

func loginHandleFunc(c *gin.Context) {
	data := LayoutData{Title: "login", Data: nil}
	c.HTML(http.StatusOK, "login.html", data)
}

func getHouseholdId(user_id int) (int, error) {
	sql := `select household_id FROM household_members where user_id=$1`
	var hid int
	err := conn.QueryRow(context.Background(), sql, user_id).Scan(&hid)

	return hid, err
}

func authHandleFunc(c *gin.Context) {
	uname := c.PostForm("uname")
	pwd := c.PostForm("pwd")

	sql := "SELECT id, pwd FROM users WHERE username=$1"

	var uid int
	var hash string

	err := conn.QueryRow(context.Background(), sql, uname).Scan(&uid, &hash)

	if err != nil {
		if err == pgx.ErrNoRows {
			fmt.Println("User not found")
		} else {
			fmt.Println("Query error:", err)
		}
		c.Redirect(302, "/login")
		return
	}

	if verifyPassword(pwd, hash) {
		id := uuid.New().String()
		session := &Session{
			user_id:      uid,
			household_id: nil,
		}
		var hid int
		hid, err = getHouseholdId(uid)
		if err == nil {
			session.household_id = &hid
		}

		sessions[id] = session
		c.SetCookie("session_id", id, 0, "/", "", false, true)
		c.Redirect(302, "/")
	} else {
		c.Redirect(302, "/login")
	}
}

func logoutHandlerFunc(c *gin.Context) {
	cookie, err := c.Cookie("session_id")
	if err == nil {
		delete(sessions, cookie)
		c.SetCookie("session_id", "", -1, "/", "", false, true)
	}

	c.Redirect(302, "/login")
}

func recipesHandleFunc(c *gin.Context) {
	data := LayoutData{Title: "Recipes", CurrentPath: c.Request.URL.Path, Data: nil}
	c.HTML(http.StatusOK, "recipes.html", data)
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session, _ := c.Cookie("session_id")
		_, ok := sessions[session]

		if !ok {
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort()
			return
		}

		c.Next()
	}
}
