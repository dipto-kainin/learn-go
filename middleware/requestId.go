package middleware

import (
	"restroBackend/helpers"

	"github.com/gin-gonic/gin"
)

// RequestID middleware generates and injects a request ID into the context and response headers
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := helpers.GenerateRequestID()
		c.Set("requestID", reqID)
		c.Writer.Header().Set("X-Request-ID", reqID)
		c.Next()
	}
}
