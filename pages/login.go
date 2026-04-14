package pages

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type Session struct {
	user_id      int
	household_id *int
}

func LoginHandleFunc(c *gin.Context) {
	data := gin.H{"Title": "Login"}
	c.HTML(http.StatusOK, "login.html", data)
}

func LogoutHandlerFunc(c *gin.Context, sessions map[string]*Session) {
	cookie, err := c.Cookie("session_id")
	if err == nil {
		delete(sessions, cookie)
		c.SetCookie("session_id", "", -1, "/", "", false, true)
	}

	c.Redirect(302, "/login")
}

func AuthHandleFunc(c *gin.Context, conn *pgx.Conn, sessions map[string]*Session) {
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
		hid, err = getHouseholdId(uid, conn)
		if err == nil {
			session.household_id = &hid
		} else {
			panic("not implemented yet")
		}

		sessions[id] = session
		c.SetCookie("session_id", id, 0, "/", "", false, true)
		c.Redirect(302, "/")
	} else {
		c.Redirect(302, "/login")
	}
}

func verifyPassword(pwd, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pwd))

	if err != nil {
		fmt.Println("Wrong credentials!!!")
		return false
	}

	return true
}

func AuthMiddleware(sessions map[string]*Session) gin.HandlerFunc {
	return func(c *gin.Context) {
		uuid, _ := c.Cookie("session_id")
		session, ok := sessions[uuid]

		if !ok {
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort()
			return
		}

		c.Set("household_id", *session.household_id)
		fmt.Printf("Household id is set to %v\n", *session.household_id)

		c.Next()
	}
}

func getHouseholdId(user_id int, conn *pgx.Conn) (int, error) {
	sql := `select household_id FROM household_members where user_id=$1`
	var hid int
	err := conn.QueryRow(context.Background(), sql, user_id).Scan(&hid)

	return hid, err
}
