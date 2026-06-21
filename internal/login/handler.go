package login

import (
	"errors"
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
	data := gin.H{"Title": "Login"}
	c.HTML(200, "login.html", data)
}

func (h *Handler) Logout(c *gin.Context) {
	sessionID, err := c.Cookie("session_id")
	if err == nil {
		h.service.Logout(c.Request.Context(), sessionID)
	}

	c.SetCookie("session_id", "", -1, "/", "", true, true)
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
		c.SetCookie("session_id", sessionID, 0, "/", "", true, true)
		c.Redirect(302, "/")
		return
	}

	if errors.Is(err, errInvalidCredentials) {
		data := gin.H{
			"Title": "Login",
			"Error": "Invalid username or password.",
		}
		c.HTML(http.StatusUnauthorized, "login.html", data)
		return
	}

	c.String(http.StatusInternalServerError, "Login failed")
}
