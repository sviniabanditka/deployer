package middleware

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

// RateLimitConfig holds configuration for the rate limiter.
type RateLimitConfig struct {
	RedisClient    *redis.Client
	MaxRequests    int
	WindowDuration time.Duration
	KeyPrefix      string
}

// RateLimiter returns a Fiber middleware that implements a token bucket
// rate limiter backed by Redis.
//
// For public endpoints: 100 requests per minute per IP (default).
// For authenticated endpoints: 1000 requests per minute per user.
func RateLimiter(cfg RateLimitConfig) fiber.Handler {
	if cfg.MaxRequests == 0 {
		cfg.MaxRequests = 100
	}
	if cfg.WindowDuration == 0 {
		cfg.WindowDuration = time.Minute
	}
	if cfg.KeyPrefix == "" {
		cfg.KeyPrefix = "ratelimit"
	}

	return func(c *fiber.Ctx) error {
		var key string

		// Use user ID if authenticated, otherwise use IP
		if userID, ok := c.Locals("userID").(string); ok && userID != "" {
			key = fmt.Sprintf("%s:user:%s", cfg.KeyPrefix, userID)
		} else {
			key = fmt.Sprintf("%s:ip:%s", cfg.KeyPrefix, c.IP())
		}

		ctx := context.Background()

		// Use Redis sliding window counter
		now := time.Now().UnixNano()
		windowStart := now - cfg.WindowDuration.Nanoseconds()

		pipe := cfg.RedisClient.Pipeline()

		// Remove expired entries
		pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart, 10))

		// Count current requests in window
		countCmd := pipe.ZCard(ctx, key)

		// Add current request
		pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now})

		// Set expiry on key
		pipe.Expire(ctx, key, cfg.WindowDuration+time.Second)

		_, err := pipe.Exec(ctx)
		if err != nil {
			// If Redis is unavailable, allow the request through
			return c.Next()
		}

		count := countCmd.Val()

		// Set rate limit headers
		remaining := int64(cfg.MaxRequests) - count
		if remaining < 0 {
			remaining = 0
		}

		c.Set("X-RateLimit-Limit", strconv.Itoa(cfg.MaxRequests))
		c.Set("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))
		c.Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(cfg.WindowDuration).Unix(), 10))

		if count > int64(cfg.MaxRequests) {
			retryAfter := cfg.WindowDuration.Seconds()
			c.Set("Retry-After", strconv.FormatInt(int64(retryAfter), 10))
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "too many requests",
			})
		}

		return c.Next()
	}
}

// PublicRateLimiter returns a rate limiter configured for public endpoints:
// 100 requests per minute per IP.
func PublicRateLimiter(redisClient *redis.Client) fiber.Handler {
	return RateLimiter(RateLimitConfig{
		RedisClient:    redisClient,
		MaxRequests:    100,
		WindowDuration: time.Minute,
		KeyPrefix:      "ratelimit:public",
	})
}

// AuthenticatedRateLimiter returns a rate limiter configured for authenticated
// endpoints: 1000 requests per minute per user.
func AuthenticatedRateLimiter(redisClient *redis.Client) fiber.Handler {
	return RateLimiter(RateLimitConfig{
		RedisClient:    redisClient,
		MaxRequests:    1000,
		WindowDuration: time.Minute,
		KeyPrefix:      "ratelimit:auth",
	})
}
