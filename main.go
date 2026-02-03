package main

import (
	"common/auth"
	"common/config"
	"common/logger"
	"common/middleware"
	"common/redis"

	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {

	// read config
	cfg := config.C(os.Getenv("ENV"))

	// logger (and trace maybe)
	slog := logger.Init(cfg.LogLevel)
	slog.Info("Starting Server...")

	// db and redis connection
	rdb := redis.NewConnection(cfg)
	defer rdb.Close()

	// init Gin router
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.String(200, "ok")
	})
	r.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	corsOpts := cors.Config{
		AllowOrigins: []string{
			"http://localhost:3000",
			"http://localhost:3001",
		},
		AllowMethods: []string{
			"GET", "POST", "OPTIONS",
		},
		AllowHeaders: []string{
			"Authorization",
			"Content-Type",
		},
		ExposeHeaders: []string{
			"Authorization",
		},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	// Add CORS middleware
	r.Use(cors.New(corsOpts))

	r.Use(middleware.LoggerMiddleware(slog))

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

	// gracefull shutdown
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)
		<-sigint

		d := time.Duration(5 * time.Second)
		fmt.Printf("shutting down int %s ...", d)
		ctx, cancel := context.WithTimeout(context.Background(), d)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
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
