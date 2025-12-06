package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/go-gin-template/internal/dto"
	"github.com/yourusername/go-gin-template/internal/middleware"
	"github.com/yourusername/go-gin-template/internal/service"
	"github.com/yourusername/go-gin-template/pkg/response"
	"github.com/yourusername/go-gin-template/pkg/validator"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	userService service.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// Register godoc
// @Summary      Register a new user
// @Description  Create a new user account
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.RegisterRequest true "Registration details"
// @Success      201 {object} response.APIResponse
// @Failure      400 {object} response.APIResponse
// @Failure      409 {object} response.APIResponse
// @Router       /auth/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errs := validator.FormatValidationErrors(err)
		if len(errs) > 0 {
			response.ValidationErrors(c, errs)
			return
		}
		response.BadRequest(c, "Invalid request body", nil)
		return
	}

	user, err := h.userService.Register(c.Request.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEmailExists):
			response.Conflict(c, "Email already exists")
		case errors.Is(err, service.ErrUsernameExists):
			response.Conflict(c, "Username already exists")
		default:
			response.InternalServerError(c, "Failed to register user")
		}
		return
	}

	response.Created(c, user.ToResponse(), "User registered successfully")
}

// Login godoc
// @Summary      User login
// @Description  Authenticate user and return tokens
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.LoginRequest true "Login credentials"
// @Success      200 {object} response.APIResponse
// @Failure      400 {object} response.APIResponse
// @Failure      401 {object} response.APIResponse
// @Router       /auth/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errs := validator.FormatValidationErrors(err)
		if len(errs) > 0 {
			response.ValidationErrors(c, errs)
			return
		}
		response.BadRequest(c, "Invalid request body", nil)
		return
	}

	tokens, err := h.userService.Login(c.Request.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			response.Unauthorized(c, "Invalid email or password")
		case errors.Is(err, service.ErrUserNotActive):
			response.Unauthorized(c, "Account is not active")
		default:
			response.InternalServerError(c, "Failed to login")
		}
		return
	}

	response.OK(c, tokens, "Login successful")
}

// RefreshToken godoc
// @Summary      Refresh access token
// @Description  Get a new access token using refresh token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.RefreshTokenRequest true "Refresh token"
// @Success      200 {object} response.APIResponse
// @Failure      401 {object} response.APIResponse
// @Router       /auth/refresh [post]
func (h *UserHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", nil)
		return
	}

	tokens, err := h.userService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Unauthorized(c, "Invalid or expired refresh token")
		return
	}

	response.OK(c, tokens, "Token refreshed successfully")
}

// GetProfile godoc
// @Summary      Get current user profile
// @Description  Get the profile of the authenticated user
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} response.APIResponse
// @Failure      401 {object} response.APIResponse
// @Router       /users/me [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "")
		return
	}

	user, err := h.userService.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.InternalServerError(c, "Failed to get user profile")
		return
	}

	response.OK(c, user.ToResponse(), "Profile retrieved successfully")
}

// UpdateProfile godoc
// @Summary      Update current user profile
// @Description  Update the profile of the authenticated user
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.UpdateUserRequest true "Profile update details"
// @Success      200 {object} response.APIResponse
// @Failure      400 {object} response.APIResponse
// @Failure      401 {object} response.APIResponse
// @Router       /users/me [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "")
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errs := validator.FormatValidationErrors(err)
		if len(errs) > 0 {
			response.ValidationErrors(c, errs)
			return
		}
		response.BadRequest(c, "Invalid request body", nil)
		return
	}

	user, err := h.userService.Update(c.Request.Context(), userID, &req)
	if err != nil {
		response.InternalServerError(c, "Failed to update profile")
		return
	}

	response.OK(c, user.ToResponse(), "Profile updated successfully")
}

// ChangePassword godoc
// @Summary      Change password
// @Description  Change the password of the authenticated user
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.ChangePasswordRequest true "Password change details"
// @Success      200 {object} response.APIResponse
// @Failure      400 {object} response.APIResponse
// @Failure      401 {object} response.APIResponse
// @Router       /users/me/password [put]
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "")
		return
	}

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errs := validator.FormatValidationErrors(err)
		if len(errs) > 0 {
			response.ValidationErrors(c, errs)
			return
		}
		response.BadRequest(c, "Invalid request body", nil)
		return
	}

	err := h.userService.ChangePassword(c.Request.Context(), userID, &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidPassword):
			response.BadRequest(c, "Current password is incorrect", nil)
		default:
			response.InternalServerError(c, "Failed to change password")
		}
		return
	}

	response.OK(c, nil, "Password changed successfully")
}

// GetUser godoc
// @Summary      Get user by ID
// @Description  Get a user by their ID (admin only)
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "User ID"
// @Success      200 {object} response.APIResponse
// @Failure      404 {object} response.APIResponse
// @Router       /users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid user ID", nil)
		return
	}

	user, err := h.userService.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			response.NotFound(c, "User not found")
			return
		}
		response.InternalServerError(c, "Failed to get user")
		return
	}

	response.OK(c, user.ToResponse(), "User retrieved successfully")
}

// ListUsers godoc
// @Summary      List users
// @Description  Get a paginated list of users (admin only)
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        page query int false "Page number" default(1)
// @Param        limit query int false "Items per page" default(10)
// @Param        search query string false "Search term"
// @Param        role query string false "Filter by role"
// @Success      200 {object} response.APIResponse
// @Failure      401 {object} response.APIResponse
// @Failure      403 {object} response.APIResponse
// @Router       /users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	var query dto.UserListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		errs := validator.FormatValidationErrors(err)
		if len(errs) > 0 {
			response.ValidationErrors(c, errs)
			return
		}
		response.BadRequest(c, "Invalid query parameters", nil)
		return
	}

	users, total, err := h.userService.List(c.Request.Context(), &query)
	if err != nil {
		response.InternalServerError(c, "Failed to list users")
		return
	}

	// Convert to response format
	userList := make([]interface{}, len(users))
	for i, user := range users {
		userList[i] = user.ToResponse()
	}

	query.SetDefaults()
	response.Paginated(c, userList, query.Page, query.PageSize, int(total), "Users retrieved successfully")
}

// DeleteUser godoc
// @Summary      Delete user
// @Description  Soft delete a user by ID (admin only)
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "User ID"
// @Success      200 {object} response.APIResponse
// @Failure      404 {object} response.APIResponse
// @Router       /users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid user ID", nil)
		return
	}

	err = h.userService.Delete(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			response.NotFound(c, "User not found")
			return
		}
		response.InternalServerError(c, "Failed to delete user")
		return
	}

	response.OK(c, nil, "User deleted successfully")
}

// RegisterRoutes registers user routes
func (h *UserHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware gin.HandlerFunc, adminMiddleware gin.HandlerFunc) {
	// Public auth routes
	auth := router.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.POST("/refresh", h.RefreshToken)
	}

	// Protected user routes
	users := router.Group("/users")
	users.Use(authMiddleware)
	{
		// Current user routes
		users.GET("/me", h.GetProfile)
		users.PUT("/me", h.UpdateProfile)
		users.PUT("/me/password", h.ChangePassword)

		// Admin routes
		admin := users.Group("")
		admin.Use(adminMiddleware)
		{
			admin.GET("", h.ListUsers)
			admin.GET("/:id", h.GetUser)
			admin.DELETE("/:id", h.DeleteUser)
		}
	}
}
