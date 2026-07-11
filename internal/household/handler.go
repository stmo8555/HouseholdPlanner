package household

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/csrf"
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

	version, err := h.service.HouseholdVersion(c.Request.Context(), hid)
	if err != nil {
		panic(err)
	}

	data := gin.H{
		"OOB":              true,
		"Code":             code,
		"HouseholdVersion": version,
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

	c.Redirect(303, "/settings")
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

	c.Redirect(303, "/settings")
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

	c.SetCookie("session_id", "", -1, "/", "", gin.Mode() == gin.ReleaseMode, true)
	c.Redirect(302, "/login")
}

func (h *handler) Settings(c *gin.Context) {
	hid := c.GetInt("household_id")
	uid := c.GetInt("user_id")

	settingsView, err := h.service.Settings(c.Request.Context(), uid, hid)
	if err != nil {
		panic(err)
	}
	decorateInvites(c, settingsView.Invites)

	version, err := h.service.HouseholdVersion(c.Request.Context(), hid)
	if err != nil {
		panic(err)
	}

	c.HTML(200, "settings.html", gin.H{
		"Title":            "Household settings",
		"CurrentPath":      "/settings",
		"SettingsView":     settingsView,
		"HouseholdVersion": version,
		"CSRFToken":        csrf.Token(c.Request),
	})
}

func (h *handler) GenerateInviteToken(c *gin.Context) {
	hid := c.GetInt("household_id")
	uid := c.GetInt("user_id")

	if _, err := h.service.GenerateInviteToken(c.Request.Context(), uid, hid); err != nil {
		panic(err)
	}

	h.renderInvites(c, uid, hid, "New invite link generated")
}

func (h *handler) RevokeInvite(c *gin.Context) {
	hid := c.GetInt("household_id")
	uid := c.GetInt("user_id")

	if err := h.service.RevokeInvite(c.Request.Context(), c.Param("token"), hid); err != nil {
		panic(err)
	}

	h.renderInvites(c, uid, hid, "Invite link revoked")
}

// renderInvites re-renders the invite list partial after a mutation, with an
// OOB version bump so the household-changed notification fires and a toast.
func (h *handler) renderInvites(c *gin.Context, uid, hid int, toast string) {
	settingsView, err := h.service.Settings(c.Request.Context(), uid, hid)
	if err != nil {
		panic(err)
	}
	decorateInvites(c, settingsView.Invites)

	version, err := h.service.HouseholdVersion(c.Request.Context(), hid)
	if err != nil {
		panic(err)
	}

	c.HTML(200, "household/invites-list", gin.H{
		"SettingsView":     settingsView,
		"OOB":              true,
		"Toast":            toast,
		"HouseholdVersion": version,
	})
}

// decorateInvites fills in the display-only fields (share link, a masked token,
// and human-friendly created/expiry labels) that aren't stored in the database.
func decorateInvites(c *gin.Context, invites []Invite) {
	for i := range invites {
		invites[i].Link = createLink(c, invites[i].Token)
		invites[i].Masked = maskedToken(invites[i].Token)
		invites[i].Created = createdLabel(invites[i].CreatedAt)
		invites[i].Expires = expiresLabel(invites[i].ExpiresAt)
		invites[i].ExpiringSoon = time.Until(invites[i].ExpiresAt) < 24*time.Hour
	}
}

func maskedToken(token string) string {
	t := strings.ReplaceAll(token, "-", "")
	if len(t) <= 6 {
		return t
	}
	return "…" + t[len(t)-6:]
}

func createdLabel(createdAt time.Time) string {
	d := time.Since(createdAt)
	switch {
	case d < 24*time.Hour:
		return "created today"
	case d < 48*time.Hour:
		return "created yesterday"
	default:
		return fmt.Sprintf("created %d days ago", int(d.Hours())/24)
	}
}

func expiresLabel(expiresAt time.Time) string {
	d := time.Until(expiresAt)
	switch {
	case d <= 0:
		return "Expired"
	case d >= 48*time.Hour:
		return fmt.Sprintf("Expires in %d days", int(d.Hours())/24)
	case d >= 24*time.Hour:
		return "Expires in 1 day"
	case d >= 2*time.Hour:
		return fmt.Sprintf("Expires in %d hours", int(d.Hours()))
	case d >= time.Hour:
		return "Expires in 1 hour"
	default:
		return "Expires soon"
	}
}

func createLink(c *gin.Context, token string) string {
	scheme := "https"
	if c.Request.TLS == nil {
		scheme = "http"
	}
	return fmt.Sprintf("%s://%s/register?invite=%s", scheme, c.Request.Host, token)
}
