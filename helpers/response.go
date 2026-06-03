package helpers

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"
)

// GenerateRequestID generates a random hex string for request tracing
func GenerateRequestID() string {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		return "unknown"
	}
	return hex.EncodeToString(b)
}

// SendSuccess sends a structured JSON success response
func SendSuccess(c *gin.Context, statusCode int, data interface{}) {
	reqID, exists := c.Get("requestID")
	reqIDStr := "unknown"
	if exists {
		if str, ok := reqID.(string); ok {
			reqIDStr = str
		}
	} else {
		reqIDStr = GenerateRequestID()
	}

	c.JSON(statusCode, gin.H{
		"success":   true,
		"requestID": reqIDStr,
		"data":      data,
	})
}

// SendError sends a structured JSON error response
func SendError(c *gin.Context, statusCode int, message string) {
	reqID, exists := c.Get("requestID")
	reqIDStr := "unknown"
	if exists {
		if str, ok := reqID.(string); ok {
			reqIDStr = str
		}
	} else {
		reqIDStr = GenerateRequestID()
	}

	c.JSON(statusCode, gin.H{
		"success":   false,
		"requestID": reqIDStr,
		"data": gin.H{
			"message": message,
		},
	})
}
