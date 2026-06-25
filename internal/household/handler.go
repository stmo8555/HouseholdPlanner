package household

import (
	"errors"
	"fmt"
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
	hid := c.GetInt("household_id")
	uid := c.GetInt("user_id")

	code, err := h.service.RegenerateHouseholdCode(c.Request.Context(), uid, hid)

	if err != nil {
		panic(err)
	}

	data := gin.H{
		"Swapped": true,
		"Code": code,
	}

	c.HTML(200, "household/household-code", data)
}

func (h *handler) RemoveMember(c *gin.Context) {
	hid := c.GetInt("household_id")
	uid := c.GetInt("user_id")

	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		panic(err)
	}

	err = h.service.RemoveMember(c.Request.Context(), uid, id, hid)

	if err != nil {
		panic(err)
	}

	h.renderMembersList(c, uid, hid, "Member removed")
}

func (h *handler) PromoteMember(c *gin.Context) {
	hid := c.GetInt("household_id")
	uid := c.GetInt("user_id")

	targetUID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		panic(err)
	}

	err = h.service.PromoteMember(c.Request.Context(), uid, targetUID, hid)
	if err != nil {
		panic(err)
	}

	h.renderMembersList(c, uid, hid, "Ownership transferred")
}

func (h *handler) renderMembersList(c *gin.Context, uid, hid int, message string) {
	settingsView, err := h.service.Settings(c.Request.Context(), uid, hid)
	if err != nil {
		panic(err)
	}

	c.HTML(200, "household/members-list", gin.H{
		"SettingsView": settingsView,
		"Swapped":      true,
		"Message":      message,
	})
}

func (h *handler) LeaveHousehold(c *gin.Context) {
	hid := c.GetInt("household_id")
	uid := c.GetInt("user_id")

	err := h.service.LeaveHousehold(c.Request.Context(), uid, hid)
	if err != nil {
		if errors.Is(err, ErrOwnerMustTransfer) {
			c.Redirect(302, "/settings")
			return
		}
		panic(err)
	}

	c.Redirect(302, "/welcome")
}

func (h *handler) DeleteHousehold(c *gin.Context) {
	hid := c.GetInt("household_id")
	uid := c.GetInt("user_id")

	err := h.service.DeleteHousehold(c.Request.Context(), uid, hid)
	if err != nil {
		if errors.Is(err, ErrHouseholdNotEmpty) || errors.Is(err, ErrOwnerMustTransfer) {
			c.Redirect(302, "/settings")
			return
		}
		panic(err)
	}

	c.Redirect(302, "/welcome")
}

func (h *handler) DeleteAccount(c *gin.Context) {
	uid := c.GetInt("user_id")

	err := h.service.DeleteAccount(c.Request.Context(), uid)
	if err != nil {
		if errors.Is(err, ErrOwnerMustTransfer) {
			c.Redirect(302, "/settings")
			return
		}
		panic(err)
	}

	c.SetCookie("session_id", "", -1, "/", "", true, true)
	c.Redirect(302, "/login")
}

func (h *handler) Settings(c *gin.Context) {
	hid := c.GetInt("household_id")
	uid := c.GetInt("user_id")

	token, err := h.service.CurrentInviteToken(c.Request.Context(), uid, hid)
	if err != nil {
		panic(err)
	}

	h.renderSettings(c, uid, hid, token)
}

func (h *handler) GenerateInviteToken(c *gin.Context) {
	hid := c.GetInt("household_id")
	uid := c.GetInt("user_id")

	token, err := h.service.GenerateInviteToken(c.Request.Context(), uid, hid)
	if err != nil {
		panic(err)
	}

	data := gin.H{
		"InviteLink": createLink(c, token),
		"Swapped":    true,
	}

	c.HTML(200, "household/invite-link", data)
}

func (h *handler) renderSettings(c *gin.Context, uid, hid int, token string) {
	settingsView, err := h.service.Settings(c.Request.Context(), uid, hid)
	if err != nil {
		panic(err)
	}

	inviteLink := createLink(c, token)

	c.HTML(200, "settings.html", gin.H{
		"Title":        "Household settings",
		"CurrentPath":  "/settings",
		"SettingsView": settingsView,
		"InviteLink":   inviteLink,
	})
}

func createLink(c *gin.Context, token string) string {
	scheme := "https"
	if c.Request.TLS == nil {
		scheme = "http"
	}
	return fmt.Sprintf("%s://%s/register?invite=%s", scheme, c.Request.Host, token)
}


