package testutil

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/todo-app/services/admin-service/internal/model/domain"
)

// TestUser creates a test user with default values
func TestUser() *domain.User {
	return &domain.User{
		ID:    uuid.New().String(),
		Name:  "Test User",
		Email: fmt.Sprintf("test-%d@example.com", time.Now().UnixNano()),
		Role:  domain.UserRoleUser,
	}
}

// TestAdminUser creates a test admin user
func TestAdminUser() *domain.User {
	user := TestUser()
	user.Name = "Test Admin"
	user.Email = fmt.Sprintf("admin-%d@example.com", time.Now().UnixNano())
	user.Role = domain.UserRoleAdmin
	return user
}

// TestTask creates a test task with default values
func TestTask(assigneeID string) *domain.Task {
	dueDate := time.Now().Add(24 * time.Hour)
	return &domain.Task{
		ID:          uuid.New().String(),
		Title:       fmt.Sprintf("Test Task %d", time.Now().UnixNano()),
		Description: "This is a test task for unit testing",
		AssigneeID:  assigneeID,
		Status:      domain.TaskStatusOpen,
		Priority:    domain.TaskPriorityMedium,
		DueDate:     &dueDate,
	}
}


// TestCategory creates a test category
func TestCategory() *domain.Category {
	return &domain.Category{
		ID:          uuid.New().String(),
		Name:        fmt.Sprintf("Test Category %d", time.Now().UnixNano()),
		Description: "Test category for unit testing",
		Color:       "#FF5733",
		CreatorID:   uuid.New().String(),
		IsPublic:    true,
	}
}

// TestTag creates a test tag
func TestTag() *domain.Tag {
	return &domain.Tag{
		ID:        uuid.New().String(),
		Name:      fmt.Sprintf("test-tag-%d", time.Now().UnixNano()),
		Color:     "#33FF57",
		CreatorID: "test-creator-id",
	}
}

