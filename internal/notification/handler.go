package notification

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	if service == nil {
		panic("nil service for handler")
	}

	return &Handler{service: service}
}

func (h *Handler) Check(c *gin.Context) {
	householdID := c.GetInt("household_id")
	currentVersion, err := h.service.HouseholdVersion(c.Request.Context(), householdID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	knownVersion, err := strconv.Atoi(c.Query("known_version"))
	if err != nil {
		knownVersion = currentVersion
	}

	c.HTML(http.StatusOK, "notifications/poller", gin.H{
		"KnownVersion": knownVersion,
		"Changed":      currentVersion > knownVersion,
	})
}

func (h *Handler) Ack(c *gin.Context) {
	householdID := c.GetInt("household_id")
	currentVersion, err := h.service.HouseholdVersion(c.Request.Context(), householdID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.HTML(http.StatusOK, "notifications/poller", gin.H{
		"KnownVersion": currentVersion,
		"Changed":      false,
	})
}
