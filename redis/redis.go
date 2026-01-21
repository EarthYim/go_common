package redis

import (
	"common/config"
	"context"

	"github.com/redis/go-redis/v9"
)

func NewConnection(cfg config.Config) redis.UniversalClient {

	rdb := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs: []string{"redis:6379"},
	})
	defer rdb.Close()

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		panic(err)
	}

	return rdb
}
