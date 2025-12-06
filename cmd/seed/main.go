package main

import (
	"fmt"
	"log"
	"os"

	"go-gin-template/internal/model"
	"go-gin-template/pkg/config"
	"go-gin-template/pkg/database"
	"go-gin-template/pkg/utils"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	db, err := database.NewPostgresDB(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	fmt.Println("Starting database seeding...")

	// Seed users
	if err := seedUsers(db); err != nil {
		log.Fatalf("Failed to seed users: %v", err)
	}

	fmt.Println("✓ Database seeding completed successfully!")
}

func seedUsers(db *database.Database) error {
	fmt.Println("Seeding users...")

	// Create admin user
	adminPassword, _ := utils.HashPassword("Admin123!")
	admin := &model.User{
		Email:     "admin@example.com",
		Username:  "admin",
		Password:  adminPassword,
		FirstName: "Admin",
		LastName:  "User",
		Role:      "admin",
		IsActive:  true,
	}

	if err := db.FirstOrCreate(admin, model.User{Email: admin.Email}).Error; err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}
	fmt.Println("  ✓ Admin user created/exists")

	// Create test users
	testUsers := []struct {
		email     string
		username  string
		firstName string
		lastName  string
		role      string
	}{
		{"user1@example.com", "user1", "John", "Doe", "user"},
		{"user2@example.com", "user2", "Jane", "Smith", "user"},
		{"moderator@example.com", "moderator", "Mod", "User", "moderator"},
	}

	password, _ := utils.HashPassword("Password123!")
	for _, u := range testUsers {
		user := &model.User{
			Email:     u.email,
			Username:  u.username,
			Password:  password,
			FirstName: u.firstName,
			LastName:  u.lastName,
			Role:      u.role,
			IsActive:  true,
		}

		if err := db.FirstOrCreate(user, model.User{Email: user.Email}).Error; err != nil {
			return fmt.Errorf("failed to create user %s: %w", u.email, err)
		}
		fmt.Printf("  ✓ User %s created/exists\n", u.email)
	}

	return nil
}

func init() {
	// Set environment for local development if not set
	if os.Getenv("DB_HOST") == "" {
		os.Setenv("DB_HOST", "localhost")
	}
}
