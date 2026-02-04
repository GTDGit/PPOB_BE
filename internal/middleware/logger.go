package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger returns a middleware that logs HTTP requests
func Logger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log after request
		latency := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		requestID := c.GetString("request_id")

		if query != "" {
			path = path + "?" + query
		}

		attrs := []any{
			slog.String("request_id", requestID),
			slog.String("method", method),
			slog.String("path", path),
			slog.Int("status", status),
			slog.Duration("latency", latency),
			slog.String("client_ip", clientIP),
		}

		// Add error if exists
		if len(c.Errors) > 0 {
			attrs = append(attrs, slog.String("error", c.Errors.String()))
		}

		// Log based on status
		if status >= 500 {
			logger.Error("HTTP request", attrs...)
		} else if status >= 400 {
			logger.Warn("HTTP request", attrs...)
		} else {
			logger.Info("HTTP request", attrs...)
		}
	}
}
