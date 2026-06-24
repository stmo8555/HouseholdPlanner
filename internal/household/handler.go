package household

import (
	// "errors"
	// "fmt"
	// "strconv"
	// "strings"

	"strconv"

	"github.com/gin-gonic/gin"
)

type handler struct {
	service *Service
}

func NewHandler(s *Service) *handler {
	if s == nil {
		panic("nil service for handler")
	}

	return &handler{service: s}
}

func (h *handler) RegenerateHouseholdCode(c *gin.Context) {
	panic("unimplemented")
}

func (h *handler) GenerateInviteToken(c *gin.Context) {
	hid := c.GetInt("household_id")
	uid := c.GetInt("user_id")
	inviteCode, err := h.service.GenerateInviteToken(c.Request.Context(),uid, hid)

	if err != nil {
		panic(err)
	}

	c.String(200, inviteCode)	
}

func (h *handler) RemoveMember(c *gin.Context) {
	hid := c.GetInt("household_id")

	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		panic(err)
	}

	err = h.service.RemoveMember(c.Request.Context(), id, hid)	

	if err != nil {
		panic(err)
	}
}

func (h *handler) PromoteMember(c *gin.Context) {
	hid := c.GetInt("household_id")
	uid := c.GetInt("user_id")


}

func (h *handler) Settings(c *gin.Context) {
	// hid := c.GetInt("household-id")

	// h.service.Settings(c.Request.Context(), hid)
	c.HTML(200, "settings.html", gin.H{
		"Title":       "Household settings",
		"CurrentPath": "/settings",
		"IsOwner":     true,
		"Household": gin.H{
			"Name": "la casa",
			"Code": "CASA42",
		},
		"Invite": gin.H{
			"Link": "https://householdplanner.app/register?invite=8f3a9c2e",
		},
		"Members": []gin.H{
			{"ID": 1, "Name": "steffe", "IsOwner": true, "IsYou": true},
			{"ID": 2, "Name": "anna", "IsOwner": false, "IsYou": false},
		},
	})
}
