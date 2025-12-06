package response

import (
	"github.com/gin-gonic/gin"
)

// APIResponse represents a standardized API response
type APIResponse struct {
	Success    bool        `json:"success"`
	StatusCode int         `json:"statusCode"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data,omitempty"`
	Metadata   interface{} `json:"metadata,omitempty"`
}

// PaginationMetadata holds pagination information
type PaginationMetadata struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ErrorDetail holds detailed error information
type ErrorDetail struct {
	Code    string            `json:"code,omitempty"`
	Details []ValidationError `json:"details,omitempty"`
}

// Success sends a success response
func Success(c *gin.Context, data interface{}, message string, statusCode int) {
	c.JSON(statusCode, APIResponse{
		Success:    true,
		StatusCode: statusCode,
		Message:    message,
		Data:       data,
	})
}

// Error sends an error response
func Error(c *gin.Context, message string, statusCode int, data interface{}) {
	c.JSON(statusCode, APIResponse{
		Success:    false,
		StatusCode: statusCode,
		Message:    message,
		Data:       data,
	})
}

// Paginated sends a paginated response
func Paginated(c *gin.Context, data interface{}, page, limit, total int, message string) {
	totalPages := (total + limit - 1) / limit

	c.JSON(200, APIResponse{
		Success:    true,
		StatusCode: 200,
		Message:    message,
		Data:       data,
		Metadata: PaginationMetadata{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

// Created sends a 201 created response
func Created(c *gin.Context, data interface{}, message string) {
	Success(c, data, message, 201)
}

// OK sends a 200 OK response
func OK(c *gin.Context, data interface{}, message string) {
	Success(c, data, message, 200)
}

// NoContent sends a 204 no content response
func NoContent(c *gin.Context) {
	c.Status(204)
}

// BadRequest sends a 400 bad request response
func BadRequest(c *gin.Context, message string, data interface{}) {
	Error(c, message, 400, data)
}

// Unauthorized sends a 401 unauthorized response
func Unauthorized(c *gin.Context, message string) {
	if message == "" {
		message = "Unauthorized"
	}
	Error(c, message, 401, nil)
}

// Forbidden sends a 403 forbidden response
func Forbidden(c *gin.Context, message string) {
	if message == "" {
		message = "Forbidden"
	}
	Error(c, message, 403, nil)
}

// NotFound sends a 404 not found response
func NotFound(c *gin.Context, message string) {
	if message == "" {
		message = "Resource not found"
	}
	Error(c, message, 404, nil)
}

// Conflict sends a 409 conflict response
func Conflict(c *gin.Context, message string) {
	if message == "" {
		message = "Resource conflict"
	}
	Error(c, message, 409, nil)
}

// UnprocessableEntity sends a 422 unprocessable entity response
func UnprocessableEntity(c *gin.Context, message string, data interface{}) {
	Error(c, message, 422, data)
}

// TooManyRequests sends a 429 too many requests response
func TooManyRequests(c *gin.Context, message string) {
	if message == "" {
		message = "Too many requests"
	}
	Error(c, message, 429, nil)
}

// InternalServerError sends a 500 internal server error response
func InternalServerError(c *gin.Context, message string) {
	if message == "" {
		message = "Internal server error"
	}
	Error(c, message, 500, nil)
}

// ServiceUnavailable sends a 503 service unavailable response
func ServiceUnavailable(c *gin.Context, message string) {
	if message == "" {
		message = "Service unavailable"
	}
	Error(c, message, 503, nil)
}

// ValidationErrors sends a validation error response
func ValidationErrors(c *gin.Context, errors []ValidationError) {
	Error(c, "Validation failed", 400, ErrorDetail{
		Code:    "VALIDATION_ERROR",
		Details: errors,
	})
}

// NewPaginationMetadata creates pagination metadata
func NewPaginationMetadata(page, limit, total int) PaginationMetadata {
	totalPages := (total + limit - 1) / limit

	return PaginationMetadata{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}
}
