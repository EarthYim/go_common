package logger

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
)

var logLevel = map[string]slog.Level{
	"debug": slog.LevelDebug,
	"info":  slog.LevelInfo,
	"error": slog.LevelError,
}

func Init(level string) *slog.Logger {
	opts := slog.HandlerOptions{
		Level: logLevel[level],
		// AddSource: true,
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &opts))
	slog.SetDefault(logger)
	return logger
}

type ctxKeyLogger struct{}

func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKeyLogger{}, logger)
}

func Logger(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(ctxKeyLogger{}).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}

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
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			// slog.String("body", string(body)),
			slog.String("request_id", r.Header.Get("X-Request-ID")),
		)

		// Attach logger to context
		ctx := WithLogger(r.Context(), reqLogger)
		c.Request = r.WithContext(ctx)

		// Proceed to next handler
		c.Next()
	}
}
