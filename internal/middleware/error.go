package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go-gin-template/pkg/logger"
	"go-gin-template/pkg/response"
	"go.uber.org/zap"
)

// AppError represents an application error with status code
type AppError struct {
	Code       int    `json:"code"`
	Message    string `json:"message"`
	Err        error  `json:"-"`
	ErrorCode  string `json:"error_code,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError creates a new AppError
func NewAppError(code int, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// NewAppErrorWithCode creates a new AppError with an error code
func NewAppErrorWithCode(code int, errorCode, message string, err error) *AppError {
	return &AppError{
		Code:      code,
		Message:   message,
		Err:       err,
		ErrorCode: errorCode,
	}
}

// Common application errors
var (
	ErrBadRequest          = NewAppError(http.StatusBadRequest, "Bad request", nil)
	ErrUnauthorized        = NewAppError(http.StatusUnauthorized, "Unauthorized", nil)
	ErrForbidden           = NewAppError(http.StatusForbidden, "Forbidden", nil)
	ErrNotFound            = NewAppError(http.StatusNotFound, "Resource not found", nil)
	ErrConflict            = NewAppError(http.StatusConflict, "Resource conflict", nil)
	ErrUnprocessableEntity = NewAppError(http.StatusUnprocessableEntity, "Unprocessable entity", nil)
	ErrTooManyRequests     = NewAppError(http.StatusTooManyRequests, "Too many requests", nil)
	ErrInternalServer      = NewAppError(http.StatusInternalServerError, "Internal server error", nil)
	ErrServiceUnavailable  = NewAppError(http.StatusServiceUnavailable, "Service unavailable", nil)
)

// Error helper functions
func BadRequest(message string) *AppError {
	return NewAppError(http.StatusBadRequest, message, nil)
}

func Unauthorized(message string) *AppError {
	if message == "" {
		message = "Unauthorized"
	}
	return NewAppError(http.StatusUnauthorized, message, nil)
}

func Forbidden(message string) *AppError {
	if message == "" {
		message = "Forbidden"
	}
	return NewAppError(http.StatusForbidden, message, nil)
}

func NotFound(message string) *AppError {
	if message == "" {
		message = "Resource not found"
	}
	return NewAppError(http.StatusNotFound, message, nil)
}

func Conflict(message string) *AppError {
	if message == "" {
		message = "Resource conflict"
	}
	return NewAppError(http.StatusConflict, message, nil)
}

func InternalServerError(message string, err error) *AppError {
	if message == "" {
		message = "Internal server error"
	}
	return NewAppError(http.StatusInternalServerError, message, err)
}

// ErrorHandler is the global error handling middleware
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			var appErr *AppError
			if errors.As(err, &appErr) {
				// Log internal errors
				if appErr.Code >= 500 {
					logger.Error("Internal server error",
						zap.Error(appErr.Err),
						zap.String("message", appErr.Message),
						zap.String("request_id", c.GetString("request_id")),
						zap.String("path", c.Request.URL.Path),
						zap.String("method", c.Request.Method),
					)
				}

				resp := &response.APIResponse{
					Success:    false,
					StatusCode: appErr.Code,
					Message:    appErr.Message,
				}

				if appErr.ErrorCode != "" {
					resp.Data = map[string]string{"error_code": appErr.ErrorCode}
				}

				c.JSON(appErr.Code, resp)
				return
			}

			// Handle unknown errors
			logger.Error("Unhandled error",
				zap.Error(err),
				zap.String("request_id", c.GetString("request_id")),
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
			)

			response.InternalServerError(c, "An unexpected error occurred")
		}
	}
}

// Recovery middleware recovers from panics
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic with stack trace
				logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("stack", string(debug.Stack())),
					zap.String("request_id", c.GetString("request_id")),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
				)

				response.InternalServerError(c, "Internal server error")
				c.Abort()
			}
		}()

		c.Next()
	}
}
