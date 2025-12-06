package repository

import (
	"context"

	"github.com/yourusername/go-gin-template/internal/model"
)

// UserRepository interface defines user repository operations
type UserRepository interface {
	// Create creates a new user
	Create(ctx context.Context, user *model.User) error

	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id uint) (*model.User, error)

	// GetByEmail retrieves a user by email
	GetByEmail(ctx context.Context, email string) (*model.User, error)

	// GetByUsername retrieves a user by username
	GetByUsername(ctx context.Context, username string) (*model.User, error)

	// Update updates a user
	Update(ctx context.Context, user *model.User) error

	// Delete soft deletes a user
	Delete(ctx context.Context, id uint) error

	// List retrieves a paginated list of users
	List(ctx context.Context, params UserListParams) ([]*model.User, int64, error)

	// ExistsByEmail checks if a user with the given email exists
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	// ExistsByUsername checks if a user with the given username exists
	ExistsByUsername(ctx context.Context, username string) (bool, error)

	// UpdateLastLogin updates the last login timestamp
	UpdateLastLogin(ctx context.Context, id uint) error
}

// UserListParams contains parameters for listing users
type UserListParams struct {
	Offset    int
	Limit     int
	Search    string
	Role      string
	IsActive  *bool
	SortBy    string
	SortOrder string
}
