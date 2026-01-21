package logger

import (
	"context"
	"log/slog"
	"os"
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

type CtxKeyLogger struct{}

func Logger(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(CtxKeyLogger{}).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}
