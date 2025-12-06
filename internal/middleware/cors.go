package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go-gin-template/pkg/config"
)

// CORS creates a CORS middleware with the given configuration
func CORS(cfg config.CORSConfig) gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     cfg.AllowedMethods,
		AllowHeaders:     cfg.AllowedHeaders,
		ExposeHeaders:    []string{"Content-Length", "Content-Type", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * 60 * 60, // 12 hours
	})
}

// DefaultCORS creates a CORS middleware with sensible defaults
func DefaultCORS() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Request-ID",
			"X-Requested-With",
		},
		ExposeHeaders:    []string{"Content-Length", "Content-Type", "X-Request-ID"},
		AllowCredentials: false,
		MaxAge:           12 * 60 * 60,
	})
}

// StrictCORS creates a strict CORS middleware for production
func StrictCORS(allowedOrigins []string) gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * 60 * 60,
	})
}
