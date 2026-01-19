package main

import (
	"common/config"
	"common/logger"
	"common/redis"
	"os"
)

func main() {

	// read config
	cfg := config.C(os.Getenv("ENV"))

	// logger (and trace maybe)
	logger := logger.Init(cfg.LogLevel)
	logger.Info("Hello")

	// db and redis connection
	_ = redis.NewConnection(cfg)

	// init Gin router

}
