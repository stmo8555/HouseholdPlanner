package login

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// lastSeenThrottle bounds how often a request refreshes users.last_seen, so an
// active session (or polling tab) doesn't write to the row on every request.
const lastSeenThrottle = 10 * time.Minute

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

		if time.Since(session.User.LastSeen) > lastSeenThrottle {
			s.TouchLastSeen(c.Request.Context(), session.User.ID)
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
