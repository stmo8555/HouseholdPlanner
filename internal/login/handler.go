package login

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Handler struct {
	Service *Service
}

func (h *Handler) Login(c *gin.Context) {
	data := gin.H{"Title": "Login"}
	c.HTML(200, "login.html", data)
}

func (h *Handler) Logout(c *gin.Context) {
	cookie, err := c.Cookie("session_id")
	if err == nil {
		h.Service.Logout(cookie)
	}

	c.SetCookie("session_id", "", -1, "/", "", true, true)
	c.Redirect(302, "/login")
}

func (h *Handler) Authenticate(c *gin.Context) {
	uname := c.PostForm("uname")
	pwd := c.PostForm("pwd")

	uuid := h.Service.Authenticate(c, uname, pwd)

	if uuid != "" {
		c.SetSameSite(http.SameSiteStrictMode)
		c.SetCookie("session_id", uuid, 0, "/", "", true, true)
		c.Redirect(302, "/")
	} else {
		c.Redirect(302, "/login")
	}
}
