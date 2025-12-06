package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go-gin-template/pkg/cache"
)

// CacheConfig holds cache middleware configuration
type CacheConfig struct {
	TTL            time.Duration
	KeyPrefix      string
	IgnoreHeaders  []string
	CacheableStatus []int
}

// DefaultCacheConfig returns default cache configuration
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		TTL:             5 * time.Minute,
		KeyPrefix:       "cache:",
		IgnoreHeaders:   []string{"Authorization"},
		CacheableStatus: []int{http.StatusOK},
	}
}

// CacheResponse represents a cached response
type CacheResponse struct {
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Body    []byte            `json:"body"`
}

// Cache middleware caches responses for GET requests
func Cache(c *cache.Cache, cfg CacheConfig) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Only cache GET requests
		if ctx.Request.Method != http.MethodGet {
			ctx.Next()
			return
		}

		// Generate cache key
		key := generateCacheKey(ctx, cfg.KeyPrefix)

		// Try to get from cache
		var cached CacheResponse
		err := c.Get(ctx.Request.Context(), key, &cached)
		if err == nil {
			// Cache hit - return cached response
			for k, v := range cached.Headers {
				ctx.Header(k, v)
			}
			ctx.Header("X-Cache", "HIT")
			ctx.Data(cached.Status, cached.Headers["Content-Type"], cached.Body)
			ctx.Abort()
			return
		}

		// Cache miss - capture response
		writer := &cacheResponseWriter{
			ResponseWriter: ctx.Writer,
			body:          make([]byte, 0),
		}
		ctx.Writer = writer

		ctx.Next()

		// Only cache successful responses
		if !isCacheableStatus(writer.status, cfg.CacheableStatus) {
			return
		}

		// Store in cache
		headers := make(map[string]string)
		for k, v := range writer.Header() {
			if len(v) > 0 {
				headers[k] = v[0]
			}
		}

		cached = CacheResponse{
			Status:  writer.status,
			Headers: headers,
			Body:    writer.body,
		}

		_ = c.Set(ctx.Request.Context(), key, cached, cfg.TTL)
		ctx.Header("X-Cache", "MISS")
	}
}

// cacheResponseWriter captures response for caching
type cacheResponseWriter struct {
	gin.ResponseWriter
	body   []byte
	status int
}

func (w *cacheResponseWriter) Write(b []byte) (int, error) {
	w.body = append(w.body, b...)
	return w.ResponseWriter.Write(b)
}

func (w *cacheResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

// generateCacheKey generates a unique cache key based on request
func generateCacheKey(ctx *gin.Context, prefix string) string {
	// Create key from method, path, and query string
	raw := ctx.Request.Method + ":" + ctx.Request.URL.Path + ":" + ctx.Request.URL.RawQuery

	// Hash the key
	hash := sha256.Sum256([]byte(raw))
	return prefix + hex.EncodeToString(hash[:])
}

// isCacheableStatus checks if status code should be cached
func isCacheableStatus(status int, cacheableStatus []int) bool {
	for _, s := range cacheableStatus {
		if status == s {
			return true
		}
	}
	return false
}

// CacheInvalidate middleware invalidates cache for specific routes
func CacheInvalidate(c *cache.Cache, patterns ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()

		// Only invalidate on successful write operations
		if ctx.Writer.Status() >= 200 && ctx.Writer.Status() < 300 {
			for _, pattern := range patterns {
				keys, _ := c.Keys(ctx.Request.Context(), pattern)
				if len(keys) > 0 {
					_ = c.Delete(ctx.Request.Context(), keys...)
				}
			}
		}
	}
}

// CacheByUser creates a cache middleware that includes user ID in the key
func CacheByUser(c *cache.Cache, cfg CacheConfig) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.Request.Method != http.MethodGet {
			ctx.Next()
			return
		}

		// Add user ID to cache key if authenticated
		userKey := "anonymous"
		if userID, exists := ctx.Get(string(UserIDKey)); exists {
			userKey = "user:" + string(rune(userID.(uint)))
		}

		key := cfg.KeyPrefix + userKey + ":" + ctx.Request.URL.Path + ":" + ctx.Request.URL.RawQuery

		var cached CacheResponse
		err := c.Get(ctx.Request.Context(), key, &cached)
		if err == nil {
			for k, v := range cached.Headers {
				ctx.Header(k, v)
			}
			ctx.Header("X-Cache", "HIT")
			ctx.Data(cached.Status, cached.Headers["Content-Type"], cached.Body)
			ctx.Abort()
			return
		}

		writer := &cacheResponseWriter{
			ResponseWriter: ctx.Writer,
			body:          make([]byte, 0),
		}
		ctx.Writer = writer

		ctx.Next()

		if !isCacheableStatus(writer.status, cfg.CacheableStatus) {
			return
		}

		headers := make(map[string]string)
		for k, v := range writer.Header() {
			if len(v) > 0 {
				headers[k] = v[0]
			}
		}

		cached = CacheResponse{
			Status:  writer.status,
			Headers: headers,
			Body:    writer.body,
		}

		_ = c.Set(ctx.Request.Context(), key, cached, cfg.TTL)
		ctx.Header("X-Cache", "MISS")
	}
}
