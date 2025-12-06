package unit

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/yourusername/go-gin-template/internal/dto"
	"github.com/yourusername/go-gin-template/internal/service"
	"github.com/yourusername/go-gin-template/pkg/auth"
	"github.com/yourusername/go-gin-template/tests/fixtures"
	"github.com/yourusername/go-gin-template/tests/mocks"
)

// UserServiceTestSuite defines the test suite for user service
type UserServiceTestSuite struct {
	suite.Suite
	mockRepo    *mocks.MockUserRepository
	jwtManager  *auth.JWTManager
	userService service.UserService
	ctx         context.Context
}

// SetupTest sets up the test suite
func (s *UserServiceTestSuite) SetupTest() {
	s.mockRepo = mocks.NewMockUserRepository()
	s.jwtManager = auth.NewJWTManager("test-secret", 24, 168)
	s.userService = service.NewUserService(s.mockRepo, s.jwtManager)
	s.ctx = context.Background()
}

// TearDownTest tears down the test suite
func (s *UserServiceTestSuite) TearDownTest() {
	s.mockRepo.AssertExpectations(s.T())
}

// TestRegisterSuccess tests successful user registration
func (s *UserServiceTestSuite) TestRegisterSuccess() {
	// Arrange
	req := &dto.RegisterRequest{
		Email:     "newuser@example.com",
		Username:  "newuser",
		Password:  "Password123!",
		FirstName: "New",
		LastName:  "User",
	}

	s.mockRepo.On("ExistsByEmail", s.ctx, req.Email).Return(false, nil)
	s.mockRepo.On("ExistsByUsername", s.ctx, req.Username).Return(false, nil)
	s.mockRepo.On("Create", s.ctx, mock.AnythingOfType("*model.User")).Return(nil)

	// Act
	user, err := s.userService.Register(s.ctx, req)

	// Assert
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), user)
	assert.Equal(s.T(), req.Email, user.Email)
	assert.Equal(s.T(), req.Username, user.Username)
	assert.Equal(s.T(), "user", user.Role)
}

// TestRegisterEmailExists tests registration with existing email
func (s *UserServiceTestSuite) TestRegisterEmailExists() {
	// Arrange
	req := &dto.RegisterRequest{
		Email:    "existing@example.com",
		Username: "newuser",
		Password: "Password123!",
	}

	s.mockRepo.On("ExistsByEmail", s.ctx, req.Email).Return(true, nil)

	// Act
	user, err := s.userService.Register(s.ctx, req)

	// Assert
	assert.Error(s.T(), err)
	assert.Nil(s.T(), user)
	assert.Equal(s.T(), service.ErrEmailExists, err)
}

// TestRegisterUsernameExists tests registration with existing username
func (s *UserServiceTestSuite) TestRegisterUsernameExists() {
	// Arrange
	req := &dto.RegisterRequest{
		Email:    "newuser@example.com",
		Username: "existinguser",
		Password: "Password123!",
	}

	s.mockRepo.On("ExistsByEmail", s.ctx, req.Email).Return(false, nil)
	s.mockRepo.On("ExistsByUsername", s.ctx, req.Username).Return(true, nil)

	// Act
	user, err := s.userService.Register(s.ctx, req)

	// Assert
	assert.Error(s.T(), err)
	assert.Nil(s.T(), user)
	assert.Equal(s.T(), service.ErrUsernameExists, err)
}

// TestLoginSuccess tests successful login
func (s *UserServiceTestSuite) TestLoginSuccess() {
	// Arrange
	testUser := fixtures.TestUser()
	req := &dto.LoginRequest{
		Email:    testUser.Email,
		Password: "Password123!",
	}

	s.mockRepo.On("GetByEmail", s.ctx, req.Email).Return(testUser, nil)
	s.mockRepo.On("UpdateLastLogin", s.ctx, testUser.ID).Return(nil)

	// Act
	tokens, err := s.userService.Login(s.ctx, req)

	// Assert
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), tokens)
	assert.NotEmpty(s.T(), tokens.AccessToken)
	assert.NotEmpty(s.T(), tokens.RefreshToken)
	assert.Equal(s.T(), "Bearer", tokens.TokenType)
}

// TestLoginUserNotFound tests login with non-existent user
func (s *UserServiceTestSuite) TestLoginUserNotFound() {
	// Arrange
	req := &dto.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "Password123!",
	}

	s.mockRepo.On("GetByEmail", s.ctx, req.Email).Return(nil, nil)

	// Act
	tokens, err := s.userService.Login(s.ctx, req)

	// Assert
	assert.Error(s.T(), err)
	assert.Nil(s.T(), tokens)
	assert.Equal(s.T(), service.ErrInvalidCredentials, err)
}

// TestLoginInvalidPassword tests login with wrong password
func (s *UserServiceTestSuite) TestLoginInvalidPassword() {
	// Arrange
	testUser := fixtures.TestUser()
	req := &dto.LoginRequest{
		Email:    testUser.Email,
		Password: "WrongPassword123!",
	}

	s.mockRepo.On("GetByEmail", s.ctx, req.Email).Return(testUser, nil)

	// Act
	tokens, err := s.userService.Login(s.ctx, req)

	// Assert
	assert.Error(s.T(), err)
	assert.Nil(s.T(), tokens)
	assert.Equal(s.T(), service.ErrInvalidCredentials, err)
}

// TestLoginInactiveUser tests login with inactive user
func (s *UserServiceTestSuite) TestLoginInactiveUser() {
	// Arrange
	testUser := fixtures.TestUser()
	testUser.IsActive = false
	req := &dto.LoginRequest{
		Email:    testUser.Email,
		Password: "Password123!",
	}

	s.mockRepo.On("GetByEmail", s.ctx, req.Email).Return(testUser, nil)

	// Act
	tokens, err := s.userService.Login(s.ctx, req)

	// Assert
	assert.Error(s.T(), err)
	assert.Nil(s.T(), tokens)
	assert.Equal(s.T(), service.ErrUserNotActive, err)
}

// TestGetByIDSuccess tests getting user by ID
func (s *UserServiceTestSuite) TestGetByIDSuccess() {
	// Arrange
	testUser := fixtures.TestUser()
	s.mockRepo.On("GetByID", s.ctx, testUser.ID).Return(testUser, nil)

	// Act
	user, err := s.userService.GetByID(s.ctx, testUser.ID)

	// Assert
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), user)
	assert.Equal(s.T(), testUser.ID, user.ID)
}

// TestGetByIDNotFound tests getting non-existent user
func (s *UserServiceTestSuite) TestGetByIDNotFound() {
	// Arrange
	s.mockRepo.On("GetByID", s.ctx, uint(999)).Return(nil, nil)

	// Act
	user, err := s.userService.GetByID(s.ctx, 999)

	// Assert
	assert.Error(s.T(), err)
	assert.Nil(s.T(), user)
	assert.Equal(s.T(), service.ErrUserNotFound, err)
}

// TestUpdateSuccess tests successful user update
func (s *UserServiceTestSuite) TestUpdateSuccess() {
	// Arrange
	testUser := fixtures.TestUser()
	firstName := "Updated"
	req := &dto.UpdateUserRequest{
		FirstName: &firstName,
	}

	s.mockRepo.On("GetByID", s.ctx, testUser.ID).Return(testUser, nil)
	s.mockRepo.On("Update", s.ctx, mock.AnythingOfType("*model.User")).Return(nil)

	// Act
	user, err := s.userService.Update(s.ctx, testUser.ID, req)

	// Assert
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), user)
	assert.Equal(s.T(), firstName, user.FirstName)
}

// TestDeleteSuccess tests successful user deletion
func (s *UserServiceTestSuite) TestDeleteSuccess() {
	// Arrange
	testUser := fixtures.TestUser()
	s.mockRepo.On("GetByID", s.ctx, testUser.ID).Return(testUser, nil)
	s.mockRepo.On("Delete", s.ctx, testUser.ID).Return(nil)

	// Act
	err := s.userService.Delete(s.ctx, testUser.ID)

	// Assert
	assert.NoError(s.T(), err)
}

// TestListUsers tests listing users
func (s *UserServiceTestSuite) TestListUsers() {
	// Arrange
	testUsers := fixtures.TestUsers(5)
	query := &dto.UserListQuery{
		Page:     1,
		PageSize: 10,
	}

	s.mockRepo.On("List", s.ctx, mock.AnythingOfType("repository.UserListParams")).
		Return(testUsers, int64(5), nil)

	// Act
	users, total, err := s.userService.List(s.ctx, query)

	// Assert
	assert.NoError(s.T(), err)
	assert.Len(s.T(), users, 5)
	assert.Equal(s.T(), int64(5), total)
}

// TestUserServiceSuite runs the test suite
func TestUserServiceSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}
