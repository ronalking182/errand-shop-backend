package middleware

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

type RateLimiter struct {
	clients map[string]*ClientInfo
	mu      sync.RWMutex
	limit   int
	window  time.Duration
}

type ClientInfo struct {
	requests  int
	lastReset time.Time
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		clients: make(map[string]*ClientInfo),
		limit:   limit,
		window:  window,
	}

	// Cleanup goroutine
	go rl.cleanup()

	return rl
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.window)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, client := range rl.clients {
			if now.Sub(client.lastReset) > rl.window {
				delete(rl.clients, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ip := c.IP()
		now := time.Now()

		rl.mu.Lock()
		client, exists := rl.clients[ip]
		if !exists {
			client = &ClientInfo{
				requests:  0,
				lastReset: now,
			}
			rl.clients[ip] = client
		}

		// Reset if window has passed
		if now.Sub(client.lastReset) > rl.window {
			client.requests = 0
			client.lastReset = now
		}

		client.requests++
		requests := client.requests
		rl.mu.Unlock()

		if requests > rl.limit {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":  "Rate limit exceeded",
				"limit":  rl.limit,
				"window": rl.window.String(),
			})
		}

		return c.Next()
	}
}

// Predefined rate limiters
func AuthRateLimit() fiber.Handler {
	return NewRateLimiter(5, time.Minute).Middleware() // 5 requests per minute for auth
}

func APIRateLimit() fiber.Handler {
	return NewRateLimiter(100, time.Minute).Middleware() // 100 requests per minute for API
}

func StrictRateLimit() fiber.Handler {
	return NewRateLimiter(3, time.Minute).Middleware() // 3 requests per minute for sensitive operations
}
