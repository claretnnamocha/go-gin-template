package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/go-gin-template/pkg/logger"
	"go.uber.org/zap"
)

// responseWriter wraps gin.ResponseWriter to capture response body
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// Logger creates a request logging middleware
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get status code
		status := c.Writer.Status()

		// Build log fields
		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		}

		if query != "" {
			fields = append(fields, zap.String("query", query))
		}

		if requestID := c.GetString("request_id"); requestID != "" {
			fields = append(fields, zap.String("request_id", requestID))
		}

		if userID, exists := c.Get(string(UserIDKey)); exists {
			fields = append(fields, zap.Any("user_id", userID))
		}

		// Add error information if present
		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("errors", c.Errors.String()))
		}

		// Log based on status code
		switch {
		case status >= 500:
			logger.Error("Server error", fields...)
		case status >= 400:
			logger.Warn("Client error", fields...)
		case status >= 300:
			logger.Info("Redirect", fields...)
		default:
			logger.Info("Request", fields...)
		}
	}
}

// DetailedLogger creates a detailed request/response logging middleware
// Use with caution in production as it logs request/response bodies
func DetailedLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Read and restore request body
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Wrap response writer to capture response body
		blw := &responseWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Build log fields
		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", latency),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.Int("request_size", len(requestBody)),
			zap.Int("response_size", blw.body.Len()),
		}

		if query != "" {
			fields = append(fields, zap.String("query", query))
		}

		if requestID := c.GetString("request_id"); requestID != "" {
			fields = append(fields, zap.String("request_id", requestID))
		}

		// Only log body for specific content types and non-empty bodies
		contentType := c.GetHeader("Content-Type")
		if len(requestBody) > 0 && len(requestBody) < 10000 {
			if isTextContentType(contentType) {
				fields = append(fields, zap.String("request_body", string(requestBody)))
			}
		}

		// Log response body for debugging (be careful with sensitive data)
		responseBody := blw.body.String()
		if len(responseBody) > 0 && len(responseBody) < 10000 {
			fields = append(fields, zap.String("response_body", responseBody))
		}

		logger.Debug("Detailed request log", fields...)
	}
}

// isTextContentType checks if the content type is text-based
func isTextContentType(contentType string) bool {
	textTypes := []string{
		"application/json",
		"application/xml",
		"text/plain",
		"text/html",
		"text/xml",
		"application/x-www-form-urlencoded",
	}

	for _, t := range textTypes {
		if contentType == t || len(contentType) > len(t) && contentType[:len(t)] == t {
			return true
		}
	}
	return false
}
