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

		now := time.Now().UTC()

		if now.After(session.ExpiresAt) || now.Sub(session.CreatedAt) > MaxSessionLifetime {
			s.repo.RemoveSession(c.Request.Context(), sessionID)
			redirectToLogin(c)
			return
		}

		if time.Until(session.ExpiresAt) < SessionTTL/2 {
			s.ExtendSession(c.Request.Context(), sessionID, SessionTTL)
		}

		c.Set("user_id", session.User.ID)
		if session.HouseholdID != nil {
			c.Set("household_id", *session.HouseholdID)
		}

		c.Next()
	}
}

// RequireHousehold guards routes that act on a household. The household is
// resolved per request from current membership (see getSession), so a user who
// has been removed from their household no longer has a household_id and is sent
// to onboarding instead of being allowed to operate on it.
func RequireHousehold() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, ok := c.Get("household_id"); !ok {
			if c.GetHeader("HX-Request") == "true" {
				c.Header("HX-Redirect", "/welcome")
				c.AbortWithStatus(http.StatusForbidden)
				return
			}

			c.Redirect(http.StatusFound, "/welcome")
			c.Abort()
			return
		}

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



