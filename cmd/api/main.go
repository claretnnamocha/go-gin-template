package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/yourusername/go-gin-template/internal/handler"
	"github.com/yourusername/go-gin-template/internal/middleware"
	"github.com/yourusername/go-gin-template/internal/model"
	"github.com/yourusername/go-gin-template/internal/repository"
	"github.com/yourusername/go-gin-template/internal/service"
	"github.com/yourusername/go-gin-template/pkg/auth"
	"github.com/yourusername/go-gin-template/pkg/cache"
	"github.com/yourusername/go-gin-template/pkg/config"
	"github.com/yourusername/go-gin-template/pkg/database"
	"github.com/yourusername/go-gin-template/pkg/logger"
	"github.com/yourusername/go-gin-template/pkg/validator"
	"go.uber.org/zap"

	_ "github.com/yourusername/go-gin-template/docs"
)

// @title           API Documentation
// @version         1.0
// @description     Go Gin REST API

// @host            localhost:1954
// @BasePath        /api/v1
// @schemes         http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

const (
	Version = "1.0.0"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	if err := logger.Init(logger.Config{
		Level:  cfg.Log.Level,
		Format: cfg.Log.Format,
		Output: cfg.Log.Output,
	}); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting application",
		zap.String("name", cfg.App.Name),
		zap.String("version", Version),
		zap.String("env", cfg.App.Env),
	)

	// Initialize validator
	validator.Init()

	// Initialize database
	db, err := database.NewPostgresDB(cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Run auto-migrations
	if err := db.AutoMigrate(&model.User{}); err != nil {
		logger.Fatal("Failed to run database migrations", zap.Error(err))
	}

	// Initialize Redis cache (optional - continues if Redis is not available)
	var redisCache *cache.Cache
	redisCache, err = cache.NewRedisCache(cfg.Redis)
	if err != nil {
		logger.Warn("Failed to connect to Redis, continuing without cache", zap.Error(err))
		redisCache = nil
	} else {
		defer redisCache.Close()
	}

	// Initialize JWT manager
	jwtManager := auth.NewJWTManager(
		cfg.JWT.Secret,
		cfg.JWT.ExpirationHours,
		cfg.JWT.RefreshExpirationHours,
	)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)

	// Initialize services
	userService := service.NewUserService(userRepo, jwtManager)

	// Initialize handlers
	userHandler := handler.NewUserHandler(userService)
	healthHandler := handler.NewHealthHandler(db, redisCache, Version)

	// Setup Gin
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Global middleware
	router.Use(middleware.Recovery())
	router.Use(middleware.RequestID())
	router.Use(middleware.Logger())
	router.Use(middleware.CORS(cfg.CORS))
	router.Use(middleware.SecurityHeaders())

	// Rate limiting
	rateLimiter := middleware.NewRateLimitFromConfig(cfg.RateLimit, redisCache)
	router.Use(middleware.RateLimit(rateLimiter))

	// Health check route
	router.GET("/health", healthHandler.Health)

	// API routes
	api := router.Group("/api")
	{
		// Swagger documentation at /api/docs
		api.GET("/docs/*any", ginSwagger.WrapHandler(
			swaggerFiles.Handler,
			ginSwagger.URL("/api/docs/doc.json"),
			ginSwagger.DefaultModelsExpandDepth(-1),
		))

		// Redirect /api/swagger to /api/docs for convenience
		api.GET("/swagger", func(c *gin.Context) {
			c.Redirect(http.StatusMovedPermanently, "/api/docs/index.html")
		})

		// API v1 routes
		v1 := api.Group("/v1")
		{
			// Auth middleware
			authMiddleware := middleware.Auth(jwtManager)
			adminMiddleware := middleware.RequireRole("admin")

			// Register routes
			userHandler.RegisterRoutes(v1, authMiddleware, adminMiddleware)
		}
	}

	// Error handling middleware (should be last)
	router.Use(middleware.ErrorHandler())

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.App.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Server starting",
			zap.Int("port", cfg.App.Port),
			zap.String("env", cfg.App.Env),
		)
		logger.Info("Swagger docs available at",
			zap.String("url", fmt.Sprintf("http://localhost:%d/api/docs", cfg.App.Port)),
		)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server stopped")
}
