package middleware

import (
	"bytes"
	"common/logger"
	"context"
	"io"
	"log/slog"

	"github.com/gin-gonic/gin"
)

func LoggerMiddleware(base *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		r := c.Request

		// Read and restore request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			base.Error("failed to read request body", slog.String("error", err.Error()))
		}
		r.Body = io.NopCloser(bytes.NewBuffer(body))

		// Create logger with request details
		reqLogger := base.With(
			slog.String("path", r.URL.Path),
			slog.String("request_id", r.Header.Get("X-Request-ID")),
		)

		reqLogger.Info("REQUEST", slog.String("method", r.Method), slog.String("body", string(body)))

		// Attach logger to context
		ctx := WithLogger(r.Context(), reqLogger)
		c.Request = r.WithContext(ctx)

		// Proceed to next handler
		c.Next()
	}
}

func WithLogger(ctx context.Context, base *slog.Logger) context.Context {
	return context.WithValue(ctx, logger.CtxKeyLogger{}, base)
}
