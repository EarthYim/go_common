package logger

import (
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
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, &opts))
}
