package ratelimiters

import (
	"log"
	"fmt"
	"time"
	
	"lqkhoi-go-http-api/internal/cache"

	"github.com/gofiber/fiber/v2"
)

func NewRedisRateLimiter(cacheRepo cache.CacheRepository, limit int, window time.Duration, prefix string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		ip := c.IP()
		key := fmt.Sprintf("%s%s", prefix, ip)

		currentCount, err := cacheRepo.Increment(ctx, key)
		if err != nil {
			log.Printf("Rate Limiter: Cache Increment failed for key %s: %v\n", key, err)
			return c.Next()
		}

		if currentCount == 1 {
			if err := cacheRepo.Expire(ctx, key, window); err != nil {
				log.Printf("Rate Limiter: Cache Expire failed for key %s: %v\n", key, err)
			}
		}

		if currentCount > int64(limit) {
			ttl, ttlErr := cacheRepo.GetTTL(ctx, key)
			retryAfter := window.Seconds()

			if ttlErr == nil && ttl > 0 {
				retryAfter = ttl.Seconds()
			} else if ttlErr != nil {
                log.Printf("Rate Limiter: Cache GetTTL failed for key %s: %v\n", key, ttlErr)
            }

			c.Set(fiber.HeaderRetryAfter, fmt.Sprintf("%.0f", retryAfter))
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"message": fmt.Sprintf("Rate limit exceeded. Try again in %.0f seconds", retryAfter),
			})
		}

		return c.Next()
	}
}