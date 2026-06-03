package middleware

import (
	"restroBackend/helpers"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func Authentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientToken := c.Request.Header.Get("token")
		if clientToken == "" {
			authHeader := c.Request.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				clientToken = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if clientToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No Authorization header provided"})
			c.Abort()
			return
		}

		claims, errMsg := helpers.ValidateToken(clientToken)
		if errMsg != "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": errMsg})
			c.Abort()
			return
		}

		c.Set("email", claims.Email)
		c.Set("first_name", claims.FirstName)
		c.Set("last_name", claims.LastName)
		c.Set("user_type", claims.UserType)

		c.Next()
	}
}

func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userType := c.GetString("user_type")
		if userType != "ADMIN" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}
