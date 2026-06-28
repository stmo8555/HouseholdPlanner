package login

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"golang.org/x/time/rate"
)

func (h *Handler) startCleanup() {
	c := cron.New()
	_, err := c.AddFunc("@every 5m", func() {
		cutoff := time.Now().Add(-15 * time.Minute)

		h.limiter.Range(func(key, value any) bool {
			if value.(*loginLimiter).lastSeen.Before(cutoff) {
				h.limiter.Delete(key)
			}

			return true
		})
	})
	if err != nil {
		panic(err)
	}

	c.Start()
}

type loginLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type Handler struct {
	service *Service
	limiter sync.Map
}

func (h *Handler) Register(c *gin.Context) {
	token := c.Param("token")
	uname := c.PostForm("uname")
	pwd := c.PostForm("pwd")

	userId, err := h.service.CreateUser(c.Request.Context(), uname, pwd, token)

	if err != nil {
		c.Redirect(302, "login.html")
		return
	}

	householdName := c.PostForm("household_name")
	inviteCode := c.PostForm("invite_code")

	if householdName != "" {
		err = h.service.CreateHousehold(c.Request.Context(), householdName, userId)
	} else if inviteCode != "" {
		err = h.service.JoinHousehold(c.Request.Context(), inviteCode, userId)
	} else {
		err = fmt.Errorf("No valid input for options")
	}

	if err != nil {
		panic(err)
	}

	c.Redirect(302, "/login?registered=1")
}
func (h *Handler) RegisterView(c *gin.Context) {
	token := c.Query("invite")
	valid, err := h.service.ValidateToken(c.Request.Context(), token)

	if err != nil {
		panic(err)
	}

	if !valid {
		c.Redirect(302, "login.html")
	}

	data := gin.H{
		"Title":   "Register",
		"HideNav": true,
		"Error":   "",
		"Token":   token,
	}

	c.HTML(200, "register.html", data)
}

func (h *Handler) WelcomeView(c *gin.Context) {
	if _, ok := c.Get("household_id"); ok {
		c.Redirect(302, "/")
		return
	}

	c.HTML(200, "welcome.html", gin.H{
		"Title":   "Set up your household",
		"HideNav": true,
		"Error":   "",
	})
}

func (h *Handler) SetupHousehold(c *gin.Context) {
	if _, ok := c.Get("household_id"); ok {
		c.Redirect(302, "/")
		return
	}

	userId := c.GetInt("user_id")
	householdName := c.PostForm("household_name")
	inviteCode := c.PostForm("invite_code")

	var err error
	switch {
	case householdName != "":
		err = h.service.CreateHousehold(c.Request.Context(), householdName, userId)
	case inviteCode != "":
		err = h.service.JoinHousehold(c.Request.Context(), inviteCode, userId)
	default:
		err = fmt.Errorf("no household option chosen")
	}

	if err != nil {
		c.HTML(http.StatusBadRequest, "welcome.html", gin.H{
			"Title":   "Set up your household",
			"HideNav": true,
			"Error":   "Could not set up your household. Check the invite code and try again.",
		})
		return
	}

	c.Redirect(302, "/")
}

func NewHandler(s *Service) *Handler {
	if s == nil {
		panic("nil service for handler")
	}

	handler := Handler{
		service: s,
		limiter: sync.Map{},
	}

	handler.startCleanup()

	return &handler
}

func (h *Handler) Login(c *gin.Context) {
	data := gin.H{
		"Title":      "Login",
		"HideNav":    true,
		"Registered": c.Query("registered") == "1",
	}
	c.HTML(200, "login.html", data)
}

func (h *Handler) Logout(c *gin.Context) {
	sessionID, err := c.Cookie("session_id")
	if err == nil {
		h.service.Logout(c.Request.Context(), sessionID)
	}

	c.SetCookie("session_id", "", -1, "/", "", gin.Mode() == gin.ReleaseMode, true)
	c.Redirect(302, "/login")
}

func (h *Handler) Authenticate(c *gin.Context) {
	ip := c.ClientIP()

	v, _ := h.limiter.LoadOrStore(ip, &loginLimiter{
		limiter:  rate.NewLimiter(rate.Every(10*time.Second/3), 3),
		lastSeen: time.Now(),
	})
	ll := v.(*loginLimiter)
	ll.lastSeen = time.Now()

	if !ll.limiter.Allow() {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error": "rate limit exceeded",
		})
		return
	}

	uname := c.PostForm("uname")
	pwd := c.PostForm("pwd")

	sessionID, err := h.service.Authenticate(c, uname, pwd)

	if err == nil {
		c.SetSameSite(http.SameSiteStrictMode)
		c.SetCookie("session_id", sessionID, 0, "/", "", gin.Mode() == gin.ReleaseMode, true)
		c.Redirect(302, "/")
		return
	}

	if errors.Is(err, errInvalidCredentials) {
		data := gin.H{
			"Title":   "Login",
			"Error":   "Invalid username or password.",
			"HideNav": true,
		}
		c.HTML(http.StatusUnauthorized, "login.html", data)
		return
	}

	c.String(http.StatusInternalServerError, "Login failed")
}
