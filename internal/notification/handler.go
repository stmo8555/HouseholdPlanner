package notification

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/stmo8555/HouseholdPlanner/internal/household"
)

type Handler struct {
	householdService *household.Service
}

func NewHandler(householdService *household.Service) *Handler {
	if householdService == nil {
		panic("nil service for handler")
	}

	return &Handler{householdService: householdService}
}

func (h *Handler) Check(c *gin.Context) {
	householdID := c.GetInt("household_id")
	currentVersion, err := h.householdService.HouseholdVersion(c.Request.Context(), householdID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	knownVersion, err := strconv.Atoi(c.Query("known_version"))
	if err != nil {
		knownVersion = currentVersion
	}

	c.HTML(http.StatusOK, "notifications/poller", gin.H{
		"OOB":              false,
		"HouseholdVersion": knownVersion,
		"Changed":          currentVersion > knownVersion,
	})
}

func (h *Handler) Ack(c *gin.Context) {
	householdID := c.GetInt("household_id")
	currentVersion, err := h.householdService.HouseholdVersion(c.Request.Context(), householdID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.HTML(http.StatusOK, "notifications/poller", gin.H{
		"OOB":              false,
		"HouseholdVersion": currentVersion,
		"Changed":          false,
	})
}
