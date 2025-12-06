package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"go-gin-template/internal/model"
	"go-gin-template/internal/repository"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

// NewMockUserRepository creates a new mock user repository
func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{}
}

// Create mocks the Create method
func (m *MockUserRepository) Create(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

// GetByID mocks the GetByID method
func (m *MockUserRepository) GetByID(ctx context.Context, id uint) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

// GetByEmail mocks the GetByEmail method
func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

// GetByUsername mocks the GetByUsername method
func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

// Update mocks the Update method
func (m *MockUserRepository) Update(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

// Delete mocks the Delete method
func (m *MockUserRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// List mocks the List method
func (m *MockUserRepository) List(ctx context.Context, params repository.UserListParams) ([]*model.User, int64, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*model.User), args.Get(1).(int64), args.Error(2)
}

// ExistsByEmail mocks the ExistsByEmail method
func (m *MockUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

// ExistsByUsername mocks the ExistsByUsername method
func (m *MockUserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	args := m.Called(ctx, username)
	return args.Bool(0), args.Error(1)
}

// UpdateLastLogin mocks the UpdateLastLogin method
func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Ensure MockUserRepository implements UserRepository
var _ repository.UserRepository = (*MockUserRepository)(nil)
