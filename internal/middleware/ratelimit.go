package middleware

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/go-gin-template/pkg/cache"
	"github.com/yourusername/go-gin-template/pkg/config"
	"github.com/yourusername/go-gin-template/pkg/response"
	"golang.org/x/time/rate"
)

// RateLimiter interface for different rate limiting implementations
type RateLimiter interface {
	Allow(key string) (bool, error)
	GetLimit() int
	GetDuration() time.Duration
}

// InMemoryRateLimiter implements rate limiting using in-memory storage
type InMemoryRateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	limit    int
	duration time.Duration
}

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// NewInMemoryRateLimiter creates a new in-memory rate limiter
func NewInMemoryRateLimiter(requests int, duration time.Duration) *InMemoryRateLimiter {
	rl := &InMemoryRateLimiter{
		visitors: make(map[string]*visitor),
		limit:    requests,
		duration: duration,
	}

	// Start cleanup goroutine
	go rl.cleanupVisitors()

	return rl
}

// Allow checks if a request is allowed
func (rl *InMemoryRateLimiter) Allow(key string) (bool, error) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[key]
	if !exists {
		// Create new rate limiter: requests per duration
		limiter := rate.NewLimiter(rate.Every(rl.duration/time.Duration(rl.limit)), rl.limit)
		rl.visitors[key] = &visitor{limiter: limiter, lastSeen: time.Now()}
		return limiter.Allow(), nil
	}

	v.lastSeen = time.Now()
	return v.limiter.Allow(), nil
}

// GetLimit returns the rate limit
func (rl *InMemoryRateLimiter) GetLimit() int {
	return rl.limit
}

// GetDuration returns the rate limit duration
func (rl *InMemoryRateLimiter) GetDuration() time.Duration {
	return rl.duration
}

// cleanupVisitors removes old visitors
func (rl *InMemoryRateLimiter) cleanupVisitors() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		for key, v := range rl.visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(rl.visitors, key)
			}
		}
		rl.mu.Unlock()
	}
}

// RedisRateLimiter implements rate limiting using Redis
type RedisRateLimiter struct {
	cache    *cache.Cache
	limit    int
	duration time.Duration
}

// NewRedisRateLimiter creates a new Redis-based rate limiter
func NewRedisRateLimiter(cache *cache.Cache, requests int, duration time.Duration) *RedisRateLimiter {
	return &RedisRateLimiter{
		cache:    cache,
		limit:    requests,
		duration: duration,
	}
}

// Allow checks if a request is allowed using Redis
func (rl *RedisRateLimiter) Allow(key string) (bool, error) {
	ctx := context.Background()
	redisKey := "rate_limit:" + key

	// Increment counter
	count, err := rl.cache.Increment(ctx, redisKey)
	if err != nil {
		return true, err // Allow on error
	}

	// Set expiration on first request
	if count == 1 {
		if err := rl.cache.Expire(ctx, redisKey, rl.duration); err != nil {
			return true, err
		}
	}

	return count <= int64(rl.limit), nil
}

// GetLimit returns the rate limit
func (rl *RedisRateLimiter) GetLimit() int {
	return rl.limit
}

// GetDuration returns the rate limit duration
func (rl *RedisRateLimiter) GetDuration() time.Duration {
	return rl.duration
}

// RateLimit creates a rate limiting middleware
func RateLimit(limiter RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use client IP as the rate limit key
		key := c.ClientIP()

		allowed, err := limiter.Allow(key)
		if err != nil {
			// Log error but allow request
			c.Next()
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(limiter.GetLimit()))
		c.Header("X-RateLimit-Duration", limiter.GetDuration().String())

		if !allowed {
			c.Header("Retry-After", strconv.Itoa(int(limiter.GetDuration().Seconds())))
			response.TooManyRequests(c, "Rate limit exceeded")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitByUser creates a rate limiting middleware that uses user ID if authenticated
func RateLimitByUser(limiter RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		var key string

		// Use user ID if authenticated, otherwise use IP
		if userID, exists := c.Get(string(UserIDKey)); exists {
			key = "user:" + strconv.FormatUint(uint64(userID.(uint)), 10)
		} else {
			key = "ip:" + c.ClientIP()
		}

		allowed, err := limiter.Allow(key)
		if err != nil {
			c.Next()
			return
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(limiter.GetLimit()))
		c.Header("X-RateLimit-Duration", limiter.GetDuration().String())

		if !allowed {
			c.Header("Retry-After", strconv.Itoa(int(limiter.GetDuration().Seconds())))
			response.TooManyRequests(c, "Rate limit exceeded")
			c.Abort()
			return
		}

		c.Next()
	}
}

// NewRateLimitFromConfig creates a rate limiter from configuration
func NewRateLimitFromConfig(cfg config.RateLimitConfig, redisCache *cache.Cache) RateLimiter {
	if redisCache != nil {
		return NewRedisRateLimiter(redisCache, cfg.Requests, cfg.Duration)
	}
	return NewInMemoryRateLimiter(cfg.Requests, cfg.Duration)
}

// RateLimitByEndpoint creates rate limiters for specific endpoints
func RateLimitByEndpoint(limits map[string]RateLimiter, defaultLimiter RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.FullPath()
		key := c.ClientIP()

		limiter := defaultLimiter
		if specificLimiter, exists := limits[path]; exists {
			limiter = specificLimiter
		}

		allowed, err := limiter.Allow(key + ":" + path)
		if err != nil {
			c.Next()
			return
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(limiter.GetLimit()))
		c.Header("X-RateLimit-Duration", limiter.GetDuration().String())

		if !allowed {
			c.Header("Retry-After", strconv.Itoa(int(limiter.GetDuration().Seconds())))
			response.TooManyRequests(c, "Rate limit exceeded for this endpoint")
			c.Abort()
			return
		}

		c.Next()
	}
}
