package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"go-gin-template/pkg/auth"
	"go-gin-template/pkg/response"
)

// ContextKey is a type for context keys
type ContextKey string

const (
	// UserIDKey is the context key for user ID
	UserIDKey ContextKey = "user_id"
	// UserEmailKey is the context key for user email
	UserEmailKey ContextKey = "user_email"
	// UserRoleKey is the context key for user role
	UserRoleKey ContextKey = "user_role"
	// ClaimsKey is the context key for JWT claims
	ClaimsKey ContextKey = "claims"
)

// Auth creates an authentication middleware
func Auth(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "Authorization header is required")
			c.Abort()
			return
		}

		// Check for Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.Unauthorized(c, "Invalid authorization header format")
			c.Abort()
			return
		}

		tokenString := parts[1]

		claims, err := jwtManager.ValidateAccessToken(tokenString)
		if err != nil {
			switch err {
			case auth.ErrExpiredToken:
				response.Unauthorized(c, "Token has expired")
			case auth.ErrInvalidToken:
				response.Unauthorized(c, "Invalid token")
			default:
				response.Unauthorized(c, "Authentication failed")
			}
			c.Abort()
			return
		}

		// Set user info in context
		c.Set(string(UserIDKey), claims.UserID)
		c.Set(string(UserEmailKey), claims.Email)
		c.Set(string(UserRoleKey), claims.Role)
		c.Set(string(ClaimsKey), claims)

		c.Next()
	}
}

// OptionalAuth is like Auth but doesn't require authentication
func OptionalAuth(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.Next()
			return
		}

		tokenString := parts[1]

		claims, err := jwtManager.ValidateAccessToken(tokenString)
		if err == nil {
			c.Set(string(UserIDKey), claims.UserID)
			c.Set(string(UserEmailKey), claims.Email)
			c.Set(string(UserRoleKey), claims.Role)
			c.Set(string(ClaimsKey), claims)
		}

		c.Next()
	}
}

// RequireRole creates a middleware that requires specific roles
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get(string(UserRoleKey))
		if !exists {
			response.Unauthorized(c, "Authentication required")
			c.Abort()
			return
		}

		role, ok := userRole.(string)
		if !ok {
			response.InternalServerError(c, "Invalid role type")
			c.Abort()
			return
		}

		// Check if user has one of the required roles
		hasRole := false
		for _, r := range roles {
			if role == r {
				hasRole = true
				break
			}
		}

		if !hasRole {
			response.Forbidden(c, "Insufficient permissions")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequirePermission creates a middleware for specific permissions
// This is a basic implementation - can be extended for more complex RBAC
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get(string(UserRoleKey))
		if !exists {
			response.Unauthorized(c, "Authentication required")
			c.Abort()
			return
		}

		role, ok := userRole.(string)
		if !ok {
			response.InternalServerError(c, "Invalid role type")
			c.Abort()
			return
		}

		// Basic role-based permission check
		// This should be extended based on your permission system
		if !hasPermission(role, permission) {
			response.Forbidden(c, "Insufficient permissions")
			c.Abort()
			return
		}

		c.Next()
	}
}

// hasPermission checks if a role has a specific permission
// This is a basic implementation - extend based on your needs
func hasPermission(role, permission string) bool {
	// Define role permissions
	permissions := map[string][]string{
		"admin": {
			"users:read", "users:write", "users:delete",
			"posts:read", "posts:write", "posts:delete",
			"admin:access",
		},
		"moderator": {
			"users:read",
			"posts:read", "posts:write", "posts:delete",
		},
		"user": {
			"users:read",
			"posts:read", "posts:write",
		},
	}

	rolePermissions, exists := permissions[role]
	if !exists {
		return false
	}

	for _, p := range rolePermissions {
		if p == permission {
			return true
		}
	}

	return false
}

// GetUserID retrieves the user ID from context
func GetUserID(c *gin.Context) (uint, bool) {
	id, exists := c.Get(string(UserIDKey))
	if !exists {
		return 0, false
	}
	userID, ok := id.(uint)
	return userID, ok
}

// GetUserEmail retrieves the user email from context
func GetUserEmail(c *gin.Context) (string, bool) {
	email, exists := c.Get(string(UserEmailKey))
	if !exists {
		return "", false
	}
	userEmail, ok := email.(string)
	return userEmail, ok
}

// GetUserRole retrieves the user role from context
func GetUserRole(c *gin.Context) (string, bool) {
	role, exists := c.Get(string(UserRoleKey))
	if !exists {
		return "", false
	}
	userRole, ok := role.(string)
	return userRole, ok
}

// GetClaims retrieves the JWT claims from context
func GetClaims(c *gin.Context) (*auth.Claims, bool) {
	claims, exists := c.Get(string(ClaimsKey))
	if !exists {
		return nil, false
	}
	jwtClaims, ok := claims.(*auth.Claims)
	return jwtClaims, ok
}
