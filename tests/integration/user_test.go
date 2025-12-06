package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/yourusername/go-gin-template/internal/dto"
	"github.com/yourusername/go-gin-template/internal/handler"
	"github.com/yourusername/go-gin-template/internal/service"
	"github.com/yourusername/go-gin-template/pkg/auth"
	"github.com/yourusername/go-gin-template/pkg/response"
	"github.com/yourusername/go-gin-template/tests/fixtures"
	"github.com/yourusername/go-gin-template/tests/mocks"
)

// UserHandlerTestSuite defines the test suite for user handler
type UserHandlerTestSuite struct {
	suite.Suite
	router      *gin.Engine
	userHandler *handler.UserHandler
	mockRepo    *mocks.MockUserRepository
	jwtManager  *auth.JWTManager
}

// SetupTest sets up the test suite
func (s *UserHandlerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)

	s.mockRepo = mocks.NewMockUserRepository()
	s.jwtManager = auth.NewJWTManager("test-secret", 24, 168)

	userService := service.NewUserService(s.mockRepo, s.jwtManager)
	s.userHandler = handler.NewUserHandler(userService)

	s.router = gin.New()
	api := s.router.Group("/api/v1")
	s.userHandler.RegisterRoutes(api, func(c *gin.Context) { c.Next() }, func(c *gin.Context) { c.Next() })
}

// TearDownTest tears down the test suite
func (s *UserHandlerTestSuite) TearDownTest() {
	s.mockRepo.AssertExpectations(s.T())
}

// TestRegisterSuccess tests successful user registration
func (s *UserHandlerTestSuite) TestRegisterSuccess() {
	// Arrange
	req := dto.RegisterRequest{
		Email:     "newuser@example.com",
		Username:  "newuser",
		Password:  "Password123!",
		FirstName: "New",
		LastName:  "User",
	}

	s.mockRepo.On("ExistsByEmail", mock.Anything, req.Email).Return(false, nil)
	s.mockRepo.On("ExistsByUsername", mock.Anything, req.Username).Return(false, nil)
	s.mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.User")).Return(nil)

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	// Act
	s.router.ServeHTTP(w, httpReq)

	// Assert
	assert.Equal(s.T(), http.StatusCreated, w.Code)

	var resp response.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(s.T(), err)
	assert.True(s.T(), resp.Success)
	assert.Equal(s.T(), "User registered successfully", resp.Message)
}

// TestRegisterEmailExists tests registration with existing email
func (s *UserHandlerTestSuite) TestRegisterEmailExists() {
	// Arrange
	req := dto.RegisterRequest{
		Email:    "existing@example.com",
		Username: "newuser",
		Password: "Password123!",
	}

	s.mockRepo.On("ExistsByEmail", mock.Anything, req.Email).Return(true, nil)

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	// Act
	s.router.ServeHTTP(w, httpReq)

	// Assert
	assert.Equal(s.T(), http.StatusConflict, w.Code)

	var resp response.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(s.T(), err)
	assert.False(s.T(), resp.Success)
	assert.Equal(s.T(), "Email already exists", resp.Message)
}

// TestRegisterValidationError tests registration with invalid data
func (s *UserHandlerTestSuite) TestRegisterValidationError() {
	// Arrange - missing required fields
	req := dto.RegisterRequest{
		Email: "invalid-email",
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	// Act
	s.router.ServeHTTP(w, httpReq)

	// Assert
	assert.Equal(s.T(), http.StatusBadRequest, w.Code)
}

// TestLoginSuccess tests successful login
func (s *UserHandlerTestSuite) TestLoginSuccess() {
	// Arrange
	testUser := fixtures.TestUser()

	req := dto.LoginRequest{
		Email:    testUser.Email,
		Password: "Password123!",
	}

	s.mockRepo.On("GetByEmail", mock.Anything, req.Email).Return(testUser, nil)
	s.mockRepo.On("UpdateLastLogin", mock.Anything, testUser.ID).Return(nil)

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	// Act
	s.router.ServeHTTP(w, httpReq)

	// Assert
	assert.Equal(s.T(), http.StatusOK, w.Code)

	var resp response.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(s.T(), err)
	assert.True(s.T(), resp.Success)
}

// TestLoginInvalidCredentials tests login with invalid credentials
func (s *UserHandlerTestSuite) TestLoginInvalidCredentials() {
	// Arrange
	req := dto.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "wrongpassword",
	}

	s.mockRepo.On("GetByEmail", mock.Anything, req.Email).Return(nil, nil)

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	// Act
	s.router.ServeHTTP(w, httpReq)

	// Assert
	assert.Equal(s.T(), http.StatusUnauthorized, w.Code)
}

// TestUserHandlerSuite runs the test suite
func TestUserHandlerSuite(t *testing.T) {
	suite.Run(t, new(UserHandlerTestSuite))
}
