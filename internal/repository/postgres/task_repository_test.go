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

func setupTaskTestDB(t *testing.T) (*db.Connection, string) {
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

	dbConn.SetServiceContext(ctx, "task-integration-test")

	// Create a test user for task assignments
	userRepo := NewUserRepository(dbConn.DB)
	testUser := &domain.User{
		Name:  "Task Test User",
		Email: fmt.Sprintf("task-test-%d@example.com", time.Now().Unix()),
		Role:  domain.UserRoleUser,
	}

	err = userRepo.Create(ctx, testUser)
	if err != nil {
		t.Fatalf("Failed to create test user for tasks: %v", err)
	}

	return dbConn, testUser.ID
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func TestTaskRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	dbConn, assigneeID := setupTaskTestDB(t)
	defer dbConn.Close()

	ctx := context.Background()
	taskRepo := NewTaskRepository(dbConn.DB)

	t.Run("CreateTask", func(t *testing.T) {
		testTask := &domain.Task{
			Title:       "Integration Test Task",
			Description: "This task is for testing the repository layer",
			AssigneeID:  assigneeID,
			Status:      domain.TaskStatusOpen,
			Priority:    domain.TaskPriorityMedium,
			DueDate:     timePtr(time.Now().Add(24 * time.Hour)),
		}

		err := taskRepo.Create(ctx, testTask)
		if err != nil {
			t.Fatalf("Failed to create task: %v", err)
		}

		if testTask.ID == "" {
			t.Error("Task ID should be set after creation")
		}
		if testTask.CreatedAt.IsZero() {
			t.Error("CreatedAt should be set after creation")
		}
		if testTask.Version == 0 {
			t.Error("Version should be set after creation")
		}
	})

	t.Run("GetTaskByID", func(t *testing.T) {
		testTask := &domain.Task{
			Title:       "Get Test Task",
			Description: "Task for testing retrieval",
			AssigneeID:  assigneeID,
			Status:      domain.TaskStatusOpen,
			Priority:    domain.TaskPriorityHigh,
		}

		err := taskRepo.Create(ctx, testTask)
		if err != nil {
			t.Fatalf("Failed to create test task: %v", err)
		}

		retrievedTask, err := taskRepo.GetByID(ctx, testTask.ID)
		if err != nil {
			t.Fatalf("Failed to get task by ID: %v", err)
		}

		if retrievedTask.ID != testTask.ID {
			t.Errorf("ID mismatch: got %s, want %s", retrievedTask.ID, testTask.ID)
		}
		if retrievedTask.Title != testTask.Title {
			t.Errorf("Title mismatch: got %s, want %s", retrievedTask.Title, testTask.Title)
		}
		if retrievedTask.Status != testTask.Status {
			t.Errorf("Status mismatch: got %v, want %v", retrievedTask.Status, testTask.Status)
		}
	})

	t.Run("UpdateTask", func(t *testing.T) {
		testTask := &domain.Task{
			Title:       "Update Test Task",
			Description: "Task for testing updates",
			AssigneeID:  assigneeID,
			Status:      domain.TaskStatusOpen,
			Priority:    domain.TaskPriorityLow,
		}

		err := taskRepo.Create(ctx, testTask)
		if err != nil {
			t.Fatalf("Failed to create test task: %v", err)
		}

		originalVersion := testTask.Version
		testTask.Title = "Updated Task Title"
		testTask.Status = domain.TaskStatusInProgress

		err = taskRepo.Update(ctx, testTask)
		if err != nil {
			t.Fatalf("Failed to update task: %v", err)
		}

		if testTask.Version <= originalVersion {
			t.Error("Version should be incremented after update")
		}

		retrievedTask, err := taskRepo.GetByID(ctx, testTask.ID)
		if err != nil {
			t.Fatalf("Failed to get updated task: %v", err)
		}

		if retrievedTask.Title != "Updated Task Title" {
			t.Errorf("Title not updated: got %s, want Updated Task Title", retrievedTask.Title)
		}
		if retrievedTask.Status != domain.TaskStatusInProgress {
			t.Errorf("Status not updated: got %v, want %v", retrievedTask.Status, domain.TaskStatusInProgress)
		}
	})

	t.Run("ListTasks", func(t *testing.T) {
		tasks, total, err := taskRepo.List(ctx, repository.TaskListOptions{
			ListOptions: repository.ListOptions{
				Page:     1,
				PageSize: 10,
			},
		})
		if err != nil {
			t.Fatalf("Failed to list tasks: %v", err)
		}

		if total < int64(len(tasks)) {
			t.Errorf("Total count (%d) should be >= returned tasks (%d)", total, len(tasks))
		}
	})

	t.Run("SoftDeleteTask", func(t *testing.T) {
		testTask := &domain.Task{
			Title:       "Delete Test Task",
			Description: "Task for testing deletion",
			AssigneeID:  assigneeID,
			Status:      domain.TaskStatusOpen,
			Priority:    domain.TaskPriorityMedium,
		}

		err := taskRepo.Create(ctx, testTask)
		if err != nil {
			t.Fatalf("Failed to create test task: %v", err)
		}

		err = taskRepo.SoftDelete(ctx, testTask.ID, testTask.Version)
		if err != nil {
			t.Fatalf("Failed to soft delete task: %v", err)
		}

		// Verify task is soft deleted
		_, err = taskRepo.GetByID(ctx, testTask.ID)
		if err == nil {
			t.Error("Expected error when getting soft deleted task")
		}
		if !domain.IsNotFoundError(err) {
			t.Errorf("Expected not found error, got: %v", err)
		}
	})
}
