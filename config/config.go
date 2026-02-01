package config

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	LogLevel string `env:"LOG_LEVEL"`

	Base64JwtPrivateKey string `env:"SECRET_JWT_PRIVATE_KEY"`
	Base64JwtPublicKey  string `env:"SECRET_JWT_PUBLIC_KEY"`

	RateLimiter RateLimiter
	Redis       RedisCfg
}

type RateLimiter struct {
	ClientLimitNormal    int `env:"RATE_LIMIT_CLIENT_NORMAL"`
	ClientLimitThrottled int `env:"RATE_LIMIT_CLIENT_THROTTLED"`

	AuthLimit   int `env:"RATE_LIMIT_AUTH"`
	GlobalLimit int `env:"RATE_LIMIT_GLOBAL"`

	LimitWindow int `env:"RATE_LIMIT_WINDOW"` // seconds
}

type RedisCfg struct {
	Addr1      string `env:"REDIS_ADDR_1"`
	Password   string `env:"REDIS_PASSWORD"`
	TlsEanbled bool   `env:"REDIS_TLS"`
	CaCertPath string `env:"REDIS_CA_CERT_PATH"`
}

var once sync.Once
var config Config

func prefix(e string) string {
	if e == "" {
		return ""
	}

	return fmt.Sprintf("%s_", e)
}

func parseEnv[T any](opts env.Options) (T, error) {
	var t T

	if err := env.Parse(&t); err != nil {
		return t, err
	}

	// override with PREFIX_XXX_XXX if when it has valu
	// this is optional no need handle error because if it not found it will use default value

	//nolint:all
	env.ParseWithOptions(&t, opts)

	return t, nil
}

func C(envPrefix string) Config {
	once.Do(func() {
		opts := env.Options{
			Prefix: prefix(envPrefix),
			// support both "30" "500ms", "2s", "1.5m", "3h"
			FuncMap: map[reflect.Type]env.ParserFunc{
				reflect.TypeOf(time.Duration(0)): func(v string) (any, error) {
					if _, err := strconv.Atoi(v); err == nil {
						v += "s"
					}
					return time.ParseDuration(v)
				},
			},
		}

		var err error
		config, err = parseEnv[Config](opts)
		if err != nil {
			log.Fatal(err)
		}
	})

	return config
}
