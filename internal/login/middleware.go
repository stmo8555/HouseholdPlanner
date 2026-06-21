package login

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(s *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := c.Cookie("session_id")
		if err != nil {
			redirectToLogin(c)
			return
		}

		session, err := s.GetSession(c.Request.Context(), sessionID)
		if err != nil {
			redirectToLogin(c)
			return
		}

		if time.Now().UTC().After(session.ExpiresAt) {
			s.repo.RemoveSession(c.Request.Context(), sessionID)

			redirectToLogin(c)
			return
		}

		c.Set("user_id", session.UserID)
		c.Set("household_id", *session.HouseholdID)

		c.Next()
	}
}

func redirectToLogin(c *gin.Context) {
	if c.GetHeader("HX-Request") == "true" {
		c.Header("HX-Redirect", "/login")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	c.Redirect(http.StatusFound, "/login")
	c.Abort()
}
