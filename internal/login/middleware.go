package login

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(s *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := c.Cookie("session_id")
		if err != nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		session, err := s.GetSession(sessionID)
		if err != nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		c.Set("user_id", session.UserID)
		c.Set("household_id", *session.HouseholdID)

		c.Next()
	}
}
