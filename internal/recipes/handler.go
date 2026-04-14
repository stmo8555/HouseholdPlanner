package recipes

import (
	"github.com/gin-gonic/gin"
	"strings"
)

type Handler struct {
	Service *Service
}


func (h *Handler) List(c *gin.Context) {
	hid := c.GetInt("household_id")
	
	recipes, err := h.Service.List(c, hid)	

	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	data := gin.H{
		"Title":       "Groceries",
		"CurrentPath": c.Request.URL.Path,
		"Data":        recipes,
	}

	c.HTML(200, "recipes.html", data)
}

func (h *Handler) Add(c *gin.Context) {
	link := strings.TrimSpace(c.PostForm("link"))
	
	if link == "" {
		c.AbortWithStatus(500)
		return
	}

	hid := c.GetInt("household_id")

	err := h.Service.Add(c, hid, link)

	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	c.Redirect(302, "/recipes")
}
