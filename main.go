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
	Title string
	Data  any
}

type Grocery struct {
	Product, Brand, Unit, Store string
	Amount, id                  int
	Picked                      bool
}

var r *gin.Engine
var conn *pgx.Conn
var sessions map[string]string

func main() {
	r = gin.Default()
	r.LoadHTMLGlob("web/templates/*.html")

	r.Static("/static/", "web/static")
	r.GET("/login", loginHandleFunc)
	r.POST("/login", authHandleFunc)
	r.GET("/logout", logoutHandlerFunc)

	auth := r.Group("/")
	auth.Use(AuthMiddleware())
	auth.GET("/chores", choresHandleFunc)
	auth.GET("/groceries", groceriesHandleFunc)
	auth.GET("/", indexHandleFunc)
	auth.GET("/recipes", recipesHandleFunc)

	var err error
	conn, err = pgx.Connect(context.Background(), "postgres://planner_user:supersecretpassword@localhost:5432/household_db?sslmode=disable")
	if err != nil {
		panic(err)
	}

	defer conn.Close(context.Background())

	sessions = make(map[string]string, 2)

	id := addUserRetreiveId("Anna", "gurk")
	joinHousehold(id, 1)

	r.Run()
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

func createHousehold(owner int, name string) {
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
}

func addUser(conn *pgx.Conn) {
	sql := `insert into users (username,pwd) Values ($1,$2) Returning id;`

	password := "apa"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	_, err := conn.Exec(context.Background(), sql, "Steffo", hashedPassword)

	if err != nil {
		panic("failed to create user: " + err.Error())
	}
}

func verifyPassword(conn *pgx.Conn, uname, password string) bool {
	sql := "SELECT pwd FROM users WHERE username=$1"
	var hash string
	err := conn.QueryRow(context.Background(), sql, uname).Scan(&hash)

	if err != nil {
		fmt.Println("Could not verify user. Err: " + err.Error())
		return false
	}

	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))

	if err != nil {
		fmt.Println("Wrong credentials!!!")
		return false
	}

	return true
}

func deleteUser(conn *pgx.Conn) {
	sql := `DELETE FROM users WHERE username=$1;`
	_, err := conn.Exec(context.Background(), sql, "Steffo")

	if err != nil {
		panic("Failed to to delete user: " + err.Error())
	}
}

func addUserRetreiveId(uname, pwd string) int {
	sql := `INSERT INTO users (username,pwd) VALUES ($1,$2) RETURNING id;`
	var id int
	err := conn.QueryRow(context.Background(), sql, uname, pwd).Scan(&id)

	if err != nil {
		panic("failed to create user: " + err.Error())
	}

	return id
}

func choresHandleFunc(c *gin.Context) {
	data := LayoutData{Title: "Chores", Data: nil}
	c.HTML(http.StatusOK, "chores.html", data)
}

func addGrocery(grocery Grocery, user_id int) {
	sql := `select household_id from users where user_id=$1`
	var household_id string
	err := conn.QueryRow(context.Background(), sql, user_id).Scan(&household_id)

	if err != nil {
		panic(err)
	}

	sql = `INSERT INTO groceries 
	(product, brand, amount, unit, store, household_id)
	VALUES ($1, $2, $3, $4, $5)`

	_, err = conn.Exec(context.Background(), sql, grocery.Product, grocery.Brand, grocery.Amount, grocery.Unit, grocery.Store)

	if err != nil {
		panic(err)
	}
}

func groceriesHandleFunc(c *gin.Context) {
	// var name string var weight int64 err = conn.QueryRow(context.Background(), "select name, weight from widgets where id=$1", 42).Scan(&name, &weight)
	// sql := `SELECT * from groceries;`
	// conn.Query(context.Background())
	groceries := make([]Grocery, 0, 5)
	groceries = append(groceries, Grocery{Amount: 5, Product: "Mjölk", Brand: "Arla", Unit: "kg", Store: "ICA", Picked: false})
	groceries = append(groceries, Grocery{Amount: 1, Product: "Gurk", Brand: "Arla", Unit: "kg", Store: "ICA", Picked: true})
	data := LayoutData{Title: "Groceries", Data: groceries}
	c.HTML(http.StatusOK, "groceries.html", data)
}

func indexHandleFunc(c *gin.Context) {
	data := LayoutData{Title: "Home", Data: nil}
	c.HTML(http.StatusOK, "index.html", data)
}

func loginHandleFunc(c *gin.Context) {
	data := LayoutData{Title: "login", Data: nil}
	c.HTML(http.StatusOK, "login.html", data)
}

func authHandleFunc(c *gin.Context) {
	uname := c.PostForm("uname")
	pwd := c.PostForm("pwd")

	if verifyPassword(conn, uname, pwd) {
		id := uuid.New().String()
		sessions[id] = uname
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
	data := LayoutData{Title: "Recipes", Data: nil}
	c.HTML(http.StatusOK, "recipes.html", data)
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session, _ := c.Cookie("session_id")
		_, ok := sessions[session]

		if !ok {
			c.AbortWithStatusJSON(401, gin.H{
				"error": "unauthorized",
			})
			return
		}

		c.Next()
	}
}
