package middleware

import (
	"common/config"
	"common/logger"

	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type RateLimiter interface {
	CheckLimit(ctx context.Context, key string, window time.Duration, limit int) error
}

func AppRateLimiterMiddleware(limiter RateLimiter, cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {

		log := logger.Logger(c.Request.Context())

		appContext, err := GetAppContext(c)
		if err != nil {
			log.Error("failed to read app context", slog.String("func", "AppRateLimiterMiddleware"))
			c.AbortWithStatus(http.StatusInternalServerError)
		}

		window, limit := GetTierProp(appContext, cfg)

		if err = limiter.CheckLimit(c.Request.Context(), ResolveLimitKey("app", appContext.UID), window, limit); err != nil {
			if err.Error() == "too many request" {
				c.AbortWithStatus(http.StatusTooManyRequests)
			}

			log.Error("rate limiter error", slog.String("err", err.Error()))
			c.AbortWithStatus(http.StatusInternalServerError)
		}

		c.Next()
	}
}

func AuthRateLimiterMiddleware(limiter RateLimiter, cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {

		log := logger.Logger(c.Request.Context())

		appContext, err := GetAppContext(c)
		if err != nil {
			log.Error("failed to read app context", slog.String("func", "AppRateLimiterMiddleware"))
			c.AbortWithStatus(http.StatusInternalServerError)
		}

		// window, limit := GetTierProp(appContext, cfg)

		if err = limiter.CheckLimit(c.Request.Context(), ResolveLimitKey("auth", appContext.UID), time.Duration(cfg.RateLimiter.LimitWindow)*time.Minute, cfg.RateLimiter.AuthLimit); err != nil {
			if err.Error() == "too many request" {
				c.AbortWithStatus(http.StatusTooManyRequests)
			}

			log.Error("rate limiter error", slog.String("err", err.Error()))
			c.AbortWithStatus(http.StatusInternalServerError)
		}

		c.Next()
	}
}

func ResolveLimitKey(scope string, uid string) string {
	return fmt.Sprintf("limit:%s:%s", scope, uid)
}

func GetTierProp(appContext AppContext, cfg config.Config) (window time.Duration, limit int) {
	if appContext.Tier == NormalTier {
		window = time.Duration(cfg.RateLimiter.LimitWindow) * time.Minute
		limit = cfg.RateLimiter.ClientLimitNormal
		return
	}

	window = time.Duration(cfg.RateLimiter.LimitWindow) * time.Minute
	limit = cfg.RateLimiter.ClientLimitThrottled
	return
}

type rateLimiter struct {
	rdb redis.UniversalClient
}

func NewRateLimiter(rdb redis.UniversalClient) *rateLimiter {
	return &rateLimiter{
		rdb: rdb,
	}
}

func (r *rateLimiter) CheckLimit(ctx context.Context, key string, window time.Duration, limit int) error {

	now := time.Now().UnixMilli()
	windowStart := now - window.Milliseconds()

	pipe := r.rdb.TxPipeline()
	pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart, 10))
	countCmd := pipe.ZCard(ctx, key)
	pipe.Expire(ctx, key, window)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}

	count := int(countCmd.Val())
	if count >= limit {
		return errors.New("too many request")
	}

	return nil
}
