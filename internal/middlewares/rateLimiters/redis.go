package ratelimiters

import (
	"fmt"
	"time"

	"github.com/redis/go-redis/v9" 
	"github.com/gofiber/fiber/v2"
)

func NewRedisRateLimiter(client *redis.Client, limit int, window time.Duration, prefix string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		ip := c.IP()
		key := fmt.Sprintf("%s%s", prefix, ip)

		currentCount, err := client.Incr(ctx, key).Result()
		if err != nil {
			fmt.Printf("Redis INCR failed for key %s: %v\n", key, err)
			return c.Next()
		}

		if currentCount == 1 {
			if err := client.Expire(ctx, key, window).Err(); err != nil {
				fmt.Printf("Redis EXPIRE failed for key %s: %v\n", key, err)
			}
		}

		if currentCount > int64(limit) {
			ttl, ttlErr := client.TTL(ctx, key).Result()
			retryAfter := window.Seconds()

			if ttlErr == nil && ttl > 0 {
				retryAfter = ttl.Seconds()
			}

			c.Set("Retry-After", fmt.Sprintf("%.0f", retryAfter))
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"message": fmt.Sprintf("Rate limit exceeded. Try again in %.0f seconds", retryAfter),
			})
		}

		return c.Next()
	}
}