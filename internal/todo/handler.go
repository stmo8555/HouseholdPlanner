package todo

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Handler struct {
	Service *Service
}

func NewHandler(s *Service) *Handler {
	if s == nil {
		panic("nil service for handler")
	}

	return &Handler{
		Service: s,
	}
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
		panic(err)
	}

	todo := Todo{Task: task, Repeat: Repeat(repeat), Frequency: freqInt, HouseholdID: hid}
	if due != "" {
		todo.Due.Time, err = time.Parse("2006-01-02", due)
		todo.Due.Valid = true
	}

	if err != nil {
		panic(err)
	}

	err = h.Service.AddTodo(c, todo)
	if err != nil {
		panic(err)
	}

	h.ListPartial(c)
}

func (h *Handler) SmartAdd(c *gin.Context) {
	hid := c.GetInt("household_id")
	text := c.PostForm("text")

	if strings.TrimSpace(text) == "" {
		c.AbortWithStatus(500)
		c.String(500, "Empty text")
		return
	}

	err := h.Service.SmartAdd(c, text, hid)

	if err != nil {
		panic(err)
	}

	h.ListPartial(c)
}

func (h *Handler) MarkDone(c *gin.Context) {
	hid := c.GetInt("household_id")

	id, err := parseID(c)

	if err != nil {
		panic(err)
	}

	err = h.Service.MarkDone(c, id, hid)

	if err != nil {
		c.AbortWithStatus(500)
		c.String(500, err.Error())
		return
	}

	h.ListPartial(c)
}

func (h *Handler) MarkUnDone(c *gin.Context) {
	hid := c.GetInt("household_id")

	id, err := parseID(c)

	if err != nil {
		panic(err)
	}

	err = h.Service.MarkUnDone(c, id, hid)

	h.ListPartial(c)
}

func (h *Handler) Delete(c *gin.Context) {
	hid := c.GetInt("household_id")

	id, err := parseID(c)

	if err != nil {
		panic(err)
	}

	err = h.Service.Delete(c, id, hid)

	if err != nil {
		c.AbortWithStatus(500)
		c.String(500, err.Error())
		return
	}

	h.ListPartial(c)
}

func (h *Handler) List(c *gin.Context) {
	data, err := h.todoListData(c)

	if err != nil {
		panic(err)
	}

	data["Title"] = "Todos"
	data["CurrentPath"] = c.Request.URL.Path

	c.HTML(200, "todos.html", data)
}

func (h *Handler) ListPartial(c *gin.Context) {
	data, err := h.todoListData(c)

	if err != nil {
		c.AbortWithError(500, err)
		c.String(500, err.Error())
		return
	}

	c.HTML(200, "todos/list", data)
}

func parseID(c *gin.Context) (int, error) {
	id := c.PostForm("id")
	return strconv.Atoi(id)
}

func (h *Handler) todoListData(c *gin.Context) (gin.H, error) {
	hid := c.GetInt("household_id")

	todoList, err := h.Service.List(c, hid)

	if err != nil {
		panic(err)
	}

	return gin.H{
		"Overdue":   todoList.Overdue,
		"Today":     todoList.Today,
		"Soon":      todoList.Soon,
		"Completed": todoList.Completed,
		"TheRest":   todoList.TheRest,
	}, nil
}
