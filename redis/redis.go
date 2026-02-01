package redis

import (
	"common/config"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewConnection(cfg config.Config) redis.UniversalClient {
	opts := &redis.UniversalOptions{
		Addrs:    []string{cfg.Redis.Addr1},
		Password: cfg.Redis.Password,
	}

	if cfg.Redis.TlsEanbled {
		caCert, err := os.ReadFile(cfg.Redis.CaCertPath)
		if err != nil {
			panic(fmt.Errorf("read redis CA cert: %w", err))
		}

		caPool := x509.NewCertPool()
		if !caPool.AppendCertsFromPEM(caCert) {
			panic("failed to append redis CA cert")
		}

		opts.TLSConfig = &tls.Config{
			RootCAs:    caPool,
			MinVersion: tls.VersionTLS12,
		}
	}

	rdb := redis.NewUniversalClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		panic(fmt.Errorf("redis ping failed: %w", err))
	}

	return rdb
}
