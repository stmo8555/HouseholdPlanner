package todo

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Service *Service
}

func (h *Handler) Add(c *gin.Context) {
	hid := c.GetInt("household_id")

	task := c.PostForm("task")
	due := c.PostForm("due")
	repeat := c.PostForm("repeat")
	frequency := c.PostForm("frequency")

	task = strings.TrimSpace(c.PostForm("task"))
	if task == "" {
		c.Redirect(http.StatusSeeOther, "/todos")
		return
	}

	freqInt, err := strconv.Atoi(frequency)

	if err != nil {
		c.AbortWithStatus(500)
		c.String(500, err.Error())
		return
	}

	todo := Todo{Task: task, Repeat: repeat, Frequency: freqInt, HouseholdID: hid}
	if due != "" {
		todo.Due.Time, err = time.Parse("2006-01-02", due)
		todo.Due.Valid = true
	}

	if err != nil {
		c.AbortWithStatus(500)
		c.String(500, err.Error())
		return
	}

	err = h.Service.AddTodo(c, todo)
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
		"Overdue":     todoList.Overdue,
		"Today":       todoList.Today,
		"Soon":        todoList.Soon,
		"Completed":   todoList.Completed,
		"TheRest":     todoList.TheRest,
	}

	c.HTML(200, "todos.html", data)
}

func parseID(c *gin.Context) (int, error) {
	id := c.PostForm("id")
	return strconv.Atoi(id)
}
