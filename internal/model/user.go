package model

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Email     string  `gorm:"uniqueIndex;size:255;not null" json:"email"`
	Username  string  `gorm:"uniqueIndex;size:100;not null" json:"username"`
	Password  string  `gorm:"size:255;not null" json:"-"`
	FirstName string  `gorm:"size:100" json:"first_name"`
	LastName  string  `gorm:"size:100" json:"last_name"`
	Role      string  `gorm:"size:50;default:user" json:"role"`
	IsActive  bool    `gorm:"default:true" json:"is_active"`
	Avatar    *string `gorm:"size:500" json:"avatar,omitempty"`
	Bio       *string `gorm:"type:text" json:"bio,omitempty"`
	Phone     *string `gorm:"size:20" json:"phone,omitempty"`
	
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	VerifiedAt  *time.Time `json:"verified_at,omitempty"`
}

// TableName specifies the table name for User model
func (User) TableName() string {
	return "users"
}

// BeforeCreate hook
func (u *User) BeforeCreate(tx *gorm.DB) error {
	// Any pre-creation logic
	return nil
}

// UserResponse represents the user data returned in API responses
type UserResponse struct {
	ID          uint       `json:"id"`
	Email       string     `json:"email"`
	Username    string     `json:"username"`
	FirstName   string     `json:"first_name"`
	LastName    string     `json:"last_name"`
	Role        string     `json:"role"`
	IsActive    bool       `json:"is_active"`
	Avatar      *string    `json:"avatar,omitempty"`
	Bio         *string    `json:"bio,omitempty"`
	Phone       *string    `json:"phone,omitempty"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	VerifiedAt  *time.Time `json:"verified_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ToResponse converts User to UserResponse
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:          u.ID,
		Email:       u.Email,
		Username:    u.Username,
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		Role:        u.Role,
		IsActive:    u.IsActive,
		Avatar:      u.Avatar,
		Bio:         u.Bio,
		Phone:       u.Phone,
		LastLoginAt: u.LastLoginAt,
		VerifiedAt:  u.VerifiedAt,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
}

// GetFullName returns the user's full name
func (u *User) GetFullName() string {
	if u.FirstName == "" && u.LastName == "" {
		return u.Username
	}
	return u.FirstName + " " + u.LastName
}

// IsAdmin checks if user has admin role
func (u *User) IsAdmin() bool {
	return u.Role == "admin"
}

// IsModerator checks if user has moderator role
func (u *User) IsModerator() bool {
	return u.Role == "moderator"
}

// IsVerified checks if user is verified
func (u *User) IsVerified() bool {
	return u.VerifiedAt != nil
}
