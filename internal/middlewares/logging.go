package middlewares

import (
	"log/slog"
	"time"

	"github.com/BagRoman01/image-sketch-processor/internal/logging"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		c.Set("request_id", requestID)

		ctx := logging.WithRequestID(c.Request.Context(), requestID)
		c.Request = c.Request.WithContext(ctx)

		c.Header("X-Request-ID", requestID)

		c.Next()

		duration := time.Since(start)

		slog.Info("HTTP request",
			"request_id", requestID,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration_ms", duration.Milliseconds(),
			"ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		)

		if len(c.Errors) > 0 {
			slog.Error("Request errors",
				"request_id", requestID,
				"errors", c.Errors.String(),
			)
		}
	}
}
