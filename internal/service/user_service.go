package service

import (
	"context"
	"errors"

	"go-gin-template/internal/dto"
	"go-gin-template/internal/model"
	"go-gin-template/internal/repository"
	"go-gin-template/pkg/auth"
	"go-gin-template/pkg/utils"
)

// Common errors
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailExists        = errors.New("email already exists")
	ErrUsernameExists     = errors.New("username already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrUserNotActive      = errors.New("user is not active")
)

// UserService interface defines user service operations
type UserService interface {
	// Register creates a new user account
	Register(ctx context.Context, req *dto.RegisterRequest) (*model.User, error)

	// Login authenticates a user and returns tokens
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.TokenResponse, error)

	// RefreshToken refreshes an access token
	RefreshToken(ctx context.Context, refreshToken string) (*dto.TokenResponse, error)

	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id uint) (*model.User, error)

	// GetByEmail retrieves a user by email
	GetByEmail(ctx context.Context, email string) (*model.User, error)

	// Update updates a user's profile
	Update(ctx context.Context, id uint, req *dto.UpdateUserRequest) (*model.User, error)

	// ChangePassword changes a user's password
	ChangePassword(ctx context.Context, id uint, req *dto.ChangePasswordRequest) error

	// Delete soft deletes a user
	Delete(ctx context.Context, id uint) error

	// List retrieves a paginated list of users
	List(ctx context.Context, query *dto.UserListQuery) ([]*model.User, int64, error)
}

// userService implements UserService interface
type userService struct {
	userRepo   repository.UserRepository
	jwtManager *auth.JWTManager
}

// NewUserService creates a new user service
func NewUserService(userRepo repository.UserRepository, jwtManager *auth.JWTManager) UserService {
	return &userService{
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

// Register creates a new user account
func (s *userService) Register(ctx context.Context, req *dto.RegisterRequest) (*model.User, error) {
	// Check if email exists
	exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrEmailExists
	}

	// Check if username exists
	exists, err = s.userRepo.ExistsByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUsernameExists
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &model.User{
		Email:     req.Email,
		Username:  req.Username,
		Password:  hashedPassword,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      "user",
		IsActive:  true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// Login authenticates a user and returns tokens
func (s *userService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.TokenResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	// Check if user is active
	if !user.IsActive {
		return nil, ErrUserNotActive
	}

	// Verify password
	if !utils.CheckPassword(req.Password, user.Password) {
		return nil, ErrInvalidCredentials
	}

	// Update last login
	_ = s.userRepo.UpdateLastLogin(ctx, user.ID)

	// Generate tokens
	accessToken, refreshToken, err := s.jwtManager.GenerateTokenPair(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	return &dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(s.jwtManager.GetTokenExpiration().Seconds()),
	}, nil
}

// RefreshToken refreshes an access token
func (s *userService) RefreshToken(ctx context.Context, refreshToken string) (*dto.TokenResponse, error) {
	// Validate refresh token
	claims, err := s.jwtManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Get user to ensure they still exist and are active
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	if !user.IsActive {
		return nil, ErrUserNotActive
	}

	// Generate new tokens
	accessToken, newRefreshToken, err := s.jwtManager.GenerateTokenPair(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	return &dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(s.jwtManager.GetTokenExpiration().Seconds()),
	}, nil
}

// GetByID retrieves a user by ID
func (s *userService) GetByID(ctx context.Context, id uint) (*model.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// GetByEmail retrieves a user by email
func (s *userService) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// Update updates a user's profile
func (s *userService) Update(ctx context.Context, id uint, req *dto.UpdateUserRequest) (*model.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Update fields if provided
	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		user.LastName = *req.LastName
	}
	if req.Bio != nil {
		user.Bio = req.Bio
	}
	if req.Phone != nil {
		user.Phone = req.Phone
	}
	if req.Avatar != nil {
		user.Avatar = req.Avatar
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// ChangePassword changes a user's password
func (s *userService) ChangePassword(ctx context.Context, id uint, req *dto.ChangePasswordRequest) error {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Verify old password
	if !utils.CheckPassword(req.OldPassword, user.Password) {
		return ErrInvalidPassword
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	user.Password = hashedPassword
	return s.userRepo.Update(ctx, user)
}

// Delete soft deletes a user
func (s *userService) Delete(ctx context.Context, id uint) error {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	return s.userRepo.Delete(ctx, id)
}

// List retrieves a paginated list of users
func (s *userService) List(ctx context.Context, query *dto.UserListQuery) ([]*model.User, int64, error) {
	query.SetDefaults()

	params := repository.UserListParams{
		Offset:    (query.Page - 1) * query.PageSize,
		Limit:     query.PageSize,
		Search:    query.Search,
		Role:      query.Role,
		IsActive:  query.IsActive,
		SortBy:    query.SortBy,
		SortOrder: query.SortOrder,
	}

	return s.userRepo.List(ctx, params)
}
