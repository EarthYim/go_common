package main

import (
	"common/auth"
	"common/config"
	"common/logger"
	"common/middleware"
	"common/redis"
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

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
	r.Use(middleware.LoggerMiddleware(slog))
	r.GET("/health", func(c *gin.Context) {
		// log := logger.Logger(c.Request.Context())
		// log.Info("Healthy")
		c.Status(http.StatusOK)
	})

	authHandler := auth.NewAuthHandler(cfg)

	r.GET("/token", authHandler.AdminLoginHandler)

	r.Use(middleware.JwtMiddleware(cfg))
	r.GET("/auth/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"msg": "you're in",
		})
	})

	srv := http.Server{
		Addr:              ":8080",
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	idleConnsClosed := make(chan struct{})

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)
		<-sigint

		d := time.Duration(5 * time.Second)
		fmt.Printf("shutting down int %s ...", d)
		// We received an interrupt signal, shut down.
		ctx, cancel := context.WithTimeout(context.Background(), d)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			// Error from closing listeners, or context timeout:
			slog.Info("HTTP server Shutdown: " + err.Error())
		}
		close(idleConnsClosed)
	}()

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		slog.Error("HTTP server ListenAndServe: " + err.Error())
		return
	}

	<-idleConnsClosed
	fmt.Println("gracefully")
}
