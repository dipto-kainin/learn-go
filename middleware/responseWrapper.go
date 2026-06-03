package middleware

import (
	"restroBackend/helpers"
	"bytes"
	"encoding/json"
	"strings"

	"github.com/gin-gonic/gin"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	return w.body.Write(b)
}

func (w *bodyLogWriter) WriteString(s string) (int, error) {
	return w.body.WriteString(s)
}

// ResponseWrapper automatically wraps all JSON responses in the unified success/error envelope format
func ResponseWrapper() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a buffered response writer to intercept responses
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		c.Next()

		status := blw.Status()
		contentType := blw.Header().Get("Content-Type")

		// We only intercept JSON responses, excluding Swagger endpoints
		if strings.Contains(contentType, "application/json") && !strings.HasPrefix(c.Request.URL.Path, "/swagger") {
			var rawBody interface{}
			bodyBytes := blw.body.Bytes()

			if len(bodyBytes) > 0 {
				if err := json.Unmarshal(bodyBytes, &rawBody); err == nil {
					// Get requestID
					reqID, exists := c.Get("requestID")
					reqIDStr := "unknown"
					if exists {
						if str, ok := reqID.(string); ok {
							reqIDStr = str
						}
					} else {
						reqIDStr = helpers.GenerateRequestID()
					}

					var envelope map[string]interface{}
					isSuccess := status >= 200 && status < 300

					if !isSuccess {
						// Format errors: { success: false, requestID: string, data: { message: string } }
						message := "An error occurred"
						if m, ok := rawBody.(map[string]interface{}); ok {
							if errVal, exists := m["error"]; exists {
								if errStr, ok := errVal.(string); ok {
									message = errStr
								}
							} else if msgVal, exists := m["message"]; exists {
								if msgStr, ok := msgVal.(string); ok {
									message = msgStr
								}
							}
						} else if strBody, ok := rawBody.(string); ok {
							message = strBody
						}

						envelope = map[string]interface{}{
							"success":   false,
							"requestID": reqIDStr,
							"data": map[string]interface{}{
								"message": message,
							},
						}
					} else {
						// Format success: { success: true, requestID: string, data: any }
						// Detect if already wrapped to avoid double nesting
						isAlreadyWrapped := false
						if m, ok := rawBody.(map[string]interface{}); ok {
							_, hasSuccess := m["success"]
							_, hasData := m["data"]
							_, hasReqID := m["requestID"]
							if hasSuccess && hasData && hasReqID {
								isAlreadyWrapped = true
							}
						}

						if isAlreadyWrapped {
							envelope = rawBody.(map[string]interface{})
						} else {
							envelope = map[string]interface{}{
								"success":   true,
								"requestID": reqIDStr,
								"data":      rawBody,
							}
						}
					}

					// Marshal the wrapped response
					newBodyBytes, err := json.Marshal(envelope)
					if err == nil {
						blw.ResponseWriter.Header().Del("Content-Length")
						blw.ResponseWriter.Write(newBodyBytes)
						return
					}
				}
			}
		}

		// Fallback: write original response if not JSON or parsing fails
		blw.ResponseWriter.Write(blw.body.Bytes())
	}
}
