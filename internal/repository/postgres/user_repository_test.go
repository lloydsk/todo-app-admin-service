package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/todo-app/services/admin-service/internal/config"
	"github.com/todo-app/services/admin-service/internal/model/domain"
	"github.com/todo-app/services/admin-service/internal/repository"
	"github.com/todo-app/services/admin-service/pkg/db"
)

func setupTestDB(t *testing.T) *db.Connection {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Name:     "todo_app",
			User:     "postgres",
			Password: "postgres",
			SSLMode:  "disable",
		},
	}

	dbConn, err := db.NewConnection(cfg.Database)
	if err != nil {
		t.Skipf("Database not available for integration tests: %v", err)
	}

	ctx := context.Background()
	if err := dbConn.HealthCheck(ctx); err != nil {
		t.Skipf("Database health check failed: %v", err)
	}

	dbConn.SetServiceContext(ctx, "integration-test")

	// Check if required tables exist
	var tableExists bool
	err = dbConn.DB.QueryRowContext(ctx, "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'users')").Scan(&tableExists)
	if err != nil || !tableExists {
		t.Skipf("Required database tables not found - skipping integration test. Error: %v, table exists: %v", err, tableExists)
	}

	return dbConn
}

func TestUserRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	dbConn := setupTestDB(t)
	defer dbConn.Close()

	ctx := context.Background()
	userRepo := NewUserRepository(dbConn.DB)

	t.Run("CreateUser", func(t *testing.T) {
		testUser := &domain.User{
			Name:  "Integration Test User",
			Email: fmt.Sprintf("integration-test-%d@example.com", time.Now().Unix()),
			Role:  domain.UserRoleUser,
		}

		err := userRepo.Create(ctx, testUser)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		if testUser.ID == "" {
			t.Error("User ID should be set after creation")
		}
		if testUser.CreatedAt.IsZero() {
			t.Error("CreatedAt should be set after creation")
		}
		if testUser.Version == 0 {
			t.Error("Version should be set after creation")
		}
	})

	t.Run("GetUserByID", func(t *testing.T) {
		// Create a test user first
		testUser := &domain.User{
			Name:  "Get Test User",
			Email: fmt.Sprintf("get-test-%d@example.com", time.Now().Unix()),
			Role:  domain.UserRoleUser,
		}

		err := userRepo.Create(ctx, testUser)
		if err != nil {
			t.Fatalf("Failed to create test user: %v", err)
		}

		retrievedUser, err := userRepo.GetByID(ctx, testUser.ID)
		if err != nil {
			t.Fatalf("Failed to get user by ID: %v", err)
		}

		if retrievedUser.ID != testUser.ID {
			t.Errorf("ID mismatch: got %s, want %s", retrievedUser.ID, testUser.ID)
		}
		if retrievedUser.Email != testUser.Email {
			t.Errorf("Email mismatch: got %s, want %s", retrievedUser.Email, testUser.Email)
		}
	})

	t.Run("GetUserByEmail", func(t *testing.T) {
		testEmail := fmt.Sprintf("email-test-%d@example.com", time.Now().Unix())
		testUser := &domain.User{
			Name:  "Email Test User",
			Email: testEmail,
			Role:  domain.UserRoleUser,
		}

		err := userRepo.Create(ctx, testUser)
		if err != nil {
			t.Fatalf("Failed to create test user: %v", err)
		}

		retrievedUser, err := userRepo.GetByEmail(ctx, testEmail)
		if err != nil {
			t.Fatalf("Failed to get user by email: %v", err)
		}

		if retrievedUser.Email != testEmail {
			t.Errorf("Email mismatch: got %s, want %s", retrievedUser.Email, testEmail)
		}
	})

	t.Run("UpdateUser", func(t *testing.T) {
		testUser := &domain.User{
			Name:  "Update Test User",
			Email: fmt.Sprintf("update-test-%d@example.com", time.Now().Unix()),
			Role:  domain.UserRoleUser,
		}

		err := userRepo.Create(ctx, testUser)
		if err != nil {
			t.Fatalf("Failed to create test user: %v", err)
		}

		originalVersion := testUser.Version
		testUser.Name = "Updated Name"

		err = userRepo.Update(ctx, testUser)
		if err != nil {
			t.Fatalf("Failed to update user: %v", err)
		}

		if testUser.Version <= originalVersion {
			t.Error("Version should be incremented after update")
		}

		retrievedUser, err := userRepo.GetByID(ctx, testUser.ID)
		if err != nil {
			t.Fatalf("Failed to get updated user: %v", err)
		}

		if retrievedUser.Name != "Updated Name" {
			t.Errorf("Name not updated: got %s, want Updated Name", retrievedUser.Name)
		}
	})

	t.Run("ListUsers", func(t *testing.T) {
		users, total, err := userRepo.List(ctx, repository.ListOptions{
			Page:     1,
			PageSize: 10,
		})
		if err != nil {
			t.Fatalf("Failed to list users: %v", err)
		}

		if len(users) == 0 {
			t.Error("Expected at least some users")
		}
		if total < int64(len(users)) {
			t.Errorf("Total count (%d) should be >= returned users (%d)", total, len(users))
		}
	})

	t.Run("SoftDeleteUser", func(t *testing.T) {
		testUser := &domain.User{
			Name:  "Delete Test User",
			Email: fmt.Sprintf("delete-test-%d@example.com", time.Now().Unix()),
			Role:  domain.UserRoleUser,
		}

		err := userRepo.Create(ctx, testUser)
		if err != nil {
			t.Fatalf("Failed to create test user: %v", err)
		}

		err = userRepo.SoftDelete(ctx, testUser.ID, testUser.Version)
		if err != nil {
			t.Fatalf("Failed to soft delete user: %v", err)
		}

		// Verify user is soft deleted (GetByID should return not found error)
		_, err = userRepo.GetByID(ctx, testUser.ID)
		if err == nil {
			t.Error("Expected error when getting soft deleted user")
		}
		if !domain.IsNotFoundError(err) {
			t.Errorf("Expected not found error, got: %v", err)
		}
	})
}
