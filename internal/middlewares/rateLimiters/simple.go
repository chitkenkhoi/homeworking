package ratelimiters

import(
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

type visitorInfo struct {
	count      int
	lastSeen   time.Time
	windowStart time.Time
}

var visitors = make(map[string]*visitorInfo)
var mu sync.Mutex

func NewSimpleRateLimiter(limit int, window time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ip := c.IP() 

		mu.Lock() 

		v, exists := visitors[ip]
		now := time.Now()

		if !exists || now.After(v.windowStart.Add(window)) {
			visitors[ip] = &visitorInfo{
				count:      1,
				lastSeen:   now,
				windowStart: now,
			}
			mu.Unlock() 
			return c.Next()
		}

		v.count++
		v.lastSeen = now
		
		if v.count > limit {
			resetTime := v.windowStart.Add(window)
			retryAfter := resetTime.Sub(now)

			c.Set("Retry-After", fmt.Sprintf("%.0f", retryAfter.Seconds()))

			mu.Unlock()
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"message": fmt.Sprintf("Rate limit exceeded. Try again in %.0f seconds", retryAfter.Seconds()),
			})
		}

		mu.Unlock()
		return c.Next()
	}
}