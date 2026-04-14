package todos

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Service *Service
}

func (h *Handler) Add(c *gin.Context) {
	hid := c.GetInt("household_id")

	todo := strings.TrimSpace(c.PostForm("todo"))

	if todo == "" {
		c.AbortWithStatus(400)
		c.String(500, "no todo")
		return
	}

	err := h.Service.AddTodo(c, todo, hid)
	if err != nil {
		c.AbortWithStatus(500)
		c.String(500, err.Error())
		return
	}

	c.Redirect(302, "/todos")
}

func (h *Handler) MarkDone(c *gin.Context) {
	hid := c.GetInt("household_id")

	id, err := parseID(c)

	if err != nil {
		c.AbortWithStatus(500)
		c.String(500, err.Error())
		return
	}

	err = h.Service.MarkDone(c, id, hid)

	if err != nil {
		c.AbortWithStatus(500)
		c.String(500, err.Error())
		return
	}

	c.Redirect(302, "/todos")
}

func (h *Handler) MarkUnDone(c *gin.Context) {
	hid := c.GetInt("household_id")

	id, err := parseID(c)

	if err != nil {
		c.AbortWithStatus(500)
		c.String(500, err.Error())
		return
	}

	err = h.Service.MarkUnDone(c, id, hid)

	if err != nil {
		c.AbortWithStatus(500)
		c.String(500, err.Error())
		return
	}

	c.Redirect(302, "/todos")
}

func (h *Handler) List(c *gin.Context) {
	hid := c.GetInt("household_id")

	todoList, err := h.Service.List(c, hid)

	if err != nil {
		c.AbortWithError(500, err)
		c.String(500, err.Error())
		return
	}

	data := gin.H{
		"Title":       "Todos",
		"CurrentPath": c.Request.URL.Path,
		"Active":       todoList.Active,
		"Completed":   todoList.Completed,
	}

	c.HTML(200, "todos.html", data)
}

func parseID(c *gin.Context) (int, error) {
	id := c.PostForm("id")
	return strconv.Atoi(id)
}
