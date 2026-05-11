package login

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Handler struct {
	service *Service
}

func CreateHandler(s *Service) *Handler {
	if s == nil {
		panic("nil service for handler")
	}

	return &Handler{
		service: s,
	}
}

func (h *Handler) Login(c *gin.Context) {
	data := gin.H{"Title": "Login"}
	c.HTML(200, "login.html", data)
}

func (h *Handler) Logout(c *gin.Context) {
	sessionID, err := c.Cookie("session_id")
	if err == nil {
		h.service.Logout(c.Request.Context(), sessionID)
	}

	c.SetCookie("session_id", "", -1, "/", "", true, true)
	c.Redirect(302, "/login")
}

func (h *Handler) Authenticate(c *gin.Context) {
	uname := c.PostForm("uname")
	pwd := c.PostForm("pwd")

	sessionID, err := h.service.Authenticate(c, uname, pwd)


	if err == nil {
		c.SetSameSite(http.SameSiteStrictMode)
		c.SetCookie("session_id", sessionID, 0, "/", "", true, true)
		c.Redirect(302, "/recipes")
	} else {
		c.Redirect(302, "/login")
	}
}
