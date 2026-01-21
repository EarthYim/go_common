package main

import (
	"common/config"
	"common/logger"
	"common/redis"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {

	// read config
	cfg := config.C(os.Getenv("ENV"))

	// logger (and trace maybe)
	slog := logger.Init(cfg.LogLevel)
	slog.Info("Starting Server...")

	// db and redis connection
	_ = redis.NewConnection(cfg)

	// init Gin router
	r := gin.Default()
	r.Use(logger.LoggerMiddleware(slog))
	r.GET("/health", func(c *gin.Context) {
		// log := logger.Logger(c.Request.Context())
		// log.Info("Healthy")
		c.Status(http.StatusOK)
	})

	r.Run()
}
