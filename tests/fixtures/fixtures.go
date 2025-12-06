package fixtures

import (
	"time"

	"github.com/yourusername/go-gin-template/internal/model"
	"github.com/yourusername/go-gin-template/pkg/utils"
)

// TestUser creates a test user with default values
func TestUser() *model.User {
	hashedPassword, _ := utils.HashPassword("Password123!")
	now := time.Now()

	return &model.User{
		ID:        1,
		Email:     "test@example.com",
		Username:  "testuser",
		Password:  hashedPassword,
		FirstName: "Test",
		LastName:  "User",
		Role:      "user",
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// TestAdmin creates a test admin user
func TestAdmin() *model.User {
	hashedPassword, _ := utils.HashPassword("AdminPassword123!")
	now := time.Now()

	return &model.User{
		ID:        2,
		Email:     "admin@example.com",
		Username:  "adminuser",
		Password:  hashedPassword,
		FirstName: "Admin",
		LastName:  "User",
		Role:      "admin",
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// TestUsers creates a slice of test users
func TestUsers(count int) []*model.User {
	users := make([]*model.User, count)
	hashedPassword, _ := utils.HashPassword("Password123!")
	now := time.Now()

	for i := 0; i < count; i++ {
		users[i] = &model.User{
			ID:        uint(i + 1),
			Email:     "user" + string(rune('0'+i)) + "@example.com",
			Username:  "user" + string(rune('0'+i)),
			Password:  hashedPassword,
			FirstName: "User",
			LastName:  string(rune('A' + i)),
			Role:      "user",
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		}
	}

	return users
}

// TestUserWithFields creates a test user with custom fields
func TestUserWithFields(email, username, role string) *model.User {
	hashedPassword, _ := utils.HashPassword("Password123!")
	now := time.Now()

	return &model.User{
		Email:     email,
		Username:  username,
		Password:  hashedPassword,
		FirstName: "Test",
		LastName:  "User",
		Role:      role,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
