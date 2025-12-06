package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/go-gin-template/pkg/cache"
	"github.com/yourusername/go-gin-template/pkg/database"
	"github.com/yourusername/go-gin-template/pkg/response"
)

// HealthStatus represents the health status of a service
type HealthStatus string

const (
	StatusHealthy   HealthStatus = "healthy"
	StatusUnhealthy HealthStatus = "unhealthy"
	StatusDegraded  HealthStatus = "degraded"
)

// ComponentHealth represents the health of a single component
type ComponentHealth struct {
	Status  HealthStatus `json:"status"`
	Message string       `json:"message,omitempty"`
	Latency string       `json:"latency,omitempty"`
}

// HealthResponse represents the overall health response
type HealthResponse struct {
	Status     HealthStatus               `json:"status"`
	Version    string                     `json:"version"`
	Timestamp  string                     `json:"timestamp"`
	Components map[string]ComponentHealth `json:"components"`
}

// HealthHandler handles health check requests
type HealthHandler struct {
	db      *database.Database
	cache   *cache.Cache
	version string
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db *database.Database, cache *cache.Cache, version string) *HealthHandler {
	return &HealthHandler{
		db:      db,
		cache:   cache,
		version: version,
	}
}

// Health godoc
// @Summary      Health check
// @Description  Check the health status of the service and its dependencies
// @Tags         health
// @Accept       json
// @Produce      json
// @Success      200 {object} HealthResponse
// @Failure      503 {object} HealthResponse
// @Router       /health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp := HealthResponse{
		Status:     StatusHealthy,
		Version:    h.version,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		Components: make(map[string]ComponentHealth),
	}

	// Check database health
	dbHealth := h.checkDatabase(ctx)
	resp.Components["database"] = dbHealth
	if dbHealth.Status != StatusHealthy {
		resp.Status = StatusDegraded
	}

	// Check Redis health
	if h.cache != nil {
		redisHealth := h.checkRedis(ctx)
		resp.Components["redis"] = redisHealth
		if redisHealth.Status != StatusHealthy && resp.Status == StatusHealthy {
			resp.Status = StatusDegraded
		}
	}

	statusCode := http.StatusOK
	if resp.Status == StatusUnhealthy {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, resp)
}

// Liveness godoc
// @Summary      Liveness probe
// @Description  Simple liveness check for Kubernetes
// @Tags         health
// @Accept       json
// @Produce      json
// @Success      200 {object} map[string]string
// @Router       /health/live [get]
func (h *HealthHandler) Liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "alive",
	})
}

// Readiness godoc
// @Summary      Readiness probe
// @Description  Readiness check for Kubernetes - checks if the service can accept traffic
// @Tags         health
// @Accept       json
// @Produce      json
// @Success      200 {object} map[string]string
// @Failure      503 {object} map[string]string
// @Router       /health/ready [get]
func (h *HealthHandler) Readiness(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	// Check database is ready
	if h.db != nil {
		if err := h.db.HealthCheck(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  "not ready",
				"message": "database is not available",
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}

// checkDatabase checks database health
func (h *HealthHandler) checkDatabase(ctx context.Context) ComponentHealth {
	if h.db == nil {
		return ComponentHealth{
			Status:  StatusUnhealthy,
			Message: "database not configured",
		}
	}

	start := time.Now()
	err := h.db.HealthCheck(ctx)
	latency := time.Since(start)

	if err != nil {
		return ComponentHealth{
			Status:  StatusUnhealthy,
			Message: err.Error(),
			Latency: latency.String(),
		}
	}

	return ComponentHealth{
		Status:  StatusHealthy,
		Message: "connected",
		Latency: latency.String(),
	}
}

// checkRedis checks Redis health
func (h *HealthHandler) checkRedis(ctx context.Context) ComponentHealth {
	if h.cache == nil {
		return ComponentHealth{
			Status:  StatusUnhealthy,
			Message: "redis not configured",
		}
	}

	start := time.Now()
	err := h.cache.HealthCheck(ctx)
	latency := time.Since(start)

	if err != nil {
		return ComponentHealth{
			Status:  StatusUnhealthy,
			Message: err.Error(),
			Latency: latency.String(),
		}
	}

	return ComponentHealth{
		Status:  StatusHealthy,
		Message: "connected",
		Latency: latency.String(),
	}
}

// RegisterRoutes registers health check routes
func (h *HealthHandler) RegisterRoutes(router *gin.Engine) {
	router.GET("/health", h.Health)
	router.GET("/health/live", h.Liveness)
	router.GET("/health/ready", h.Readiness)
}

// SimpleHealth returns a simple health response for basic checks
func SimpleHealth(c *gin.Context) {
	response.OK(c, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}, "OK")
}
